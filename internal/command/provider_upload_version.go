package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

type providerManifestType struct {
	Metadata struct {
		ProtocolVersions []string `json:"protocol_versions"`
	} `json:"metadata"`
}

type providerVersionMetadataType struct {
	Version string `json:"version"`
}

type artifactType struct {
	Name            string `json:"name"`
	Path            string `json:"path"`
	OperatingSystem string `json:"goos"`
	Architecture    string `json:"goarch"`
	Type            string `json:"type"`
}

// providerUploadVersionCommand is the top-level structure for the provider upload-version command.
type providerUploadVersionCommand struct {
	meta *Metadata
}

// NewProviderUploadVersionCommandFactory returns a providerUploadVersionCommand struct.
func NewProviderUploadVersionCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return providerUploadVersionCommand{
			meta: meta,
		}, nil
	}
}

func (p providerUploadVersionCommand) Run(args []string) int {
	p.meta.Logger.Debugf("Starting the 'provider upload-version' command with %d arguments:", len(args))
	for ix, arg := range args {
		p.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := p.meta.ReadSettings()
	if err != nil {
		p.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		p.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return p.doProviderUploadVersion(ctx, client, args)
}

func (p providerUploadVersionCommand) doProviderUploadVersion(ctx context.Context, client *tharsis.Client, opts []string) int {
	p.meta.Logger.Debugf("will do provider upload-version, %d opts", len(opts))

	defs := buildProviderUploadVersionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(p.meta.BinaryName+" provider upload-version", defs, opts)
	if err != nil {
		p.meta.Logger.Error(output.FormatError("failed to parse provider upload-version options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		p.meta.Logger.Error(output.FormatError("missing provider upload-version <provider-path>", nil), p.HelpProviderUploadVersion())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive provider upload-version arguments: %s", cmdArgs)
		p.meta.Logger.Error(output.FormatError(msg, nil), p.HelpProviderUploadVersion())
		return 1
	}

	providerPath := cmdArgs[0]
	directoryPath := getOption("directory-path", "", cmdOpts)[0]

	// Error is already logged.
	if !isResourcePathValid(p.meta, providerPath) {
		return 1
	}

	// Make sure the directory path exists and is a directory--to give more precise messages.
	if err = p.checkDirPath(directoryPath); err != nil {
		p.meta.Logger.Error(output.FormatError("invalid directory path", err))
		return 1
	}

	p.meta.UI.Output(fmt.Sprintf("• starting provider version upload: provider=%s directory=%s", providerPath, directoryPath))

	providerManifest, err := p.readProviderManifest(ctx, directoryPath)
	if err != nil {
		p.meta.UI.Error(output.FormatError("failed to read provider manifest", err))
		return 1
	}

	providerVersionMetadata, err := p.readProviderVersionMetdata(ctx, directoryPath)
	if err != nil {
		p.meta.UI.Error(output.FormatError("failed to read provider version metadata", err))
		return 1
	}

	artifacts, err := p.readProviderVersionArtifacts(ctx, directoryPath)
	if err != nil {
		p.meta.UI.Error(output.FormatError("failed to read provider version artifacts", err))
		return 1
	}

	p.meta.UI.Output(fmt.Sprintf("• creating provider version %s", providerVersionMetadata.Version))

	// Create provider version
	providerVersion, err := client.TerraformProviderVersion.CreateProviderVersion(ctx, &types.CreateTerraformProviderVersionInput{
		ProviderPath: providerPath,
		Version:      providerVersionMetadata.Version,
		Protocols:    providerManifest.Metadata.ProtocolVersions,
	})
	if err != nil {
		p.meta.UI.Error(output.FormatError("failed to create provider version", err))
		return 1
	}

	if err = p.uploadReadme(ctx, client, directoryPath, providerVersion); err != nil {
		p.meta.UI.Error(output.FormatError("failed to upload readme file", err))
		return 1
	}

	var checksumMap map[string]string

	// Find Checksum file
	for _, artifact := range artifacts {
		artifactCopy := artifact
		if artifact.Type == "Checksum" {
			checksumMap, err = p.uploadChecksums(ctx, client, directoryPath, providerVersion, &artifactCopy)
			if err != nil {
				p.meta.UI.Error(output.FormatError("failed to upload provider checksums", err))
				return 1
			}
		}
		if artifact.Type == "Signature" {
			if err := p.uploadSignature(ctx, client, directoryPath, providerVersion, &artifactCopy); err != nil {
				p.meta.UI.Error(output.FormatError("failed to upload provider checksums signature", err))
				return 1
			}
		}
	}

	// Create & Upload platforms
	for _, artifact := range artifacts {
		if artifact.Type == "Archive" {
			artifactCopy := artifact
			if err := p.uploadPlatformArchive(ctx, client, directoryPath, providerVersion, &artifactCopy, checksumMap); err != nil {
				p.meta.UI.Error(output.FormatError("failed to upload archive", err))
				return 1
			}
		}
	}

	p.meta.UI.Output("• provider version upload succeeded")

	return 0
}

func (p providerUploadVersionCommand) uploadReadme(
	ctx context.Context,
	client *tharsis.Client,
	dir string,
	providerVersion *types.TerraformProviderVersion,
) error {
	p.meta.UI.Output("• checking for README file")

	matches, err := filepath.Glob(filepath.Join(dir, "README*"))
	if err != nil {
		return fmt.Errorf("error occurred while checking for README: %v", err)
	}

	if len(matches) == 0 {
		p.meta.UI.Output("• skipping readme upload")
		return nil
	}

	p.meta.UI.Output(fmt.Sprintf("• uploading README file %s", matches[0]))

	reader, err := os.Open(matches[0])
	if err != nil {
		return fmt.Errorf("failed to read README file: %v", err)
	}
	defer reader.Close()

	// upload readme file
	if err := client.TerraformProviderVersion.UploadProviderReadme(ctx, providerVersion.Metadata.ID, reader); err != nil {
		return fmt.Errorf("failed to upload readme file: %v", err)
	}

	p.meta.UI.Output("• completed readme upload")

	return nil
}

func (p providerUploadVersionCommand) uploadChecksums(
	ctx context.Context,
	client *tharsis.Client,
	dir string,
	providerVersion *types.TerraformProviderVersion,
	artifact *artifactType,
) (map[string]string, error) {
	checksumMap := map[string]string{}

	p.meta.UI.Output(fmt.Sprintf("• uploading checksums: file=%s", artifact.Path))

	data, err := os.ReadFile(filepath.Join(dir, artifact.Path))
	if err != nil {
		return nil, fmt.Errorf("failed to find %s file: %v", artifact.Path, err)
	}

	checksums := strings.Split(string(data), "\n")
	for _, checksum := range checksums {
		parsedChecksum := strings.Split(checksum, "  ")
		if len(parsedChecksum) == 2 {
			checksumMap[parsedChecksum[1]] = parsedChecksum[0]
		}
	}

	reader, err := os.Open(filepath.Join(dir, artifact.Path))
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %v", artifact.Path, err)
	}
	defer reader.Close()

	// upload sums file
	if err := client.TerraformProviderVersion.UploadProviderChecksums(ctx, providerVersion.Metadata.ID, reader); err != nil {
		return nil, fmt.Errorf("failed to upload checksum file: %v", err)
	}

	p.meta.UI.Output("• completed checksums upload")

	return checksumMap, nil
}

func (p providerUploadVersionCommand) uploadSignature(
	ctx context.Context,
	client *tharsis.Client,
	dir string,
	providerVersion *types.TerraformProviderVersion,
	artifact *artifactType,
) error {
	p.meta.UI.Output(fmt.Sprintf("• uploading checksums signature: file=%s", artifact.Path))

	// upload signature
	reader, err := os.Open(filepath.Join(dir, artifact.Path))
	if err != nil {
		return fmt.Errorf("failed to read %s file: %v", artifact.Path, err)
	}
	defer reader.Close()

	// upload sums file
	if err := client.TerraformProviderVersion.UploadProviderChecksumSignature(ctx, providerVersion.Metadata.ID, reader); err != nil {
		return fmt.Errorf("failed to upload checksum signature file: %v", err)
	}

	p.meta.UI.Output("• completed checksums signature upload")

	return nil
}

func (p providerUploadVersionCommand) uploadPlatformArchive(
	ctx context.Context,
	client *tharsis.Client,
	dir string,
	providerVersion *types.TerraformProviderVersion,
	artifact *artifactType,
	checksumMap map[string]string,
) error {
	checksum, ok := checksumMap[artifact.Name]
	if !ok {
		return fmt.Errorf("failed to find checksum for file %s", artifact.Path)
	}

	p.meta.UI.Output(fmt.Sprintf("• uploading platform %s_%s", artifact.OperatingSystem, artifact.Architecture))

	platform, err := client.TerraformProviderPlatform.CreateProviderPlatform(ctx, &types.CreateTerraformProviderPlatformInput{
		ProviderVersionID: providerVersion.Metadata.ID,
		OperatingSystem:   artifact.OperatingSystem,
		Architecture:      artifact.Architecture,
		SHASum:            checksum,
		Filename:          artifact.Name,
	})
	if err != nil {
		return fmt.Errorf("failed to create platform: %v", err)
	}

	// upload binary
	reader, err := os.Open(filepath.Join(dir, artifact.Path))
	if err != nil {
		return fmt.Errorf("failed to read %s file: %v", artifact.Path, err)
	}
	defer reader.Close()

	// upload binary file
	if err := client.TerraformProviderPlatform.UploadProviderPlatformBinary(ctx, platform.Metadata.ID, reader); err != nil {
		return fmt.Errorf("failed to upload platform binary file: %v", err)
	}

	p.meta.UI.Output(fmt.Sprintf("• completed upload for platform %s_%s", artifact.OperatingSystem, artifact.Architecture))

	return nil
}

func (p providerUploadVersionCommand) readProviderManifest(ctx context.Context, dir string) (*providerManifestType, error) {
	p.meta.UI.Output("• reading terraform-registry-manifest.json")

	// Locate `terraform-registry-manifest.json` file which includes the protocol versions
	data, err := os.ReadFile(filepath.Join(dir, "terraform-registry-manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to find terraform-registry-manifest.json file: %v", err)
	}

	var providerManifest providerManifestType

	if err = json.Unmarshal(data, &providerManifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal terraform-registry-manifest.json file: %v", err)
	}

	return &providerManifest, nil
}

func (p providerUploadVersionCommand) readProviderVersionMetdata(ctx context.Context, dir string) (*providerVersionMetadataType, error) {
	p.meta.UI.Output("• reading metadata.json")

	// Load version string from the metadata.json file in the dist directory
	data, err := os.ReadFile(filepath.Join(dir, "dist", "metadata.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to find metadata.json file: %v", err)
	}

	var providerVersionMetadata providerVersionMetadataType
	if err = json.Unmarshal(data, &providerVersionMetadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata.json file: %v", err)
	}

	return &providerVersionMetadata, nil
}

func (p providerUploadVersionCommand) readProviderVersionArtifacts(ctx context.Context, dir string) ([]artifactType, error) {
	p.meta.UI.Output("• reading artifacts.json")

	// Load the artifacts.json file to get the list of files/platforms that will be uploaded
	data, err := os.ReadFile(filepath.Join(dir, "dist", "artifacts.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to find artifacts.json file: %v", err)
	}

	var artifacts []artifactType
	if err = json.Unmarshal(data, &artifacts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal artifacts.json file: %v", err)
	}

	return artifacts, nil
}

func (p providerUploadVersionCommand) checkDirPath(directoryPath string) error {
	dirStat, err := os.Stat(directoryPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory path does not exist: %s", directoryPath)
	}
	if err != nil {
		return fmt.Errorf("failed to stat directory path %s: %s", directoryPath, err)
	}
	if !dirStat.IsDir() {
		return fmt.Errorf("path is not a directory: %s", directoryPath)
	}
	return nil
}

// buildProviderUploadVersionDefs returns defs used by provider upload-version command.
func buildProviderUploadVersionDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"directory-path": {
			Arguments: []string{"Directory_Path"},
			Synopsis:  "The path of the terraform provider's directory.",
		},
	}
}

func (p providerUploadVersionCommand) Synopsis() string {
	return "Upload a new provider version to the provider registry."
}

func (p providerUploadVersionCommand) Help() string {
	return p.HelpProviderUploadVersion()
}

// HelpProviderUploadVersion produces the help string for the 'provider upload-version' command.
func (p providerUploadVersionCommand) HelpProviderUploadVersion() string {
	return fmt.Sprintf(`
Usage: %s [global options] provider upload-version [options] <full_path>

   The provider upload-version command uploads a new
   provider version to the provider registry.

%s

`, p.meta.BinaryName, buildHelpText(buildProviderUploadVersionDefs()))
}

// The End.

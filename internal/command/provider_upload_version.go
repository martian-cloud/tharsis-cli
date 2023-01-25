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

func (puc providerUploadVersionCommand) Run(args []string) int {
	puc.meta.Logger.Debugf("Starting the 'provider upload-version' command with %d arguments:", len(args))
	for ix, arg := range args {
		puc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := puc.meta.ReadSettings()
	if err != nil {
		puc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		puc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return puc.doProviderUploadVersion(ctx, client, args)
}

func (puc providerUploadVersionCommand) doProviderUploadVersion(ctx context.Context, client *tharsis.Client, opts []string) int {
	puc.meta.Logger.Debugf("will do provider upload-version, %d opts", len(opts))

	defs := puc.buildProviderUploadVersionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(puc.meta.BinaryName+" provider upload-version", defs, opts)
	if err != nil {
		puc.meta.Logger.Error(output.FormatError("failed to parse provider upload-version options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		puc.meta.Logger.Error(output.FormatError("missing provider upload-version <provider-path>", nil), puc.HelpProviderUploadVersion())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive provider upload-version arguments: %s", cmdArgs)
		puc.meta.Logger.Error(output.FormatError(msg, nil), puc.HelpProviderUploadVersion())
		return 1
	}

	providerPath := cmdArgs[0]
	directoryPath := getOption("directory-path", "", cmdOpts)[0]

	// Error is already logged.
	if !isResourcePathValid(puc.meta, providerPath) {
		return 1
	}

	// Make sure the directory path exists and is a directory--to give more precise messages.
	if err = puc.checkDirPath(directoryPath); err != nil {
		puc.meta.Logger.Error(output.FormatError("invalid directory path", err))
		return 1
	}

	puc.meta.UI.Output(fmt.Sprintf("• starting provider version upload: provider=%s directory=%s", providerPath, directoryPath))

	providerManifest, err := puc.readProviderManifest(ctx, directoryPath)
	if err != nil {
		puc.meta.UI.Error(output.FormatError("failed to read provider manifest", err))
		return 1
	}

	providerVersionMetadata, err := puc.readProviderVersionMetdata(ctx, directoryPath)
	if err != nil {
		puc.meta.UI.Error(output.FormatError("failed to read provider version metadata", err))
		return 1
	}

	artifacts, err := puc.readProviderVersionArtifacts(ctx, directoryPath)
	if err != nil {
		puc.meta.UI.Error(output.FormatError("failed to read provider version artifacts", err))
		return 1
	}

	puc.meta.UI.Output(fmt.Sprintf("• creating provider version %s", providerVersionMetadata.Version))

	// Create provider version
	providerVersion, err := client.TerraformProviderVersion.CreateProviderVersion(ctx, &types.CreateTerraformProviderVersionInput{
		ProviderPath: providerPath,
		Version:      providerVersionMetadata.Version,
		Protocols:    providerManifest.Metadata.ProtocolVersions,
	})
	if err != nil {
		puc.meta.UI.Error(output.FormatError("failed to create provider version", err))
		return 1
	}

	if err = puc.uploadReadme(ctx, client, directoryPath, providerVersion); err != nil {
		puc.meta.UI.Error(output.FormatError("failed to upload readme file", err))
		return 1
	}

	var checksumMap map[string]string

	// Find Checksum file
	for _, artifact := range artifacts {
		artifactCopy := artifact
		if artifact.Type == "Checksum" {
			checksumMap, err = puc.uploadChecksums(ctx, client, directoryPath, providerVersion, &artifactCopy)
			if err != nil {
				puc.meta.UI.Error(output.FormatError("failed to upload provider checksums", err))
				return 1
			}
		}
		if artifact.Type == "Signature" {
			if err := puc.uploadSignature(ctx, client, directoryPath, providerVersion, &artifactCopy); err != nil {
				puc.meta.UI.Error(output.FormatError("failed to upload provider checksums signature", err))
				return 1
			}
		}
	}

	// Create & Upload platforms
	for _, artifact := range artifacts {
		if artifact.Type == "Archive" {
			artifactCopy := artifact
			if err := puc.uploadPlatformArchive(ctx, client, directoryPath, providerVersion, &artifactCopy, checksumMap); err != nil {
				puc.meta.UI.Error(output.FormatError("failed to upload archive", err))
				return 1
			}
		}
	}

	puc.meta.UI.Output("• provider version upload succeeded")

	return 0
}

func (puc providerUploadVersionCommand) uploadReadme(
	ctx context.Context,
	client *tharsis.Client,
	dir string,
	providerVersion *types.TerraformProviderVersion,
) error {
	puc.meta.UI.Output("• checking for README file")

	matches, err := filepath.Glob(filepath.Join(dir, "README*"))
	if err != nil {
		return fmt.Errorf("error occurred while checking for README: %v", err)
	}

	if len(matches) == 0 {
		puc.meta.UI.Output("• skipping readme upload")
		return nil
	}

	puc.meta.UI.Output(fmt.Sprintf("• uploading README file %s", matches[0]))

	reader, err := os.Open(matches[0])
	if err != nil {
		return fmt.Errorf("failed to read README file: %v", err)
	}
	defer reader.Close()

	// upload readme file
	if err := client.TerraformProviderVersion.UploadProviderReadme(ctx, providerVersion.Metadata.ID, reader); err != nil {
		return fmt.Errorf("failed to upload readme file: %v", err)
	}

	puc.meta.UI.Output("• completed readme upload")

	return nil
}

func (puc providerUploadVersionCommand) uploadChecksums(
	ctx context.Context,
	client *tharsis.Client,
	dir string,
	providerVersion *types.TerraformProviderVersion,
	artifact *artifactType,
) (map[string]string, error) {
	checksumMap := map[string]string{}

	puc.meta.UI.Output(fmt.Sprintf("• uploading checksums: file=%s", artifact.Path))

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

	puc.meta.UI.Output("• completed checksums upload")

	return checksumMap, nil
}

func (puc providerUploadVersionCommand) uploadSignature(
	ctx context.Context,
	client *tharsis.Client,
	dir string,
	providerVersion *types.TerraformProviderVersion,
	artifact *artifactType,
) error {
	puc.meta.UI.Output(fmt.Sprintf("• uploading checksums signature: file=%s", artifact.Path))

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

	puc.meta.UI.Output("• completed checksums signature upload")

	return nil
}

func (puc providerUploadVersionCommand) uploadPlatformArchive(
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

	puc.meta.UI.Output(fmt.Sprintf("• uploading platform %s_%s", artifact.OperatingSystem, artifact.Architecture))

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

	puc.meta.UI.Output(fmt.Sprintf("• completed upload for platform %s_%s", artifact.OperatingSystem, artifact.Architecture))

	return nil
}

func (puc providerUploadVersionCommand) readProviderManifest(ctx context.Context, dir string) (*providerManifestType, error) {
	puc.meta.UI.Output("• reading terraform-registry-manifest.json")

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

func (puc providerUploadVersionCommand) readProviderVersionMetdata(ctx context.Context, dir string) (*providerVersionMetadataType, error) {
	puc.meta.UI.Output("• reading metadata.json")

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

func (puc providerUploadVersionCommand) readProviderVersionArtifacts(ctx context.Context, dir string) ([]artifactType, error) {
	puc.meta.UI.Output("• reading artifacts.json")

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

func (puc providerUploadVersionCommand) checkDirPath(directoryPath string) error {
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
func (puc providerUploadVersionCommand) buildProviderUploadVersionDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"directory-path": {
			Arguments: []string{"Directory_Path"},
			Synopsis:  "The path of the terraform provider's directory.",
		},
	}
}

func (puc providerUploadVersionCommand) Synopsis() string {
	return "Upload a new provider version to the provider registry."
}

func (puc providerUploadVersionCommand) Help() string {
	return puc.HelpProviderUploadVersion()
}

// HelpProviderUploadVersion produces the help string for the 'provider upload-version' command.
func (puc providerUploadVersionCommand) HelpProviderUploadVersion() string {
	return fmt.Sprintf(`
Usage: %s [global options] provider upload-version [options] <full_path>

   The provider upload-version command uploads a new
   provider version to the provider registry.

%s

`, puc.meta.BinaryName, buildHelpText(puc.buildProviderUploadVersionDefs()))
}

// The End.

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

type terraformProviderManifestType struct {
	Metadata struct {
		ProtocolVersions []string `json:"protocol_versions"`
	} `json:"metadata"`
}

type terraformProviderVersionMetadataType struct {
	Version string `json:"version"`
}

type artifactType struct {
	Name            string `json:"name"`
	Path            string `json:"path"`
	OperatingSystem string `json:"goos"`
	Architecture    string `json:"goarch"`
	Type            string `json:"type"`
}

// terraformProviderUploadVersionCommand is the top-level structure for the terraform-provider upload-version command.
type terraformProviderUploadVersionCommand struct {
	meta *Metadata
}

// NewTerraformProviderUploadVersionCommandFactory returns a (Terraform provider Upload-Version) Command struct.
func NewTerraformProviderUploadVersionCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return terraformProviderUploadVersionCommand{
			meta: meta,
		}, nil
	}
}

func (tpuc terraformProviderUploadVersionCommand) Run(args []string) int {
	tpuc.meta.Logger.Debugf("Starting the 'terraform-provider upload-version' command with %d arguments:", len(args))
	for ix, arg := range args {
		tpuc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := tpuc.meta.ReadSettings()
	if err != nil {
		tpuc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		tpuc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return tpuc.doTerraformProviderUploadVersion(ctx, client, args)
}

func (tpuc terraformProviderUploadVersionCommand) doTerraformProviderUploadVersion(ctx context.Context,
	client *tharsis.Client, opts []string) int {
	tpuc.meta.Logger.Debugf("will do terraform-provider upload-version, %d opts", len(opts))

	defs := tpuc.buildTerraformProviderUploadVersionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(tpuc.meta.BinaryName+" terraform-provider upload-version",
		defs, opts)
	if err != nil {
		tpuc.meta.Logger.Error(output.FormatError("failed to parse terraform-provider upload-version options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		tpuc.meta.Logger.Error(output.FormatError("missing terraform-provider upload-version <provider-path>", nil),
			tpuc.HelpTerraformProviderUploadVersion())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive terraform-provider upload-version arguments: %s", cmdArgs)
		tpuc.meta.Logger.Error(output.FormatError(msg, nil), tpuc.HelpTerraformProviderUploadVersion())
		return 1
	}

	tfProviderPath := cmdArgs[0]
	directoryPath := getOption("directory-path", ".", cmdOpts)[0] // default to "." (should work in *x and Windows)

	// Error is already logged.
	if !isResourcePathValid(tpuc.meta, tfProviderPath) {
		return 1
	}

	// Make sure the directory path exists and is a directory--to give more precise messages.
	if err = tpuc.checkDirPath(directoryPath); err != nil {
		tpuc.meta.Logger.Error(output.FormatError("invalid directory path", err))
		return 1
	}

	tpuc.meta.UI.Output(fmt.Sprintf("• starting terraform-provider version upload: provider-full-path=%s directory=%s",
		tfProviderPath, directoryPath))

	tfProviderManifest, err := tpuc.readTerraformProviderManifest(ctx, directoryPath)
	if err != nil {
		tpuc.meta.UI.Error(output.FormatError("failed to read Terraform provider manifest", err))
		return 1
	}

	tfProviderVersionMetadata, err := tpuc.readTerraformProviderVersionMetdata(ctx, directoryPath)
	if err != nil {
		tpuc.meta.UI.Error(output.FormatError("failed to read Terraform provider version metadata", err))
		return 1
	}

	artifacts, err := tpuc.readTerraformProviderVersionArtifacts(ctx, directoryPath)
	if err != nil {
		tpuc.meta.UI.Error(output.FormatError("failed to read Terraform provider version artifacts", err))
		return 1
	}

	tpuc.meta.UI.Output(fmt.Sprintf("• creating Terraform provider version %s", tfProviderVersionMetadata.Version))

	// Create Terraform provider version
	tfProviderVersion, err := client.TerraformProviderVersion.CreateProviderVersion(ctx,
		&types.CreateTerraformProviderVersionInput{
			ProviderPath: tfProviderPath,
			Version:      tfProviderVersionMetadata.Version,
			Protocols:    tfProviderManifest.Metadata.ProtocolVersions,
		})
	if err != nil {
		tpuc.meta.UI.Error(output.FormatError("failed to create Terraform provider version", err))
		return 1
	}

	if err = tpuc.uploadReadme(ctx, client, directoryPath, tfProviderVersion); err != nil {
		tpuc.meta.UI.Error(output.FormatError("failed to upload readme file", err))
		return 1
	}

	var checksumMap map[string]string

	// Find Checksum file
	for _, artifact := range artifacts {
		artifactCopy := artifact
		if artifact.Type == "Checksum" {
			checksumMap, err = tpuc.uploadChecksums(ctx, client, directoryPath, tfProviderVersion, &artifactCopy)
			if err != nil {
				tpuc.meta.UI.Error(output.FormatError("failed to upload Terraform provider checksums", err))
				return 1
			}
		}
		if artifact.Type == "Signature" {
			if err := tpuc.uploadSignature(ctx, client, directoryPath, tfProviderVersion, &artifactCopy); err != nil {
				tpuc.meta.UI.Error(output.FormatError("failed to upload Terraform provider checksums signature", err))
				return 1
			}
		}
	}

	// Create & Upload platforms
	for _, artifact := range artifacts {
		if artifact.Type == "Archive" {
			artifactCopy := artifact
			if err := tpuc.uploadPlatformArchive(ctx, client, directoryPath, tfProviderVersion, &artifactCopy, checksumMap); err != nil {
				tpuc.meta.UI.Error(output.FormatError("failed to upload archive", err))
				return 1
			}
		}
	}

	tpuc.meta.UI.Output("• Terraform provider version upload succeeded")

	return 0
}

func (tpuc terraformProviderUploadVersionCommand) uploadReadme(
	ctx context.Context,
	client *tharsis.Client,
	dir string,
	tfProviderVersion *types.TerraformProviderVersion,
) error {
	tpuc.meta.UI.Output("• checking for README file")

	matches, err := filepath.Glob(filepath.Join(dir, "README*"))
	if err != nil {
		return fmt.Errorf("error occurred while checking for README: %v", err)
	}

	if len(matches) == 0 {
		tpuc.meta.UI.Output("• skipping readme upload")
		return nil
	}

	tpuc.meta.UI.Output(fmt.Sprintf("• uploading README file %s", matches[0]))

	reader, err := os.Open(matches[0])
	if err != nil {
		return fmt.Errorf("failed to read README file: %v", err)
	}
	defer reader.Close()

	// upload readme file
	if err := client.TerraformProviderVersion.UploadProviderReadme(ctx,
		tfProviderVersion.Metadata.ID, reader); err != nil {
		return fmt.Errorf("failed to upload readme file: %v", err)
	}

	tpuc.meta.UI.Output("• completed readme upload")

	return nil
}

func (tpuc terraformProviderUploadVersionCommand) uploadChecksums(
	ctx context.Context,
	client *tharsis.Client,
	dir string,
	tfProviderVersion *types.TerraformProviderVersion,
	artifact *artifactType,
) (map[string]string, error) {
	checksumMap := map[string]string{}

	tpuc.meta.UI.Output(fmt.Sprintf("• uploading checksums: file=%s", artifact.Path))

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
	if err := client.TerraformProviderVersion.UploadProviderChecksums(ctx,
		tfProviderVersion.Metadata.ID, reader); err != nil {
		return nil, fmt.Errorf("failed to upload checksum file: %v", err)
	}

	tpuc.meta.UI.Output("• completed checksums upload")

	return checksumMap, nil
}

func (tpuc terraformProviderUploadVersionCommand) uploadSignature(
	ctx context.Context,
	client *tharsis.Client,
	dir string,
	tfProviderVersion *types.TerraformProviderVersion,
	artifact *artifactType,
) error {
	tpuc.meta.UI.Output(fmt.Sprintf("• uploading checksums signature: file=%s", artifact.Path))

	// upload signature
	reader, err := os.Open(filepath.Join(dir, artifact.Path))
	if err != nil {
		return fmt.Errorf("failed to read %s file: %v", artifact.Path, err)
	}
	defer reader.Close()

	// upload sums file
	if err := client.TerraformProviderVersion.UploadProviderChecksumSignature(ctx,
		tfProviderVersion.Metadata.ID, reader); err != nil {
		return fmt.Errorf("failed to upload checksum signature file: %v", err)
	}

	tpuc.meta.UI.Output("• completed checksums signature upload")

	return nil
}

func (tpuc terraformProviderUploadVersionCommand) uploadPlatformArchive(
	ctx context.Context,
	client *tharsis.Client,
	dir string,
	tfProviderVersion *types.TerraformProviderVersion,
	artifact *artifactType,
	checksumMap map[string]string,
) error {
	checksum, ok := checksumMap[artifact.Name]
	if !ok {
		return fmt.Errorf("failed to find checksum for file %s", artifact.Path)
	}

	tpuc.meta.UI.Output(fmt.Sprintf("• uploading platform %s_%s", artifact.OperatingSystem, artifact.Architecture))

	platform, err := client.TerraformProviderPlatform.CreateProviderPlatform(ctx,
		&types.CreateTerraformProviderPlatformInput{
			ProviderVersionID: tfProviderVersion.Metadata.ID,
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

	tpuc.meta.UI.Output(fmt.Sprintf("• completed upload for platform %s_%s", artifact.OperatingSystem, artifact.Architecture))

	return nil
}

func (tpuc terraformProviderUploadVersionCommand) readTerraformProviderManifest(ctx context.Context,
	dir string) (*terraformProviderManifestType, error) {
	tpuc.meta.UI.Output("• reading terraform-registry-manifest.json")

	// Locate `terraform-registry-manifest.json` file which includes the protocol versions
	data, err := os.ReadFile(filepath.Join(dir, "terraform-registry-manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to find terraform-registry-manifest.json file: %v", err)
	}

	var tfProviderManifest terraformProviderManifestType

	if err = json.Unmarshal(data, &tfProviderManifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal terraform-registry-manifest.json file: %v", err)
	}

	return &tfProviderManifest, nil
}

func (tpuc terraformProviderUploadVersionCommand) readTerraformProviderVersionMetdata(ctx context.Context,
	dir string) (*terraformProviderVersionMetadataType, error) {
	tpuc.meta.UI.Output("• reading metadata.json")

	// Load version string from the metadata.json file in the dist directory
	data, err := os.ReadFile(filepath.Join(dir, "dist", "metadata.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to find metadata.json file: %v", err)
	}

	var tfProviderVersionMetadata terraformProviderVersionMetadataType
	if err = json.Unmarshal(data, &tfProviderVersionMetadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata.json file: %v", err)
	}

	return &tfProviderVersionMetadata, nil
}

func (tpuc terraformProviderUploadVersionCommand) readTerraformProviderVersionArtifacts(ctx context.Context,
	dir string) ([]artifactType, error) {
	tpuc.meta.UI.Output("• reading artifacts.json")

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

func (tpuc terraformProviderUploadVersionCommand) checkDirPath(directoryPath string) error {
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

// buildTerraformProviderUploadVersionDefs returns defs used by terraform-provider upload-version command.
func (tpuc terraformProviderUploadVersionCommand) buildTerraformProviderUploadVersionDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"directory-path": {
			Arguments: []string{"Directory_Path"},
			Synopsis:  "The path of the terraform provider's directory.",
		},
	}
}

func (tpuc terraformProviderUploadVersionCommand) Synopsis() string {
	return "Upload a new Terraform provider version to the provider registry."
}

func (tpuc terraformProviderUploadVersionCommand) Help() string {
	return tpuc.HelpTerraformProviderUploadVersion()
}

// HelpTerraformProviderUploadVersion produces the help string for the 'terraform-provider upload-version' command.
func (tpuc terraformProviderUploadVersionCommand) HelpTerraformProviderUploadVersion() string {
	return fmt.Sprintf(`
Usage: %s [global options] terraform-provider upload-version [options] <full_path>

   The terraform-provider upload-version command uploads a new
   Terraform provider version to the provider registry.

%s

`, tpuc.meta.BinaryName, buildHelpText(tpuc.buildTerraformProviderUploadVersionDefs()))
}

// The End.

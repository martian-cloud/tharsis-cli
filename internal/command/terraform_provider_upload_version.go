package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tfe"
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

type terraformProviderUploadVersionCommand struct {
	*BaseCommand

	sg        terminal.StepGroup
	directory string
}

func (c *terraformProviderUploadVersionCommand) validate() error {
	const message = "provider-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewTerraformProviderUploadVersionCommandFactory returns a new terraformProviderUploadVersionCommand.
func NewTerraformProviderUploadVersionCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderUploadVersionCommand{
			BaseCommand: baseCommand,
			sg:          baseCommand.UI.StepGroup(),
		}, nil
	}
}

func (c *terraformProviderUploadVersionCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider upload-version"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	dirStat, err := os.Stat(c.directory)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to stat directory path")
		return 1
	}

	if !dirStat.IsDir() {
		c.UI.Errorf("path is not a directory: %s", c.directory)
		return 1
	}

	step := c.sg.Add("Get provider")
	provider, err := c.grpcClient.TerraformProvidersClient.GetTerraformProviderByID(c.Context, &pb.GetTerraformProviderByIDRequest{
		Id: c.arguments[0],
	})
	if err != nil {
		step.Abort()
		c.UI.ErrorWithSummary(err, "failed to get provider")
		return 1
	}
	step.Done()

	manifest, err := c.readManifest()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to read manifest")
		return 1
	}

	versionMetadata, err := c.readVersionMetadata()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to read version metadata")
		return 1
	}

	artifacts, err := c.readArtifacts()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to read artifacts")
		return 1
	}

	step = c.sg.Add("Create provider version %q", versionMetadata.Version)
	providerVersion, err := c.grpcClient.TerraformProvidersClient.CreateTerraformProviderVersion(c.Context, &pb.CreateTerraformProviderVersionRequest{
		ProviderId: provider.Metadata.Id,
		Version:    versionMetadata.Version,
		Protocols:  manifest.Metadata.ProtocolVersions,
	})
	if err != nil {
		step.Abort()
		c.UI.ErrorWithSummary(err, "failed to create provider version")
		return 1
	}
	step.Done()

	curSettings, err := c.getCurrentSettings()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get settings")
		return 1
	}

	tokenGetter, err := curSettings.CurrentProfile.NewTokenGetter(c.Context)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get token")
		return 1
	}

	tfeClient, err := tfe.NewRESTClient(curSettings.CurrentProfile.Endpoint, tokenGetter, c.HTTPClient)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create REST client")
		return 1
	}

	if err = c.uploadReadme(tfeClient, providerVersion.Metadata.Id); err != nil {
		c.UI.ErrorWithSummary(err, "failed to upload README")
		return 1
	}

	var checksumMap map[string]string
	for _, artifact := range artifacts {
		if artifact.Type == "Checksum" {
			checksumMap, err = c.uploadChecksums(tfeClient, providerVersion.Metadata.Id, artifact.Path)
			if err != nil {
				c.UI.ErrorWithSummary(err, "failed to upload checksums")
				return 1
			}
		}
		if artifact.Type == "Signature" {
			if err := c.uploadSignature(tfeClient, providerVersion.Metadata.Id, artifact.Path); err != nil {
				c.UI.ErrorWithSummary(err, "failed to upload signature")
				return 1
			}
		}
	}

	for _, artifact := range artifacts {
		if artifact.Type == "Archive" {
			if err := c.uploadPlatformArchive(tfeClient, providerVersion.Metadata.Id, &artifact, checksumMap); err != nil {
				c.UI.ErrorWithSummary(err, "failed to upload platform archive")
				return 1
			}
		}
	}

	c.UI.Successf("\nProvider version uploaded successfully!")
	return 0
}

func (c *terraformProviderUploadVersionCommand) readManifest() (*terraformProviderManifestType, error) {
	step := c.sg.Add("Read terraform-registry-manifest.json")

	data, err := os.ReadFile(filepath.Join(c.directory, "terraform-registry-manifest.json"))
	if err != nil {
		step.Abort()
		return nil, err
	}

	var manifest terraformProviderManifestType
	if err = json.Unmarshal(data, &manifest); err != nil {
		step.Abort()
		return nil, err
	}

	step.Done()
	return &manifest, nil
}

func (c *terraformProviderUploadVersionCommand) readVersionMetadata() (*terraformProviderVersionMetadataType, error) {
	step := c.sg.Add("Read metadata.json")

	data, err := os.ReadFile(filepath.Join(c.directory, "dist", "metadata.json"))
	if err != nil {
		step.Abort()
		return nil, err
	}

	var metadata terraformProviderVersionMetadataType
	if err = json.Unmarshal(data, &metadata); err != nil {
		step.Abort()
		return nil, err
	}

	step.Done()
	return &metadata, nil
}

func (c *terraformProviderUploadVersionCommand) readArtifacts() ([]artifactType, error) {
	step := c.sg.Add("Read artifacts.json")

	data, err := os.ReadFile(filepath.Join(c.directory, "dist", "artifacts.json"))
	if err != nil {
		step.Abort()
		return nil, err
	}

	var artifacts []artifactType
	if err = json.Unmarshal(data, &artifacts); err != nil {
		step.Abort()
		return nil, err
	}

	step.Done()
	return artifacts, nil
}

func (c *terraformProviderUploadVersionCommand) uploadReadme(restClient tfe.RESTClient, providerVersionID string) error {
	step := c.sg.Add("Upload README")

	matches, err := filepath.Glob(filepath.Join(c.directory, "README*"))
	if err != nil {
		step.Abort()
		return err
	}

	if len(matches) == 0 {
		step.Update("README upload not needed")
		step.Done()
		return nil
	}

	if err := restClient.UploadProviderReadme(c.Context, &tfe.UploadProviderReadmeInput{
		ProviderVersionID: providerVersionID,
		ReadmePath:        matches[0],
	}); err != nil {
		step.Abort()
		return err
	}

	step.Done()
	return nil
}

func (c *terraformProviderUploadVersionCommand) uploadChecksums(restClient tfe.RESTClient, providerVersionID, artifactPath string) (map[string]string, error) {
	step := c.sg.Add("Upload checksums")

	checksumMap := map[string]string{}

	data, err := os.ReadFile(filepath.Join(c.directory, artifactPath))
	if err != nil {
		step.Abort()
		return nil, err
	}

	for checksum := range strings.SplitSeq(string(data), "\n") {
		hash, filename, ok := strings.Cut(checksum, "  ")
		if ok {
			checksumMap[filename] = hash
		}
	}

	if err := restClient.UploadProviderChecksums(c.Context, &tfe.UploadProviderChecksumsInput{
		ProviderVersionID: providerVersionID,
		ChecksumsPath:     filepath.Join(c.directory, artifactPath),
	}); err != nil {
		step.Abort()
		return nil, err
	}

	step.Done()
	return checksumMap, nil
}

func (c *terraformProviderUploadVersionCommand) uploadSignature(restClient tfe.RESTClient, providerVersionID, artifactPath string) error {
	step := c.sg.Add("Upload checksums signature")

	if err := restClient.UploadProviderChecksumSignature(c.Context, &tfe.UploadProviderChecksumSignatureInput{
		ProviderVersionID: providerVersionID,
		SignaturePath:     filepath.Join(c.directory, artifactPath),
	}); err != nil {
		step.Abort()
		return err
	}

	step.Done()
	return nil
}

func (c *terraformProviderUploadVersionCommand) uploadPlatformArchive(restClient tfe.RESTClient, providerVersionID string, artifact *artifactType, checksumMap map[string]string) error {
	step := c.sg.Add("Upload platform %s_%s", artifact.OperatingSystem, artifact.Architecture)

	checksum, ok := checksumMap[artifact.Name]
	if !ok {
		step.Abort()
		return fmt.Errorf("failed to find checksum for file %s", artifact.Path)
	}

	platform, err := c.grpcClient.TerraformProvidersClient.CreateTerraformProviderPlatform(c.Context, &pb.CreateTerraformProviderPlatformRequest{
		ProviderVersionId: providerVersionID,
		Os:                artifact.OperatingSystem,
		Arch:              artifact.Architecture,
		ShaSum:            checksum,
		Filename:          artifact.Name,
	})
	if err != nil {
		step.Abort()
		return err
	}

	if err := restClient.UploadProviderPlatformBinary(c.Context, &tfe.UploadProviderPlatformBinaryInput{
		PlatformID: platform.Metadata.Id,
		BinaryPath: filepath.Join(c.directory, artifact.Path),
	}); err != nil {
		step.Abort()
		return err
	}

	step.Done()
	return nil
}

func (*terraformProviderUploadVersionCommand) Synopsis() string {
	return "Upload a new Terraform provider version to the provider registry."
}

func (*terraformProviderUploadVersionCommand) Description() string {
	return `
   The terraform-provider upload-version command uploads a new
   Terraform provider version to the provider registry.
`
}

func (*terraformProviderUploadVersionCommand) Usage() string {
	return "tharsis [global options] terraform-provider upload-version [options] <provider-id>"
}

func (*terraformProviderUploadVersionCommand) Example() string {
	return `
tharsis terraform-provider upload-version \
  --directory ./my-provider \
  trn:terraform_provider:my-group/my-provider
`
}

func (c *terraformProviderUploadVersionCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.directory,
		"directory",
		".",
		"The path of the terraform provider's directory.",
	)
	return f
}

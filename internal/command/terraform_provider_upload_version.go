package command

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tfe"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
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
	directory *string
}

var _ Command = (*terraformProviderUploadVersionCommand)(nil)

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

	dirStat, err := os.Stat(*c.directory)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to stat directory path")
		return 1
	}

	if !dirStat.IsDir() {
		c.UI.Errorf("path is not a directory: %s", *c.directory)
		return 1
	}

	provider, err := c.getProvider()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get provider")
		return 1
	}

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

	providerVersion, err := c.createProviderVersion(provider.Metadata.Id, versionMetadata.Version, manifest.Metadata.ProtocolVersions)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create provider version")
		return 1
	}

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

	c.sg.Wait()
	c.UI.Successf("\nProvider version uploaded successfully!")
	return 0
}

func (c *terraformProviderUploadVersionCommand) getProvider() (provider *pb.TerraformProvider, err error) {
	step := c.sg.Add("Get provider")
	defer func() { c.finalizeStep(step, err) }()

	provider, err = c.grpcClient.TerraformProvidersClient.GetTerraformProviderByID(c.Context, &pb.GetTerraformProviderByIDRequest{
		Id: trn.ToTRN(trn.ResourceTypeTerraformProvider, c.arguments[0]),
	})
	if err != nil {
		return nil, err
	}

	step.Update("Get provider (%s)", provider.Name)

	return provider, nil
}

func (c *terraformProviderUploadVersionCommand) readManifest() (manifest *terraformProviderManifestType, err error) {
	step := c.sg.Add("Read terraform-registry-manifest.json")
	defer func() { c.finalizeStep(step, err) }()

	data, err := os.ReadFile(filepath.Join(*c.directory, "terraform-registry-manifest.json"))
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return manifest, nil
}

func (c *terraformProviderUploadVersionCommand) readVersionMetadata() (metadata *terraformProviderVersionMetadataType, err error) {
	step := c.sg.Add("Read metadata.json")
	defer func() { c.finalizeStep(step, err) }()

	data, err := os.ReadFile(filepath.Join(*c.directory, "dist", "metadata.json"))
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

func (c *terraformProviderUploadVersionCommand) readArtifacts() (artifacts []artifactType, err error) {
	step := c.sg.Add("Read artifacts.json")
	defer func() { c.finalizeStep(step, err) }()

	data, err := os.ReadFile(filepath.Join(*c.directory, "dist", "artifacts.json"))
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &artifacts); err != nil {
		return nil, err
	}

	return artifacts, nil
}

func (c *terraformProviderUploadVersionCommand) createProviderVersion(providerID, version string, protocols []string) (pv *pb.TerraformProviderVersion, err error) {
	step := c.sg.Add("Create provider version %q", version)
	defer func() { c.finalizeStep(step, err) }()

	pv, err = c.grpcClient.TerraformProvidersClient.CreateTerraformProviderVersion(c.Context, &pb.CreateTerraformProviderVersionRequest{
		ProviderId: providerID,
		Version:    version,
		Protocols:  protocols,
	})

	return pv, err
}

func (c *terraformProviderUploadVersionCommand) uploadReadme(restClient tfe.RESTClient, providerVersionID string) (err error) {
	step := c.sg.Add("Upload README")
	defer func() { c.finalizeStep(step, err) }()

	matches, err := filepath.Glob(filepath.Join(*c.directory, "README*"))
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		step.Update("README upload not needed")
		return nil
	}

	return restClient.UploadProviderReadme(c.Context, &tfe.UploadProviderReadmeInput{
		ProviderVersionID: providerVersionID,
		ReadmePath:        matches[0],
	})
}

func (c *terraformProviderUploadVersionCommand) uploadChecksums(restClient tfe.RESTClient, providerVersionID, artifactPath string) (checksumMap map[string]string, err error) {
	step := c.sg.Add("Upload checksums")
	defer func() { c.finalizeStep(step, err) }()

	checksumMap = map[string]string{}

	data, err := os.ReadFile(filepath.Join(*c.directory, artifactPath))
	if err != nil {
		return nil, err
	}

	for checksum := range strings.SplitSeq(string(data), "\n") {
		hash, filename, ok := strings.Cut(checksum, "  ")
		if ok {
			checksumMap[filename] = hash
		}
	}

	if err = restClient.UploadProviderChecksums(c.Context, &tfe.UploadProviderChecksumsInput{
		ProviderVersionID: providerVersionID,
		ChecksumsPath:     filepath.Join(*c.directory, artifactPath),
	}); err != nil {
		return nil, err
	}

	return checksumMap, nil
}

func (c *terraformProviderUploadVersionCommand) uploadSignature(restClient tfe.RESTClient, providerVersionID, artifactPath string) (err error) {
	step := c.sg.Add("Upload checksums signature")
	defer func() { c.finalizeStep(step, err) }()

	return restClient.UploadProviderChecksumSignature(c.Context, &tfe.UploadProviderChecksumSignatureInput{
		ProviderVersionID: providerVersionID,
		SignaturePath:     filepath.Join(*c.directory, artifactPath),
	})
}

func (c *terraformProviderUploadVersionCommand) uploadPlatformArchive(restClient tfe.RESTClient, providerVersionID string, artifact *artifactType, checksumMap map[string]string) (err error) {
	step := c.sg.Add("Upload platform %s_%s", artifact.OperatingSystem, artifact.Architecture)
	defer func() { c.finalizeStep(step, err) }()

	checksum, ok := checksumMap[artifact.Name]
	if !ok {
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
		return err
	}

	return restClient.UploadProviderPlatformBinary(c.Context, &tfe.UploadProviderPlatformBinaryInput{
		PlatformID: platform.Metadata.Id,
		BinaryPath: filepath.Join(*c.directory, artifact.Path),
	})
}

func (c *terraformProviderUploadVersionCommand) finalizeStep(step terminal.Step, err error) {
	if err != nil {
		step.Abort()
		c.sg.Wait()

		return
	}

	step.Done()
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
  -directory ./my-provider \
  trn:terraform_provider:<group_path>/<name>
`
}

func (c *terraformProviderUploadVersionCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.directory,
		"directory",
		"The path of the terraform provider's directory.",
		flag.Default("."),
	)

	return f
}

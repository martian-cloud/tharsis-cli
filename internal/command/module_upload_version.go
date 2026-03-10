package command

import (
	"encoding/hex"
	"flag"
	"os"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/slug"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tfe"
)

const (
	moduleVersionStatusUploaded = "uploaded"
	moduleVersionStatusErrored  = "errored"
)

type moduleUploadVersionCommand struct {
	*BaseCommand

	sg            terminal.StepGroup
	directoryPath string
	version       string
}

// NewModuleUploadVersionCommandFactory returns a moduleUploadVersionCommand struct.
func NewModuleUploadVersionCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleUploadVersionCommand{
			BaseCommand: baseCommand,
			sg:          baseCommand.UI.StepGroup(),
		}, nil
	}
}

func (c *moduleUploadVersionCommand) validate() error {
	const message = "module-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.version, validation.Required),
	)
}

func (c *moduleUploadVersionCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module upload-version"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	dirStat, err := os.Stat(c.directoryPath)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to stat directory path")
		return 1
	}

	if !dirStat.IsDir() {
		c.UI.Errorf("path is not a directory: %s", c.directoryPath)
		return 1
	}

	step := c.sg.Add("Get module")
	module, err := c.grpcClient.TerraformModulesClient.GetTerraformModuleByID(c.Context, &pb.GetTerraformModuleByIDRequest{
		Id: c.arguments[0],
	})
	if err != nil {
		step.Abort()
		c.UI.ErrorWithSummary(err, "failed to get module")
		return 1
	}
	step.Done()

	step = c.sg.Add("Create module package")
	slugFile, err := os.CreateTemp("", "terraform-slug.tgz")
	if err != nil {
		step.Abort()
		c.UI.ErrorWithSummary(err, "failed to create module version")
		return 1
	}
	defer os.Remove(slugFile.Name())

	slug, err := slug.NewSlug(c.directoryPath, slugFile.Name())
	if err != nil {
		step.Abort()
		c.UI.ErrorWithSummary(err, "failed to create module package")
		return 1
	}
	step.Done()

	step = c.sg.Add("Create module version %q", c.version)
	moduleVersion, err := c.grpcClient.TerraformModulesClient.CreateTerraformModuleVersion(c.Context, &pb.CreateTerraformModuleVersionRequest{
		ModuleId: module.Metadata.Id,
		Version:  c.version,
		ShaSum:   hex.EncodeToString(slug.SHASum),
	})
	if err != nil {
		step.Abort()
		c.UI.ErrorWithSummary(err, "failed to create module version")
		return 1
	}
	step.Done()

	step = c.sg.Add("Upload module version")
	uploadStartTime := time.Now()

	curSettings, err := c.getCurrentSettings()
	if err != nil {
		step.Abort()
		c.UI.ErrorWithSummary(err, "failed to get settings")
		return 1
	}

	tokenGetter, err := curSettings.CurrentProfile.NewTokenGetter(c.Context)
	if err != nil {
		step.Abort()
		c.UI.ErrorWithSummary(err, "failed to get token getter")
		return 1
	}

	tfeClient, err := tfe.NewRESTClient(curSettings.CurrentProfile.Endpoint, tokenGetter, c.HTTPClient)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create REST client")
		return 1
	}

	if err = tfeClient.UploadModuleVersion(c.Context, &tfe.UploadModuleVersionInput{
		ModuleVersionID: moduleVersion.Metadata.Id,
		PackagePath:     slugFile.Name(),
	}); err != nil {
		step.Abort()
		c.UI.ErrorWithSummary(err, "failed to upload module version")

		if _, err = c.grpcClient.TerraformModulesClient.DeleteTerraformModuleVersion(c.Context, &pb.DeleteTerraformModuleVersionRequest{
			Id: moduleVersion.Metadata.Id,
		}); err != nil {
			c.UI.ErrorWithSummary(err, "failed to delete module version")
		}

		return 1
	}

	var updatedModuleVersion *pb.TerraformModuleVersion
	for {
		updatedModuleVersion, err = c.grpcClient.TerraformModulesClient.GetTerraformModuleVersionByID(c.Context, &pb.GetTerraformModuleVersionByIDRequest{
			Id: moduleVersion.Metadata.Id,
		})
		if err != nil {
			step.Abort()
			c.UI.ErrorWithSummary(err, "failed to check module version upload status")
			return 1
		}

		if updatedModuleVersion.Status == moduleVersionStatusUploaded || updatedModuleVersion.Status == moduleVersionStatusErrored {
			break
		}

		time.Sleep(2 * time.Second)
	}

	if updatedModuleVersion.Status == moduleVersionStatusErrored {
		step.Abort()
		c.UI.Errorf("module version upload failed: %s", updatedModuleVersion.Error)
		c.UI.Output(updatedModuleVersion.Diagnostics)

		if _, err = c.grpcClient.TerraformModulesClient.DeleteTerraformModuleVersion(c.Context, &pb.DeleteTerraformModuleVersionRequest{
			Id: moduleVersion.Metadata.Id,
		}); err != nil {
			c.UI.ErrorWithSummary(err, "failed to delete module version")
		}

		return 1
	}

	step.Done()

	c.UI.Successf("\nModule version uploaded successfully! (elapsed: %s)", time.Since(uploadStartTime))

	return 0
}

func (*moduleUploadVersionCommand) Synopsis() string {
	return "Upload a new module version to the module registry."
}

func (*moduleUploadVersionCommand) Description() string {
	return `
   The module upload-version command uploads a new
   module version to the module registry.
`
}

func (*moduleUploadVersionCommand) Usage() string {
	return "tharsis [global options] module upload-version [options] <module-id>"
}

func (*moduleUploadVersionCommand) Example() string {
	return `
tharsis module upload-version \
  --version 1.0.0 \
  --directory-path ./my-module \
  trn:terraform_module:my-group/my-module/aws
`
}

func (c *moduleUploadVersionCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.directoryPath,
		"directory-path",
		".",
		"The path of the terraform module's directory.",
	)
	f.StringVar(
		&c.version,
		"version",
		"",
		"The semantic version for the new module version (required).",
	)
	return f
}

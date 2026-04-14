package command

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/slug"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tfe"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

const (
	moduleVersionStatusUploaded = "uploaded"
	moduleVersionStatusErrored  = "errored"
)

type moduleUploadVersionCommand struct {
	*BaseCommand

	sg            terminal.StepGroup
	directoryPath *string
	version       *string
	toJSON        *bool
}

var _ Command = (*moduleUploadVersionCommand)(nil)

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
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: module id")
	}

	return nil
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

	// Suppress step group output when JSON is requested.
	if c.toJSON != nil && *c.toJSON {
		c.sg = terminal.NewNoopUI().StepGroup()
	}

	dirStat, err := os.Stat(*c.directoryPath)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to stat directory path")
		return 1
	}

	if !dirStat.IsDir() {
		c.UI.Errorf("path is not a directory: %s", *c.directoryPath)
		return 1
	}

	module, err := c.getModule()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get module")
		return 1
	}

	slugFile, shaSum, err := c.createModulePackage()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create module package")
		return 1
	}
	defer os.Remove(slugFile)

	moduleVersion, err := c.createModuleVersion(module.Metadata.Id, shaSum)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create module version")
		return 1
	}

	if err = c.uploadModuleVersion(moduleVersion.Metadata.Id, slugFile); err != nil {
		c.UI.ErrorWithSummary(err, "failed to upload module version")

		if _, dErr := c.grpcClient.TerraformModulesClient.DeleteTerraformModuleVersion(c.Context, &pb.DeleteTerraformModuleVersionRequest{
			Id: moduleVersion.Metadata.Id,
		}); dErr != nil {
			c.UI.ErrorWithSummary(dErr, "failed to delete module version")
		}

		return 1
	}

	c.sg.Wait()

	return c.Output(moduleVersion, c.toJSON)
}

func (c *moduleUploadVersionCommand) getModule() (module *pb.TerraformModule, err error) {
	step := c.sg.Add("Get module")
	defer func() { c.finalizeStep(step, err) }()

	module, err = c.grpcClient.TerraformModulesClient.GetTerraformModuleByID(c.Context, &pb.GetTerraformModuleByIDRequest{
		Id: trn.ToTRN(trn.ResourceTypeTerraformModule, c.arguments[0]),
	})
	if err != nil {
		return nil, err
	}

	step.Update("Get module (%s)", module.Name)

	return module, nil
}

func (c *moduleUploadVersionCommand) createModulePackage() (slugPath string, shaSum []byte, err error) {
	step := c.sg.Add("Create module package")
	defer func() { c.finalizeStep(step, err) }()

	slugFile, err := os.CreateTemp("", "terraform-slug.tgz")
	if err != nil {
		return "", nil, err
	}

	s, err := slug.NewSlug(*c.directoryPath, slugFile.Name())
	if err != nil {
		os.Remove(slugFile.Name())
		return "", nil, err
	}

	return slugFile.Name(), s.SHASum, nil
}

func (c *moduleUploadVersionCommand) createModuleVersion(moduleID string, shaSum []byte) (version *pb.TerraformModuleVersion, err error) {
	step := c.sg.Add("Create module version %q", *c.version)
	defer func() { c.finalizeStep(step, err) }()

	version, err = c.grpcClient.TerraformModulesClient.CreateTerraformModuleVersion(c.Context, &pb.CreateTerraformModuleVersionRequest{
		ModuleId: moduleID,
		Version:  *c.version,
		ShaSum:   hex.EncodeToString(shaSum),
	})

	return version, err
}

func (c *moduleUploadVersionCommand) uploadModuleVersion(moduleVersionID, slugPath string) (err error) {
	step := c.sg.Add("Upload module version")
	defer func() { c.finalizeStep(step, err) }()

	uploadStartTime := time.Now()

	curSettings, err := c.getCurrentSettings()
	if err != nil {
		return err
	}

	tokenGetter, err := curSettings.CurrentProfile.NewTokenGetter(c.Context)
	if err != nil {
		return err
	}

	tfeClient, err := tfe.NewRESTClient(curSettings.CurrentProfile.Endpoint, tokenGetter, c.HTTPClient)
	if err != nil {
		return err
	}

	if err = tfeClient.UploadModuleVersion(c.Context, &tfe.UploadModuleVersionInput{
		ModuleVersionID: moduleVersionID,
		PackagePath:     slugPath,
	}); err != nil {
		return err
	}

	// Poll for upload completion.
	for {
		updatedVersion, pErr := c.grpcClient.TerraformModulesClient.GetTerraformModuleVersionByID(c.Context, &pb.GetTerraformModuleVersionByIDRequest{
			Id: moduleVersionID,
		})
		if pErr != nil {
			return pErr
		}

		if updatedVersion.Status == moduleVersionStatusErrored {
			return fmt.Errorf("module version upload failed: %s\n%s", updatedVersion.Error, updatedVersion.Diagnostics)
		}

		if updatedVersion.Status == moduleVersionStatusUploaded {
			break
		}

		time.Sleep(2 * time.Second)
	}

	step.Update("Upload module version (elapsed: %s)", time.Since(uploadStartTime))

	return nil
}

func (c *moduleUploadVersionCommand) finalizeStep(step terminal.Step, err error) {
	if err != nil {
		step.Abort()
		c.sg.Wait()

		return
	}

	step.Done()
}

func (*moduleUploadVersionCommand) Synopsis() string {
	return "Upload a new module version to the module registry."
}

func (*moduleUploadVersionCommand) Description() string {
	return `
   Packages and uploads a new module version to the
   registry. Use -json to output the created module
   version as JSON and suppress progress updates,
   useful for piping the digest to cosign.
`
}

func (*moduleUploadVersionCommand) Usage() string {
	return "tharsis [global options] module upload-version [options] <module-id>"
}

func (*moduleUploadVersionCommand) Example() string {
	return `
tharsis module upload-version \
  -version "1.0.0" \
  -directory-path "./my-module" \
  trn:terraform_module:<group_path>/<module_name>/<system>
`
}

func (c *moduleUploadVersionCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.directoryPath,
		"directory-path",
		"The path of the terraform module's directory.",
		flag.Default("."),
	)
	f.StringVar(
		&c.version,
		"version",
		"The semantic version for the new module version.",
		flag.Required(),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Output the module digest as JSON.",
	)

	return f
}

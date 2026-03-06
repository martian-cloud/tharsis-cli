package command

import (
	"flag"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type moduleGetVersionCommand struct {
	*BaseCommand

	toJSON bool
}

// NewModuleGetVersionCommandFactory returns a moduleGetVersionCommand struct.
func NewModuleGetVersionCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleGetVersionCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleGetVersionCommand) validate() error {
	const message = "version-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *moduleGetVersionCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module get-version"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetTerraformModuleVersionByIDRequest{
		Id: c.arguments[0],
	}

	c.Logger.Debug("module get version input", "input", input)

	version, err := c.client.TerraformModulesClient.GetTerraformModuleVersionByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get module version")
		return 1
	}

	return outputModuleVersion(c.UI, c.toJSON, version)
}

func (*moduleGetVersionCommand) Synopsis() string {
	return "Get a module version by ID or TRN."
}

func (*moduleGetVersionCommand) Description() string {
	return `
   The module get-version command retrieves details about a specific module version.
`
}

func (*moduleGetVersionCommand) Usage() string {
	return "tharsis [global options] module get-version [options] <version-id>"
}

func (*moduleGetVersionCommand) Example() string {
	return `
tharsis module get-version trn:terraform_module_version:ops/installer/aws/1.0.0
`
}

func (c *moduleGetVersionCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)
	return f
}

func outputModuleVersion(ui terminal.UI, toJSON bool, version *pb.TerraformModuleVersion) int {
	if toJSON {
		if err := ui.JSON(version); err != nil {
			ui.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		t := terminal.NewTable("id", "version", "status", "latest", "sha_sum")
		t.Rich([]string{
			version.Metadata.Id,
			version.SemanticVersion,
			version.Status,
			strconv.FormatBool(version.Latest),
			version.ShaSum,
		}, nil)

		ui.Table(t)
	}

	return 0
}

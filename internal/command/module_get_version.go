package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type moduleGetVersionCommand struct {
	*BaseCommand

	version *string
	toJSON  *bool
}

var _ Command = (*moduleGetVersionCommand)(nil)

// NewModuleGetVersionCommandFactory returns a moduleGetVersionCommand struct.
func NewModuleGetVersionCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleGetVersionCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleGetVersionCommand) validate() error {
	const message = "id is required"
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

	moduleVersionID := c.arguments[0]

	// Handle deprecated -version flag and module path arg.
	if c.version != nil {
		moduleVersionID = trn.NewResourceTRN(trn.ResourceTypeTerraformModuleVersion, moduleVersionID, *c.version)
	}

	input := &pb.GetTerraformModuleVersionByIDRequest{
		Id: moduleVersionID,
	}

	version, err := c.grpcClient.TerraformModulesClient.GetTerraformModuleVersionByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get module version")
		return 1
	}

	return c.OutputProto(version, c.toJSON)
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
tharsis module get-version trn:terraform_module_version:<group_path>/<module_name>/<system>/<version>
`
}

func (c *moduleGetVersionCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Output in JSON format.",
		flag.Default(false),
	)
	f.StringVar(
		&c.version,
		"version",
		"A semver compliant version tag to use as a filter.",
		flag.Deprecated("pass version TRN as argument"),
	)

	return f
}

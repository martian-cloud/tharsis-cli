package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type terraformProviderMirrorGetVersionCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*terraformProviderMirrorGetVersionCommand)(nil)

// NewTerraformProviderMirrorGetVersionCommandFactory returns a terraformProviderMirrorGetVersionCommand struct.
func NewTerraformProviderMirrorGetVersionCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderMirrorGetVersionCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderMirrorGetVersionCommand) validate() error {
	const message = "version-mirror-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *terraformProviderMirrorGetVersionCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider-mirror get-version"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetTerraformProviderVersionMirrorByIDRequest{
		Id: c.arguments[0],
	}

	versionMirror, err := c.grpcClient.TerraformProviderMirrorsClient.GetTerraformProviderVersionMirrorByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get terraform provider version mirror")
		return 1
	}

	return c.Output(versionMirror, c.toJSON)
}

func (*terraformProviderMirrorGetVersionCommand) Synopsis() string {
	return "Get a mirrored terraform provider version."
}

func (*terraformProviderMirrorGetVersionCommand) Description() string {
	return `
   The terraform-provider-mirror get-version command retrieves a terraform provider
   version from the provider mirror.
`
}

func (*terraformProviderMirrorGetVersionCommand) Usage() string {
	return "tharsis [global options] terraform-provider-mirror get-version [options] <version-mirror-id>"
}

func (*terraformProviderMirrorGetVersionCommand) Example() string {
	return `
tharsis terraform-provider-mirror get-version <version-mirror-id>
`
}

func (c *terraformProviderMirrorGetVersionCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Output in JSON format.",
		flag.Default(false),
	)

	return f
}

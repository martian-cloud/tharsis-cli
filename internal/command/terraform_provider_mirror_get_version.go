package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

// terraformProviderMirrorGetVersionCommand is the top-level structure for the terraform-provider-mirror get-version command.
type terraformProviderMirrorGetVersionCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*terraformProviderMirrorGetVersionCommand)(nil)

func (c *terraformProviderMirrorGetVersionCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewTerraformProviderMirrorGetVersionCommandFactory returns a terraformProviderMirrorGetVersionCommand struct.
func NewTerraformProviderMirrorGetVersionCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderMirrorGetVersionCommand{
			BaseCommand: baseCommand,
		}, nil
	}
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

	result, err := c.grpcClient.TerraformProviderMirrorsClient.GetTerraformProviderVersionMirrorByID(c.Context, &pb.GetTerraformProviderVersionMirrorByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get terraform provider version mirror")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*terraformProviderMirrorGetVersionCommand) Synopsis() string {
	return "Get a provider version mirror."
}

func (*terraformProviderMirrorGetVersionCommand) Usage() string {
	return "tharsis [global options] terraform-provider-mirror get-version [options] <id>"
}

func (*terraformProviderMirrorGetVersionCommand) Description() string {
	return `
   Retrieves details about a mirrored provider
   version including its semantic version and
   sync status.
`
}

func (*terraformProviderMirrorGetVersionCommand) Example() string {
	return `
tharsis terraform-provider-mirror get-version <version_mirror_id>
`
}

func (c *terraformProviderMirrorGetVersionCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

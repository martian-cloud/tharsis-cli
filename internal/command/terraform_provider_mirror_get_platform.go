package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

// terraformProviderMirrorGetPlatformCommand is the top-level structure for the terraform-provider-mirror get-platform command.
type terraformProviderMirrorGetPlatformCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*terraformProviderMirrorGetPlatformCommand)(nil)

func (c *terraformProviderMirrorGetPlatformCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewTerraformProviderMirrorGetPlatformCommandFactory returns a terraformProviderMirrorGetPlatformCommand struct.
func NewTerraformProviderMirrorGetPlatformCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderMirrorGetPlatformCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderMirrorGetPlatformCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider-mirror get-platform"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	result, err := c.grpcClient.TerraformProviderMirrorsClient.GetTerraformProviderPlatformMirrorByID(c.Context, &pb.GetTerraformProviderPlatformMirrorByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get terraform provider platform mirror")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*terraformProviderMirrorGetPlatformCommand) Synopsis() string {
	return "Get a provider platform mirror."
}

func (*terraformProviderMirrorGetPlatformCommand) Usage() string {
	return "tharsis [global options] terraform-provider-mirror get-platform [options] <id>"
}

func (*terraformProviderMirrorGetPlatformCommand) Description() string {
	return `
   Retrieves details about a mirrored provider
   platform including its OS, architecture, and
   mirror status.
`
}

func (*terraformProviderMirrorGetPlatformCommand) Example() string {
	return `
tharsis terraform-provider-mirror get-platform <platform_mirror_id>
`
}

func (c *terraformProviderMirrorGetPlatformCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

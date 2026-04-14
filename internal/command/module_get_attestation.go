package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

// moduleGetAttestationCommand is the top-level structure for the module get-attestation command.
type moduleGetAttestationCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*moduleGetAttestationCommand)(nil)

func (c *moduleGetAttestationCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewModuleGetAttestationCommandFactory returns a moduleGetAttestationCommand struct.
func NewModuleGetAttestationCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleGetAttestationCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleGetAttestationCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module get-attestation"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	result, err := c.grpcClient.TerraformModulesClient.GetTerraformModuleAttestationByID(c.Context, &pb.GetTerraformModuleAttestationByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get module attestation")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*moduleGetAttestationCommand) Synopsis() string {
	return "Get a module attestation."
}

func (*moduleGetAttestationCommand) Usage() string {
	return "tharsis [global options] module get-attestation [options] <id>"
}

func (*moduleGetAttestationCommand) Description() string {
	return `
   Retrieves details about a module attestation
   including its data, digest, and associated
   module version.
`
}

func (*moduleGetAttestationCommand) Example() string {
	return `
tharsis module get-attestation <attestation_id>
`
}

func (c *moduleGetAttestationCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

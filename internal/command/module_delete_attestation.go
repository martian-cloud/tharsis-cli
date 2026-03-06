package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// moduleDeleteAttestationCommand is the top-level structure for the module delete attestation command.
type moduleDeleteAttestationCommand struct {
	*BaseCommand

	force bool
}

var _ Command = (*moduleDeleteAttestationCommand)(nil)

func (c *moduleDeleteAttestationCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewModuleDeleteAttestationCommandFactory returns a moduleDeleteAttestationCommand struct.
func NewModuleDeleteAttestationCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleDeleteAttestationCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleDeleteAttestationCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module delete-attestation"),
		WithInputValidator(c.validate),
		WithClient(true),
		WithForcePrompt("Are you sure you want to delete this module attestation?"),
	); code != 0 {
		return code
	}

	input := &pb.DeleteTerraformModuleAttestationRequest{
		Id: c.arguments[0],
	}

	c.Logger.Debug("module delete attestation input", "input", input)

	if _, err := c.client.TerraformModulesClient.DeleteTerraformModuleAttestation(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete module attestation")
		return 1
	}

	c.UI.Output("Module attestation deleted successfully!")

	return 0
}

func (*moduleDeleteAttestationCommand) Synopsis() string {
	return "Delete a module attestation."
}

func (*moduleDeleteAttestationCommand) Usage() string {
	return "tharsis [global options] module delete-attestation [options] <id>"
}

func (*moduleDeleteAttestationCommand) Description() string {
	return `
   The module delete-attestation command deletes a module attestation.
`
}

func (*moduleDeleteAttestationCommand) Example() string {
	return `
tharsis module delete-attestation trn:terraform_module_attestation:ops/installer/aws:VE1W
`
}

func (c *moduleDeleteAttestationCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.force,
		"force",
		false,
		"Force delete the module attestation.",
	)

	return f
}

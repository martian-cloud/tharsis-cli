package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// moduleUpdateAttestationCommand is the top-level structure for the module update attestation command.
type moduleUpdateAttestationCommand struct {
	*BaseCommand

	description *string
	toJSON      bool
}

var _ Command = (*moduleUpdateAttestationCommand)(nil)

func (c *moduleUpdateAttestationCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewModuleUpdateAttestationCommandFactory returns a moduleUpdateAttestationCommand struct.
func NewModuleUpdateAttestationCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleUpdateAttestationCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleUpdateAttestationCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module update-attestation"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.UpdateTerraformModuleAttestationRequest{
		Id:          c.arguments[0],
		Description: c.description,
	}

	c.Logger.Debug("module update attestation input", "input", input)

	updatedAttestation, err := c.client.TerraformModulesClient.UpdateTerraformModuleAttestation(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update module attestation")
		return 1
	}

	return outputModuleAttestation(c.UI, c.toJSON, updatedAttestation)
}

func (*moduleUpdateAttestationCommand) Synopsis() string {
	return "Update a module attestation."
}

func (*moduleUpdateAttestationCommand) Usage() string {
	return "tharsis [global options] module update-attestation [options] <id>"
}

func (*moduleUpdateAttestationCommand) Description() string {
	return `
   The module update-attestation command updates an existing module attestation.
`
}

func (*moduleUpdateAttestationCommand) Example() string {
	return `
tharsis module update-attestation \
  --description "Updated description" \
  trn:terraform_module_attestation:ops/installer/aws:VE1W
`
}

func (c *moduleUpdateAttestationCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"description",
		"Description for the attestation.",
		func(s string) error {
			c.description = &s
			return nil
		},
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}

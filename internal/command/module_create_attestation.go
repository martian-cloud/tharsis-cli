package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// moduleCreateAttestationCommand is the top-level structure for the module create attestation command.
type moduleCreateAttestationCommand struct {
	*BaseCommand

	description     string
	attestationData string
	toJSON          bool
}

var _ Command = (*moduleCreateAttestationCommand)(nil)

func (c *moduleCreateAttestationCommand) validate() error {
	const message = "module-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.attestationData, validation.Required),
	)
}

// NewModuleCreateAttestationCommandFactory returns a moduleCreateAttestationCommand struct.
func NewModuleCreateAttestationCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleCreateAttestationCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleCreateAttestationCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module create-attestation"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.CreateTerraformModuleAttestationRequest{
		ModuleId:        c.arguments[0],
		Description:     c.description,
		AttestationData: c.attestationData,
	}

	c.Logger.Debug("module create attestation input", "input", input)

	createdAttestation, err := c.grpcClient.TerraformModulesClient.CreateTerraformModuleAttestation(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create module attestation")
		return 1
	}

	return outputModuleAttestation(c.UI, c.toJSON, createdAttestation)
}

func (*moduleCreateAttestationCommand) Synopsis() string {
	return "Create a new module attestation."
}

func (*moduleCreateAttestationCommand) Usage() string {
	return "tharsis [global options] module create-attestation [options] <module-id>"
}

func (*moduleCreateAttestationCommand) Description() string {
	return `
   The module create-attestation command creates a new module attestation.
`
}

func (*moduleCreateAttestationCommand) Example() string {
	return `
tharsis module create-attestation \
  --description "Attestation for v1.0.0" \
  --attestation-data '{"key":"value"}' \
  trn:terraform_module:ops/installer/aws
`
}

func (c *moduleCreateAttestationCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.description,
		"description",
		"",
		"Description for the attestation.",
	)
	f.StringVar(
		&c.attestationData,
		"attestation-data",
		"",
		"The attestation data (must be a Base64-encoded string).",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}

func outputModuleAttestation(ui terminal.UI, toJSON bool, attestation *pb.TerraformModuleAttestation) int {
	if toJSON {
		if err := ui.JSON(attestation); err != nil {
			ui.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		t := terminal.NewTable("id", "module_id", "description", "predicate_type", "schema_type")
		t.Rich([]string{
			attestation.Metadata.Id,
			attestation.ModuleId,
			attestation.Description,
			attestation.PredicateType,
			attestation.SchemaType,
		}, nil)

		ui.Table(t)
	}

	return 0
}

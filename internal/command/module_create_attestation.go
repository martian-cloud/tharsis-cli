package command

import (
	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// moduleCreateAttestationCommand is the top-level structure for the module create attestation command.
type moduleCreateAttestationCommand struct {
	*BaseCommand

	description *string
	data        *string
	toJSON      *bool
}

var _ Command = (*moduleCreateAttestationCommand)(nil)

func (c *moduleCreateAttestationCommand) validate() error {
	const message = "module-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
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
		ModuleId:        trn.ToTRN(trn.ResourceTypeTerraformModule, c.arguments[0]),
		Description:     ptr.ToString(c.description),
		AttestationData: *c.data,
	}

	createdAttestation, err := c.grpcClient.TerraformModulesClient.CreateTerraformModuleAttestation(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create module attestation")
		return 1
	}

	return c.Output(createdAttestation, c.toJSON)
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
  -description "Attestation for v1.0.0" \
  -data "aGVsbG8sIHdvcmxk" \
  trn:terraform_module:<module_path>
`
}

func (c *moduleCreateAttestationCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.description,
		"description",
		"Description for the attestation.",
	)
	f.StringVar(
		&c.data,
		"data",
		"The attestation data (must be a Base64-encoded string).",
		flag.Required(),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}

package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
)

type groupSetTerraformVarsCommand struct {
	*BaseCommand

	tfVarFiles []string
}

// NewGroupSetTerraformVarsCommandFactory returns a groupSetTerraformVarsCommand struct.
func NewGroupSetTerraformVarsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupSetTerraformVarsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupSetTerraformVarsCommand) validate() error {
	const message = "group-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.tfVarFiles, validation.Required),
	)
}

func (c *groupSetTerraformVarsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group set-terraform-vars"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: toTRN(trn.ResourceTypeGroup, c.arguments[0])})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group")
		return 1
	}

	parser := varparser.NewVariableParser(nil, false)

	variables, err := parser.ParseTerraformVariables(&varparser.ParseTerraformVariablesInput{TfVarFilePaths: c.tfVarFiles})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to process terraform variables")
		return 1
	}

	pbVariables := make([]*pb.SetNamespaceVariablesInputVariable, len(variables))
	for i, v := range variables {
		pbVariables[i] = &pb.SetNamespaceVariablesInputVariable{
			Key:   v.Key,
			Value: v.Value,
		}
	}

	input := &pb.SetNamespaceVariablesRequest{
		NamespacePath: group.FullPath,
		Category:      pb.VariableCategory_terraform,
		Variables:     pbVariables,
	}

	if _, err = c.grpcClient.NamespaceVariablesClient.SetNamespaceVariables(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to set terraform variables")
		return 1
	}

	c.UI.Successf("Terraform variables set successfully in group!")
	return 0
}

func (*groupSetTerraformVarsCommand) Synopsis() string {
	return "Set terraform variables for a group."
}

func (*groupSetTerraformVarsCommand) Description() string {
	return `
   The group set-terraform-vars command sets terraform variables for a group.
   Command will overwrite any existing Terraform variables in the target group!
   Note: This command does not support sensitive variables.
`
}

func (*groupSetTerraformVarsCommand) Usage() string {
	return "tharsis [global options] group set-terraform-vars [options] <group-id>"
}

func (*groupSetTerraformVarsCommand) Example() string {
	return `
tharsis group set-terraform-vars \
  --tf-var-file terraform.tfvars \
  trn:group:<group_path>
`
}

func (c *groupSetTerraformVarsCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"tf-var-file",
		"Path to a .tfvars file (can be specified multiple times).",
		func(s string) error {
			c.tfVarFiles = append(c.tfVarFiles, s)
			return nil
		},
	)

	return f
}

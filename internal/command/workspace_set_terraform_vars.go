package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
)

type workspaceSetTerraformVarsCommand struct {
	*BaseCommand

	tfVarFiles []string
}

// NewWorkspaceSetTerraformVarsCommandFactory returns a workspaceSetTerraformVarsCommand struct.
func NewWorkspaceSetTerraformVarsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceSetTerraformVarsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceSetTerraformVarsCommand) validate() error {
	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.tfVarFiles, validation.Required),
	)
}

func (c *workspaceSetTerraformVarsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace set-terraform-vars"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{
		Id: trn.ToTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
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
		NamespacePath: workspace.FullPath,
		Category:      pb.VariableCategory_terraform,
		Variables:     pbVariables,
	}

	if _, err = c.grpcClient.NamespaceVariablesClient.SetNamespaceVariables(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to set terraform variables")
		return 1
	}

	c.UI.Successf("Terraform variables set successfully in workspace!")
	return 0
}

func (*workspaceSetTerraformVarsCommand) Synopsis() string {
	return "Set terraform variables for a workspace."
}

func (*workspaceSetTerraformVarsCommand) Description() string {
	return `
   The workspace set-terraform-vars command sets terraform variables for a workspace.
   Command will overwrite any existing Terraform variables in the target workspace!
   Note: This command does not support sensitive variables.
`
}

func (*workspaceSetTerraformVarsCommand) Usage() string {
	return "tharsis [global options] workspace set-terraform-vars [options] <workspace-id>"
}

func (*workspaceSetTerraformVarsCommand) Example() string {
	return `
tharsis workspace set-terraform-vars \
  --tf-var-file terraform.tfvars \
  trn:workspace:<workspace_path>
`
}

func (c *workspaceSetTerraformVarsCommand) Flags() *flag.FlagSet {
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

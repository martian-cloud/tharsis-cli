package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
)

type workspaceSetTerraformVarsCommand struct {
	*BaseCommand

	tfVarFiles []string
}

var _ Command = (*workspaceSetTerraformVarsCommand)(nil)

// NewWorkspaceSetTerraformVarsCommandFactory returns a workspaceSetTerraformVarsCommand struct.
func NewWorkspaceSetTerraformVarsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceSetTerraformVarsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceSetTerraformVarsCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: workspace id")
	}

	return nil
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
   Replaces all Terraform variables in a workspace from a
   tfvars file. Does not support sensitive variables.
`
}

func (*workspaceSetTerraformVarsCommand) Usage() string {
	return "tharsis [global options] workspace set-terraform-vars [options] <workspace-id>"
}

func (*workspaceSetTerraformVarsCommand) Example() string {
	return `
tharsis workspace set-terraform-vars \
  -tf-var-file "terraform.tfvars" \
  trn:workspace:<workspace_path>
`
}

func (c *workspaceSetTerraformVarsCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringSliceVar(
		&c.tfVarFiles,
		"tf-var-file",
		"Path to a .tfvars file.",
		flag.Required(),
	)

	return f
}

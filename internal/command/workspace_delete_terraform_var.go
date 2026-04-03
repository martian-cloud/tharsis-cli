package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceDeleteTerraformVarCommand struct {
	*BaseCommand

	key     *string
	version *int64
}

var _ Command = (*workspaceDeleteTerraformVarCommand)(nil)

// NewWorkspaceDeleteTerraformVarCommandFactory returns a workspaceDeleteTerraformVarCommand struct.
func NewWorkspaceDeleteTerraformVarCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceDeleteTerraformVarCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceDeleteTerraformVarCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: workspace id")
	}

	return nil
}

func (c *workspaceDeleteTerraformVarCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace delete-terraform-var"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	// Get workspace to retrieve full path.
	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{
		Id: trn.ToTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
		return 1
	}

	deleteInput := &pb.DeleteNamespaceVariableRequest{
		Id:      trn.NewResourceTRN(trn.ResourceTypeVariable, workspace.FullPath, pb.VariableCategory_terraform.String(), *c.key),
		Version: c.version,
	}

	if _, err = c.grpcClient.NamespaceVariablesClient.DeleteNamespaceVariable(c.Context, deleteInput); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete terraform variable")
		return 1
	}

	c.UI.Successf("Terraform variable deleted successfully!")
	return 0
}

func (*workspaceDeleteTerraformVarCommand) Synopsis() string {
	return "Delete a terraform variable from a workspace."
}

func (*workspaceDeleteTerraformVarCommand) Description() string {
	return `
   Removes a Terraform variable from a workspace.
`
}

func (*workspaceDeleteTerraformVarCommand) Usage() string {
	return "tharsis [global options] workspace delete-terraform-var [options] <workspace-id>"
}

func (*workspaceDeleteTerraformVarCommand) Example() string {
	return `
tharsis workspace delete-terraform-var \
  -key "region" \
  trn:workspace:<workspace_path>
`
}

func (c *workspaceDeleteTerraformVarCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.key,
		"key",
		"Variable key.",
		flag.Required(),
	)
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)

	return f
}

package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupDeleteTerraformVarCommand struct {
	*BaseCommand

	key     *string
	version *int64
}

var _ Command = (*groupDeleteTerraformVarCommand)(nil)

// NewGroupDeleteTerraformVarCommandFactory returns a groupDeleteTerraformVarCommand struct.
func NewGroupDeleteTerraformVarCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupDeleteTerraformVarCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupDeleteTerraformVarCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: group id")
	}

	return nil
}

func (c *groupDeleteTerraformVarCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group delete-terraform-var"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	// Get group to retrieve full path.
	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: trn.ToTRN(trn.ResourceTypeGroup, c.arguments[0])})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group")
		return 1
	}

	deleteInput := &pb.DeleteNamespaceVariableRequest{
		Id:      trn.NewResourceTRN(trn.ResourceTypeVariable, group.FullPath, pb.VariableCategory_terraform.String(), *c.key),
		Version: c.version,
	}

	if _, err = c.grpcClient.NamespaceVariablesClient.DeleteNamespaceVariable(c.Context, deleteInput); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete terraform variable")
		return 1
	}

	c.UI.Successf("Terraform variable deleted successfully!")
	return 0
}

func (*groupDeleteTerraformVarCommand) Synopsis() string {
	return "Delete a terraform variable from a group."
}

func (*groupDeleteTerraformVarCommand) Description() string {
	return `
   Removes a Terraform variable from a group.
`
}

func (*groupDeleteTerraformVarCommand) Usage() string {
	return "tharsis [global options] group delete-terraform-var [options] <group-id>"
}

func (*groupDeleteTerraformVarCommand) Example() string {
	return `
tharsis group delete-terraform-var \
  -key "region" \
  trn:group:<group_path>
`
}

func (c *groupDeleteTerraformVarCommand) Flags() *flag.Set {
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

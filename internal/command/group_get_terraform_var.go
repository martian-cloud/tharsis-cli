package command

import (
	"errors"

	"github.com/aws/smithy-go/ptr"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupGetTerraformVarCommand struct {
	*BaseCommand

	key           *string
	showSensitive *bool
	toJSON        *bool
}

var _ Command = (*groupGetTerraformVarCommand)(nil)

// NewGroupGetTerraformVarCommandFactory returns a groupGetTerraformVarCommand struct.
func NewGroupGetTerraformVarCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupGetTerraformVarCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupGetTerraformVarCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: group id")
	}

	return nil
}

func (c *groupGetTerraformVarCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group get-terraform-var"),
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

	input := &pb.GetNamespaceVariableByIDRequest{
		Id: trn.NewResourceTRN(trn.ResourceTypeVariable, group.FullPath, pb.VariableCategory_terraform.String(), *c.key),
	}

	variable, err := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariableByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get terraform variable")
		return 1
	}

	// If showing sensitive value, fetch the variable version.
	if *c.showSensitive && variable.Sensitive {
		versionInput := &pb.GetNamespaceVariableVersionByIDRequest{
			Id:                    variable.LatestVersionId,
			IncludeSensitiveValue: true,
		}

		version, err := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariableVersionByID(c.Context, versionInput)
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to get variable version")
			return 1
		}

		// Set the value from the version
		variable.Value = version.Value
	}

	if variable.Sensitive && !*c.showSensitive {
		variable.Value = ptr.String("[SENSITIVE]")
	}

	return c.Output(variable, c.toJSON)
}

func (*groupGetTerraformVarCommand) Synopsis() string {
	return "Get a terraform variable for a group."
}

func (*groupGetTerraformVarCommand) Description() string {
	return `
   Retrieves a Terraform variable from a group.
`
}

func (*groupGetTerraformVarCommand) Usage() string {
	return "tharsis [global options] group get-terraform-var [options] <group-id>"
}

func (*groupGetTerraformVarCommand) Example() string {
	return `
tharsis group get-terraform-var \
  -key "region" \
  trn:group:<group_path>
`
}

func (c *groupGetTerraformVarCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.key,
		"key",
		"Variable key.",
		flag.Required(),
	)
	f.BoolVar(
		&c.showSensitive,
		"show-sensitive",
		"Show the actual value of sensitive variables (requires appropriate permissions).",
		flag.Default(false),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

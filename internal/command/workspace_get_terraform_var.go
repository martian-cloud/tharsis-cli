package command

import (
	"errors"

	"github.com/aws/smithy-go/ptr"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceGetTerraformVarCommand struct {
	*BaseCommand

	key           *string
	showSensitive *bool
	toJSON        *bool
}

var _ Command = (*workspaceGetTerraformVarCommand)(nil)

// NewWorkspaceGetTerraformVarCommandFactory returns a workspaceGetTerraformVarCommand struct.
func NewWorkspaceGetTerraformVarCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceGetTerraformVarCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceGetTerraformVarCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: workspace id")
	}

	return nil
}

func (c *workspaceGetTerraformVarCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace get-terraform-var"),
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

	input := &pb.GetNamespaceVariableByIDRequest{
		Id: trn.NewResourceTRN(trn.ResourceTypeVariable, workspace.FullPath, pb.VariableCategory_terraform.String(), *c.key),
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

		// Set the value from the version.
		variable.Value = version.Value
	}

	if variable.Sensitive && !*c.showSensitive {
		variable.Value = ptr.String("[SENSITIVE]")
	}

	return c.Output(variable, c.toJSON)
}

func (*workspaceGetTerraformVarCommand) Synopsis() string {
	return "Get a terraform variable for a workspace."
}

func (*workspaceGetTerraformVarCommand) Description() string {
	return `
   Retrieves a Terraform variable from a workspace.
`
}

func (*workspaceGetTerraformVarCommand) Usage() string {
	return "tharsis [global options] workspace get-terraform-var [options] <workspace-id>"
}

func (*workspaceGetTerraformVarCommand) Example() string {
	return `
tharsis workspace get-terraform-var \
  -key "region" \
  trn:workspace:<workspace_path>
`
}

func (c *workspaceGetTerraformVarCommand) Flags() *flag.Set {
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

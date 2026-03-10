package command

import (
	"flag"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceDeleteTerraformVarCommand struct {
	*BaseCommand

	key     string
	version *int64
}

// NewWorkspaceDeleteTerraformVarCommandFactory returns a workspaceDeleteTerraformVarCommand struct.
func NewWorkspaceDeleteTerraformVarCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceDeleteTerraformVarCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceDeleteTerraformVarCommand) validate() error {
	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.key, validation.Required),
	)
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

	// Get workspace to retrieve full path
	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
		return 1
	}

	// Build TRN: trn:variable:namespace-path/terraform/key
	variableTRN := trn.NewResourceTRN(trn.ResourceTypeVariable, workspace.FullPath, "terraform", c.key)

	deleteInput := &pb.DeleteNamespaceVariableRequest{
		Id:      variableTRN,
		Version: c.version,
	}

	c.Logger.Debug("workspace delete-terraform-var input", "input", deleteInput)

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
   The workspace delete-terraform-var command deletes a terraform variable from a workspace.
`
}

func (*workspaceDeleteTerraformVarCommand) Usage() string {
	return "tharsis [global options] workspace delete-terraform-var [options] <workspace-id>"
}

func (*workspaceDeleteTerraformVarCommand) Example() string {
	return `
tharsis workspace delete-terraform-var \
  --key region \
  trn:workspace:ops/my-workspace
`
}

func (c *workspaceDeleteTerraformVarCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.key,
		"key",
		"",
		"Variable key.",
	)
	f.Func(
		"version",
		"Metadata version of the resource to be deleted. In most cases, this is not required.",
		func(s string) error {
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			c.version = &v
			return nil
		},
	)

	return f
}

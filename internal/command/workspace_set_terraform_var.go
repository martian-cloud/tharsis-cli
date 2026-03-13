package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type workspaceSetTerraformVarCommand struct {
	*BaseCommand

	key       string
	value     string
	sensitive bool
}

// NewWorkspaceSetTerraformVarCommandFactory returns a workspaceSetTerraformVarCommand struct.
func NewWorkspaceSetTerraformVarCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceSetTerraformVarCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceSetTerraformVarCommand) validate() error {
	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.key, validation.Required),
		validation.Field(&c.value, validation.Required),
	)
}

func (c *workspaceSetTerraformVarCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace set-terraform-var"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	// Get workspace to retrieve full path
	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{
		Id: toTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
		return 1
	}

	// Build TRN and check if variable exists
	existingVar, err := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariableByID(c.Context, &pb.GetNamespaceVariableByIDRequest{
		Id: trn.NewResourceTRN(trn.ResourceTypeVariable, workspace.FullPath, pb.VariableCategory_terraform.String(), c.key),
	})
	if err != nil && status.Code(err) != codes.NotFound {
		c.UI.ErrorWithSummary(err, "failed to check existing variable")
		return 1
	}

	if existingVar != nil {
		// Variable exists - check if sensitivity matches
		if existingVar.Sensitive != c.sensitive {
			c.UI.Errorf("cannot change sensitive flag - delete and recreate the variable instead")
			return 1
		}

		// Update existing variable
		updateInput := &pb.UpdateNamespaceVariableRequest{
			Id:    existingVar.Metadata.Id,
			Key:   c.key,
			Value: c.value,
		}

		if _, err = c.grpcClient.NamespaceVariablesClient.UpdateNamespaceVariable(c.Context, updateInput); err != nil {
			c.UI.ErrorWithSummary(err, "failed to update terraform variable")
			return 1
		}
	} else {
		// Create new variable
		createInput := &pb.CreateNamespaceVariableRequest{
			NamespacePath: workspace.FullPath,
			Category:      pb.VariableCategory_terraform,
			Key:           c.key,
			Value:         c.value,
			Sensitive:     c.sensitive,
		}

		c.Logger.Debug("workspace set-terraform-var input", "input", createInput)

		if _, err = c.grpcClient.NamespaceVariablesClient.CreateNamespaceVariable(c.Context, createInput); err != nil {
			c.UI.ErrorWithSummary(err, "failed to set terraform variable")
			return 1
		}
	}

	c.UI.Successf("Terraform variable set successfully in workspace!")
	return 0
}

func (*workspaceSetTerraformVarCommand) Synopsis() string {
	return "Set a terraform variable for a workspace."
}

func (*workspaceSetTerraformVarCommand) Description() string {
	return `
   The workspace set-terraform-var command creates or updates a terraform variable for a workspace.
`
}

func (*workspaceSetTerraformVarCommand) Usage() string {
	return "tharsis [global options] workspace set-terraform-var [options] <workspace-id>"
}

func (*workspaceSetTerraformVarCommand) Example() string {
	return `
tharsis workspace set-terraform-var \
  --key region \
  --value us-east-1 \
  trn:workspace:<workspace_path>
`
}

func (c *workspaceSetTerraformVarCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.key,
		"key",
		"",
		"Variable key.",
	)
	f.StringVar(
		&c.value,
		"value",
		"",
		"Variable value.",
	)
	f.BoolVar(
		&c.sensitive,
		"sensitive",
		false,
		"Mark variable as sensitive.",
	)

	return f
}

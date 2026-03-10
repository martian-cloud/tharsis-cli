package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceGetTerraformVarCommand struct {
	*BaseCommand

	key           string
	showSensitive bool
	toJSON        bool
}

// NewWorkspaceGetTerraformVarCommandFactory returns a workspaceGetTerraformVarCommand struct.
func NewWorkspaceGetTerraformVarCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceGetTerraformVarCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceGetTerraformVarCommand) validate() error {
	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.key, validation.Required),
	)
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

	// Get workspace to retrieve full path
	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
		return 1
	}

	// Build TRN: trn:variable:namespace-path/terraform/key
	variableTRN := trn.NewResourceTRN(trn.ResourceTypeVariable, workspace.FullPath, "terraform", c.key)

	input := &pb.GetNamespaceVariableByIDRequest{
		Id: variableTRN,
	}

	c.Logger.Debug("workspace get-terraform-var input", "input", input)

	variable, err := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariableByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get terraform variable")
		return 1
	}

	// If showing sensitive value, fetch the variable version
	if c.showSensitive && variable.Sensitive {
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

	return outputNamespaceVariable(c.UI, c.toJSON, c.showSensitive, variable)
}

func (*workspaceGetTerraformVarCommand) Synopsis() string {
	return "Get a terraform variable for a workspace."
}

func (*workspaceGetTerraformVarCommand) Description() string {
	return `
   The workspace get-terraform-var command retrieves a terraform variable for a workspace.
`
}

func (*workspaceGetTerraformVarCommand) Usage() string {
	return "tharsis [global options] workspace get-terraform-var [options] <workspace-id>"
}

func (*workspaceGetTerraformVarCommand) Example() string {
	return `
tharsis workspace get-terraform-var \
  --key region \
  trn:workspace:ops/my-workspace
`
}

func (c *workspaceGetTerraformVarCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.key,
		"key",
		"",
		"Variable key.",
	)
	f.BoolVar(
		&c.showSensitive,
		"show-sensitive",
		false,
		"Show the actual value of sensitive variables (requires appropriate permissions).",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)

	return f
}

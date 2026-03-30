package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// workspaceGetCommand is the top-level structure for the workspace get command.
type workspaceGetCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*workspaceGetCommand)(nil)

func (c *workspaceGetCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewWorkspaceGetCommandFactory returns a workspaceGetCommand struct.
func NewWorkspaceGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceGetCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetWorkspaceByIDRequest{
		Id: trn.ToTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
	}

	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
		return 1
	}

	return c.Output(workspace, c.toJSON)
}

func (*workspaceGetCommand) Synopsis() string {
	return "Get a single workspace."
}

func (*workspaceGetCommand) Usage() string {
	return "tharsis [global options] workspace get [options] <id>"
}

func (*workspaceGetCommand) Description() string {
	return `
   The workspace get command prints information about one
   workspace.
`
}

func (*workspaceGetCommand) Example() string {
	return `
tharsis workspace get trn:workspace:<workspace_path>
`
}

func (c *workspaceGetCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// workspaceGetCommand is the top-level structure for the workspace get command.
type workspaceGetCommand struct {
	*BaseCommand

	toJSON bool
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
		Id: c.arguments[0],
	}

	c.Logger.Debug("workspace get input", "input", input)

	workspace, err := c.client.WorkspacesClient.GetWorkspaceByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
		return 1
	}

	return outputWorkspace(c.UI, c.toJSON, workspace)
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
tharsis workspace get trn:workspace:ops/my-group/my-workspace
`
}

func (c *workspaceGetCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}

func outputWorkspace(ui terminal.UI, toJSON bool, workspace *pb.Workspace) int {
	if toJSON {
		if err := ui.JSON(workspace); err != nil {
			ui.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		t := terminal.NewTable("id", "name", "description", "full_path")
		t.Rich([]string{
			workspace.Metadata.Id,
			workspace.Name,
			workspace.Description,
			workspace.FullPath,
		}, nil)

		ui.Table(t)
	}

	return 0
}

package command

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
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

	return outputWorkspace(c.UI, *c.toJSON, workspace)
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
		flag.Default(false),
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
		values := []terminal.NamedValue{
			{Name: "ID", Value: workspace.Metadata.Id},
			{Name: "TRN", Value: workspace.Metadata.Trn},
			{Name: "Name", Value: workspace.Name},
			{Name: "Full Path", Value: workspace.FullPath},
			{Name: "Description", Value: workspace.Description},
			{Name: "Locked", Value: workspace.Locked},
			{Name: "Dirty State", Value: workspace.DirtyState},
			{Name: "Created By", Value: workspace.CreatedBy},
			{Name: "Created At", Value: workspace.Metadata.CreatedAt.AsTime().Local().Format(humanTimeFormat)},
			{Name: "Updated At", Value: workspace.Metadata.UpdatedAt.AsTime().Local().Format(humanTimeFormat)},
		}

		if len(workspace.Labels) > 0 {
			labels := make([]string, 0, len(workspace.Labels))
			for k, v := range workspace.Labels {
				labels = append(labels, k+"="+v)
			}

			values = append(values, terminal.NamedValue{Name: "Labels", Value: strings.Join(labels, ", ")})
		}

		ui.NamedValues(values)
	}

	return 0
}

package command

import (
	"flag"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// workspaceDeleteCommand is the top-level structure for the workspace delete command.
type workspaceDeleteCommand struct {
	*BaseCommand

	version *int64
	force   bool
}

var _ Command = (*workspaceDeleteCommand)(nil)

func (c *workspaceDeleteCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewWorkspaceDeleteCommandFactory returns a workspaceDeleteCommand struct.
func NewWorkspaceDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceDeleteCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace delete"),
		WithInputValidator(c.validate),
		WithClient(true),
		WithForcePrompt("Are you sure you want to delete this workspace?"),
	); code != 0 {
		return code
	}

	input := &pb.DeleteWorkspaceRequest{
		Id:      c.arguments[0],
		Version: c.version,
		Force:   &c.force,
	}

	c.Logger.Debug("workspace delete input", "input", input)

	if _, err := c.client.WorkspacesClient.DeleteWorkspace(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete a workspace")
		return 1
	}

	c.UI.Successf("Workspace deleted successfully!")
	return 0
}

func (*workspaceDeleteCommand) Synopsis() string {
	return "Delete a workspace."
}

func (*workspaceDeleteCommand) Usage() string {
	return "tharsis [global options] workspace delete [options] <id>"
}

func (*workspaceDeleteCommand) Description() string {
	return `
   The workspace delete command deletes a workspace. Includes
   a force flag to delete the workspace even if resources are
   deployed (dangerous!).

   Use with caution as deleting a workspace is irreversible!
`
}

func (*workspaceDeleteCommand) Example() string {
	return `
tharsis workspace delete --force trn:workspace:ops/my-group/my-workspace
`
}

func (c *workspaceDeleteCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"version",
		"Metadata version of the resource to be deleted. "+
			"In most cases, this is not required.",
		func(s string) error {
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			c.version = &v
			return nil
		},
	)
	f.BoolVar(
		&c.force,
		"force",
		false,
		"Force the workspace to delete even if resources are deployed.",
	)

	return f
}

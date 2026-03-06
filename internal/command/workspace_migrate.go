package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type workspaceMigrateCommand struct {
	*BaseCommand

	newGroupID string
	toJSON     bool
}

// NewWorkspaceMigrateCommandFactory returns a workspaceMigrateCommand struct.
func NewWorkspaceMigrateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceMigrateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceMigrateCommand) validate() error {
	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.newGroupID, validation.Required),
	)
}

func (c *workspaceMigrateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace migrate"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.MigrateWorkspaceRequest{
		WorkspaceId: c.arguments[0],
		NewGroupId:  c.newGroupID,
	}

	c.Logger.Debug("workspace migrate input", "input", input)

	workspace, err := c.client.WorkspacesClient.MigrateWorkspace(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to migrate workspace")
		return 1
	}

	return outputWorkspace(c.UI, c.toJSON, workspace)
}

func (*workspaceMigrateCommand) Synopsis() string {
	return "Migrate a workspace to a new group."
}

func (*workspaceMigrateCommand) Description() string {
	return `
   The workspace migrate command migrates a workspace to a different group.
`
}

func (*workspaceMigrateCommand) Usage() string {
	return "tharsis [global options] workspace migrate [options] <workspace-id>"
}

func (*workspaceMigrateCommand) Example() string {
	return `
tharsis workspace migrate \
  --new-group-id trn:group:ops/infrastructure \
  trn:workspace:ops/my-workspace
`
}

func (c *workspaceMigrateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.newGroupID,
		"new-group-id",
		"",
		"New parent group ID.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)

	return f
}

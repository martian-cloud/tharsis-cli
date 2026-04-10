package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type teamDeleteCommand struct {
	*BaseCommand

	version *int64
}

var _ Command = (*teamDeleteCommand)(nil)

func (c *teamDeleteCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewTeamDeleteCommandFactory returns a teamDeleteCommand struct.
func NewTeamDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &teamDeleteCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *teamDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("team delete"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	if _, err := c.grpcClient.TeamsClient.DeleteTeam(c.Context, &pb.DeleteTeamRequest{
		Id:      c.arguments[0],
		Version: c.version,
	}); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete team")
		return 1
	}

	c.UI.Successf("Team deleted successfully!")
	return 0
}

func (*teamDeleteCommand) Synopsis() string {
	return "Delete a team."
}

func (*teamDeleteCommand) Usage() string {
	return "tharsis [global options] team delete [options] <id>"
}

func (*teamDeleteCommand) Description() string {
	return `
   Permanently deletes a team. This is irreversible and
   revokes all team-based namespace access for its members.
`
}

func (*teamDeleteCommand) Example() string {
	return `
tharsis team delete <team_id>
`
}

func (c *teamDeleteCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)

	return f
}

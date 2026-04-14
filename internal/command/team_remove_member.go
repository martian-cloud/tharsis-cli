package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type teamRemoveMemberCommand struct {
	*BaseCommand

	teamName *string
}

var _ Command = (*teamRemoveMemberCommand)(nil)

func (c *teamRemoveMemberCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: username")
	}

	return nil
}

// NewTeamRemoveMemberCommandFactory returns a teamRemoveMemberCommand struct.
func NewTeamRemoveMemberCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &teamRemoveMemberCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *teamRemoveMemberCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("team remove-member"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	if _, err := c.grpcClient.TeamsClient.RemoveUserFromTeam(c.Context, &pb.RemoveUserFromTeamRequest{
		Username: c.arguments[0],
		TeamName: *c.teamName,
	}); err != nil {
		c.UI.ErrorWithSummary(err, "failed to remove member from team")
		return 1
	}

	c.UI.Successf("Team member removed successfully!")
	return 0
}

func (*teamRemoveMemberCommand) Synopsis() string {
	return "Remove a user from a team."
}

func (*teamRemoveMemberCommand) Usage() string {
	return "tharsis [global options] team remove-member [options] <username>"
}

func (*teamRemoveMemberCommand) Description() string {
	return `
   Removes a user from a team, revoking their team-based
   access to any namespaces the team is assigned to.
`
}

func (*teamRemoveMemberCommand) Example() string {
	return `
tharsis team remove-member -team-name "<team_name>" <username>
`
}

func (c *teamRemoveMemberCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.teamName,
		"team-name",
		"Name of the team.",
		flag.Required(),
	)

	return f
}

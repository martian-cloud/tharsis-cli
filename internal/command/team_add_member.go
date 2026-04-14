package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type teamAddMemberCommand struct {
	*BaseCommand

	teamName     *string
	isMaintainer *bool
	toJSON       *bool
}

var _ Command = (*teamAddMemberCommand)(nil)

func (c *teamAddMemberCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: username")
	}

	return nil
}

// NewTeamAddMemberCommandFactory returns a teamAddMemberCommand struct.
func NewTeamAddMemberCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &teamAddMemberCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *teamAddMemberCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("team add-member"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	member, err := c.grpcClient.TeamsClient.AddUserToTeam(c.Context, &pb.AddUserToTeamRequest{
		Username:     c.arguments[0],
		TeamName:     *c.teamName,
		IsMaintainer: *c.isMaintainer,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to add member to team")
		return 1
	}

	return c.Output(member, c.toJSON)
}

func (*teamAddMemberCommand) Synopsis() string {
	return "Add a user to a team."
}

func (*teamAddMemberCommand) Usage() string {
	return "tharsis [global options] team add-member [options] <username>"
}

func (*teamAddMemberCommand) Description() string {
	return `
   Adds a user to a team by username. Use -maintainer to
   grant the user team maintenance privileges.
`
}

func (*teamAddMemberCommand) Example() string {
	return `
tharsis team add-member -team-name "<team_name>" -maintainer <username>
`
}

func (c *teamAddMemberCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.teamName,
		"team-name",
		"Name of the team.",
		flag.Required(),
	)
	f.BoolVar(
		&c.isMaintainer,
		"maintainer",
		"Whether the user is a team maintainer.",
		flag.Default(false),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

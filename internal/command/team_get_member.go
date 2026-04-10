package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type teamGetMemberCommand struct {
	*BaseCommand

	teamName *string
	toJSON   *bool
}

var _ Command = (*teamGetMemberCommand)(nil)

func (c *teamGetMemberCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: username")
	}

	return nil
}

// NewTeamGetMemberCommandFactory returns a teamGetMemberCommand struct.
func NewTeamGetMemberCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &teamGetMemberCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *teamGetMemberCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("team get-member"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	member, err := c.grpcClient.TeamsClient.GetTeamMember(c.Context, &pb.GetTeamMemberRequest{
		Username: c.arguments[0],
		TeamName: *c.teamName,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get team member")
		return 1
	}

	return c.Output(member, c.toJSON)
}

func (*teamGetMemberCommand) Synopsis() string {
	return "Get a team member."
}

func (*teamGetMemberCommand) Usage() string {
	return "tharsis [global options] team get-member [options] <username>"
}

func (*teamGetMemberCommand) Description() string {
	return `
   Returns the team membership details for a user, including
   whether they are a maintainer.
`
}

func (*teamGetMemberCommand) Example() string {
	return `
tharsis team get-member -team-name "<team_name>" <username>
`
}

func (c *teamGetMemberCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.teamName,
		"team-name",
		"Name of the team.",
		flag.Required(),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

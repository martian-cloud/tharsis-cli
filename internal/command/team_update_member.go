package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type teamUpdateMemberCommand struct {
	*BaseCommand

	teamName     *string
	isMaintainer *bool
	version      *int64
	toJSON       *bool
}

var _ Command = (*teamUpdateMemberCommand)(nil)

func (c *teamUpdateMemberCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: username")
	}

	return nil
}

// NewTeamUpdateMemberCommandFactory returns a teamUpdateMemberCommand struct.
func NewTeamUpdateMemberCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &teamUpdateMemberCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *teamUpdateMemberCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("team update-member"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	member, err := c.grpcClient.TeamsClient.UpdateTeamMember(c.Context, &pb.UpdateTeamMemberRequest{
		Username:     c.arguments[0],
		TeamName:     *c.teamName,
		IsMaintainer: *c.isMaintainer,
		Version:      c.version,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update team member")
		return 1
	}

	return c.Output(member, c.toJSON)
}

func (*teamUpdateMemberCommand) Synopsis() string {
	return "Update a team member."
}

func (*teamUpdateMemberCommand) Usage() string {
	return "tharsis [global options] team update-member [options] <username>"
}

func (*teamUpdateMemberCommand) Description() string {
	return `
   Updates a team member's role, such as promoting or
   demoting maintainer status.
`
}

func (*teamUpdateMemberCommand) Example() string {
	return `
tharsis team update-member -team-name "<team_name>" -maintainer <username>
`
}

func (c *teamUpdateMemberCommand) Flags() *flag.Set {
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
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

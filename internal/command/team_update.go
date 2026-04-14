package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type teamUpdateCommand struct {
	*BaseCommand

	description *string
	version     *int64
	toJSON      *bool
}

var _ Command = (*teamUpdateCommand)(nil)

func (c *teamUpdateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewTeamUpdateCommandFactory returns a teamUpdateCommand struct.
func NewTeamUpdateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &teamUpdateCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *teamUpdateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("team update"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	updatedTeam, err := c.grpcClient.TeamsClient.UpdateTeam(c.Context, &pb.UpdateTeamRequest{
		Id:          c.arguments[0],
		Description: c.description,
		Version:     c.version,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update team")
		return 1
	}

	return c.Output(updatedTeam, c.toJSON)
}

func (*teamUpdateCommand) Synopsis() string {
	return "Update a team."
}

func (*teamUpdateCommand) Usage() string {
	return "tharsis [global options] team update [options] <id>"
}

func (*teamUpdateCommand) Description() string {
	return `
   Updates a team's description. Use
   -description to set the new value.
`
}

func (*teamUpdateCommand) Example() string {
	return `
tharsis team update -description "<description>" <team_id>
`
}

func (c *teamUpdateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.description,
		"description",
		"Description for the team.",
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

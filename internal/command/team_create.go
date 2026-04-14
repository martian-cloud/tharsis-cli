package command

import (
	"errors"

	"github.com/aws/smithy-go/ptr"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type teamCreateCommand struct {
	*BaseCommand

	description *string
	toJSON      *bool
}

var _ Command = (*teamCreateCommand)(nil)

func (c *teamCreateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: name")
	}

	return nil
}

// NewTeamCreateCommandFactory returns a teamCreateCommand struct.
func NewTeamCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &teamCreateCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *teamCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("team create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	createdTeam, err := c.grpcClient.TeamsClient.CreateTeam(c.Context, &pb.CreateTeamRequest{
		Name:        c.arguments[0],
		Description: ptr.ToString(c.description),
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create team")
		return 1
	}

	return c.Output(createdTeam, c.toJSON)
}

func (*teamCreateCommand) Synopsis() string {
	return "Create a new team."
}

func (*teamCreateCommand) Usage() string {
	return "tharsis [global options] team create [options] <name>"
}

func (*teamCreateCommand) Description() string {
	return `
   Creates a new team. Teams group users together for access
   management. Assign teams to namespaces to grant members
   access.
`
}

func (*teamCreateCommand) Example() string {
	return `
tharsis team create -description "<description>" <name>
`
}

func (c *teamCreateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.description,
		"description",
		"Description for the team.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

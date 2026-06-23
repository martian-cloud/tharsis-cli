package command

import (
	"errors"
	"fmt"
	"time"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type adminModeActivateCommand struct {
	*BaseCommand

	duration *time.Duration
	toJSON   *bool
}

var _ Command = (*adminModeActivateCommand)(nil)

func (c *adminModeActivateCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

// NewAdminModeActivateCommandFactory returns an adminModeActivateCommand struct.
func NewAdminModeActivateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &adminModeActivateCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *adminModeActivateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("admin activate"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	req := &pb.ActivateAdminModeRequest{}
	if c.duration != nil {
		if c.duration.Truncate(time.Minute) != *c.duration {
			c.UI.ErrorWithSummary(fmt.Errorf("duration must be a whole number of minutes"), "invalid duration value")
			return 1
		}
		minutes := int32(c.duration.Minutes())
		req.DurationMinutes = &minutes
	}

	user, err := c.grpcClient.UsersClient.ActivateAdminMode(c.Context, req)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to activate admin mode")
		return 1
	}

	return c.Output(user, c.toJSON)
}

func (*adminModeActivateCommand) Synopsis() string {
	return "Activate admin mode."
}

func (*adminModeActivateCommand) Usage() string {
	return "tharsis [global options] admin activate [options]"
}

func (*adminModeActivateCommand) Description() string {
	return `
   Activates admin mode for the currently
   authenticated user. Admin mode grants
   elevated privileges for a limited duration.

   If no duration is specified, the API defaults
   to 30 minutes. Maximum duration is 6 hours.
`
}

func (*adminModeActivateCommand) Example() string {
	return `
tharsis admin activate
tharsis admin activate -duration 2h
`
}

func (c *adminModeActivateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.DurationVar(
		&c.duration,
		"duration",
		"Duration for admin mode (e.g. 30m, 2h). Defaults to 30m, max 6h.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

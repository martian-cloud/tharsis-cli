package command

import (
	"errors"

	"google.golang.org/protobuf/types/known/emptypb"
)

type adminModeDeactivateCommand struct {
	*BaseCommand
}

var _ Command = (*adminModeDeactivateCommand)(nil)

func (c *adminModeDeactivateCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

// NewAdminModeDeactivateCommandFactory returns an adminModeDeactivateCommand struct.
func NewAdminModeDeactivateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &adminModeDeactivateCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *adminModeDeactivateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("admin deactivate"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	if _, err := c.grpcClient.UsersClient.DeactivateAdminMode(c.Context, &emptypb.Empty{}); err != nil {
		c.UI.ErrorWithSummary(err, "failed to deactivate admin mode")
		return 1
	}

	c.UI.Successf("Admin mode deactivated successfully.")
	return 0
}

func (*adminModeDeactivateCommand) Synopsis() string {
	return "Deactivate admin mode."
}

func (*adminModeDeactivateCommand) Usage() string {
	return "tharsis [global options] admin deactivate"
}

func (*adminModeDeactivateCommand) Description() string {
	return `
   Deactivates admin mode for the currently
   authenticated user.
`
}

func (*adminModeDeactivateCommand) Example() string {
	return `
tharsis admin deactivate
`
}

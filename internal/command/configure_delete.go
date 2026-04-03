package command

import (
	"errors"

	"github.com/posener/complete"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
)

// configureDeleteCommand is the structure for the configure delete command.
type configureDeleteCommand struct {
	*BaseCommand
}

var _ Command = (*configureDeleteCommand)(nil)

func (c *configureDeleteCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: name")
	}

	return nil
}

// NewConfigureDeleteCommandFactory returns a configureDeleteCommand struct.
func NewConfigureDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &configureDeleteCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *configureDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("configure delete"),
		WithInputValidator(c.validate),
	); code != 0 {
		return code
	}

	profileName := c.arguments[0]

	gotSettings, err := settings.ReadSettings()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to read settings")
		return 1
	}

	// Helpful errors.

	if len(gotSettings.Profiles) == 0 {
		c.UI.Errorf("no profiles currently exist. Please use the 'tharsis configure' command to create one")
		return 1
	}

	if _, ok := gotSettings.Profiles[profileName]; !ok {
		c.UI.Errorf("profile %s not found", profileName)
		return 1
	}

	delete(gotSettings.Profiles, profileName)

	if err := gotSettings.WriteSettingsFile(); err != nil {
		c.UI.ErrorWithSummary(err, "failed to write settings file")
		return 1
	}

	c.UI.Successf("Profile %s and associated credentials deleted successfully!", profileName)
	return 0
}

func (c *configureDeleteCommand) Synopsis() string {
	return "Remove a profile."
}

func (c *configureDeleteCommand) Usage() string {
	return "tharsis configure delete <name>"
}

func (c *configureDeleteCommand) Description() string {
	return `
   Removes a profile and its stored credentials.
`
}

func (c *configureDeleteCommand) Example() string {
	return `
tharsis configure delete prod-example
`
}

func (c *configureDeleteCommand) PredictArgs() complete.Predictor {
	return complete.PredictFunc(func(complete.Args) []string {
		gotSettings, err := settings.ReadSettings()
		if err != nil {
			return nil
		}

		names := make([]string, 0, len(gotSettings.Profiles))
		for name := range gotSettings.Profiles {
			names = append(names, name)
		}

		return names
	})
}

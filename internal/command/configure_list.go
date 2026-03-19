package command

import (
	"flag"
	"fmt"
	"sort"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
)

// configureListCommand is the structure for the configure list command.
type configureListCommand struct {
	*BaseCommand
}

var _ Command = (*configureListCommand)(nil)

func (c *configureListCommand) validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Empty),
	)
}

// NewConfigureListCommandFactory returns a configureListCommand struct.
func NewConfigureListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &configureListCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *configureListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("configure list"),
		WithInputValidator(c.validate),
	); code != 0 {
		return code
	}

	// Attempt to read the existing settings.
	gotSettings, err := settings.ReadSettings(nil)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to read pre-existing settings")
		return 1
	}

	return c.showAllProfiles(gotSettings)
}

func (c *configureListCommand) showAllProfiles(settings *settings.Settings) int {
	// First sort the profile names.
	profileNames := []string{}
	for profileName := range settings.Profiles {
		profileNames = append(profileNames, profileName)
	}
	sort.Strings(profileNames)

	t := terminal.NewTable("Profile", "HTTP Endpoint", "TLS Skip Verify")

	// Format and print the output.
	for _, profileName := range profileNames {
		p := settings.Profiles[profileName]
		t.Rich([]string{
			profileName,
			p.Endpoint,
			fmt.Sprintf("%t", p.TLSSkipVerify),
		}, nil)
	}

	c.UI.Table(t)

	return 0
}

func (c *configureListCommand) Synopsis() string {
	return "Show all profiles."
}

func (c *configureListCommand) Usage() string {
	return "tharsis configure list"
}

func (c *configureListCommand) Example() string {
	return `
tharsis configure list
`
}

func (c *configureListCommand) Description() string {
	return `
   The configure list command prints information about all profiles.
`
}

func (c *configureListCommand) Flags() *flag.FlagSet {
	return nil
}

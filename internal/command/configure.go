package command

import (
	"flag"
	"fmt"
	"os"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
)

// configureCommand is the structure for the configure command.
type configureCommand struct {
	*BaseCommand

	profileName   string
	httpEndpoint  string
	tlsSkipVerify bool
}

var _ Command = (*configureCommand)(nil)

func (c *configureCommand) validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Empty),
		validation.Field(&c.httpEndpoint, is.URL),
	)
}

// NewConfigureCommandFactory returns a configureCommand struct.
func NewConfigureCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &configureCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *configureCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("configure"),
		WithFlags(c.Flags()),
		WithInputValidator(c.validate),
	); code != 0 {
		return code
	}

	if (c.profileName == "") && (c.httpEndpoint == "") {
		// If options are not specified, prompt for values.

		var err error
		c.profileName, err = c.UI.Input(&terminal.Input{
			Prompt: "Enter the profile name: ",
			// Rest of the fields are ignored.
		})
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to request profile name")
			return 1
		}
		if c.profileName == "" {
			// If nothing entered manually, default to default.
			c.profileName = "default"
		}

		var httpEndpointAskPrompt string

		// Show the default endpoints if they're set.
		if c.DefaultHTTPEndpoint != "" {
			httpEndpointAskPrompt = fmt.Sprintf("Enter the HTTP API URL (default = %s): ", c.DefaultHTTPEndpoint)
		} else {
			httpEndpointAskPrompt = "Enter the HTTP API URL: "
		}

		c.httpEndpoint, err = c.UI.Input(&terminal.Input{Prompt: httpEndpointAskPrompt})
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to request HTTP endpoint URL")
			return 1
		}

		if c.httpEndpoint == "" {
			// If nothing entered manually, default to the value passed into the
			// main package at build time.
			c.httpEndpoint = c.DefaultHTTPEndpoint
		}
	}

	if c.profileName == "" || c.httpEndpoint == "" {
		// If only one option is specified, error out.
		// This can happen if the only one option is supplied or if the
		// interactive response leaves the endpoint URL blank and
		// the default value from build time is blank.
		c.UI.Errorf("Please specify all --profile=..., --http-endpoint=... options.")
		return 1
	}

	if err := c.validate(); err != nil {
		c.UI.ErrorWithSummary(err, "invalid values provided")
		return 1
	}

	// Remove any trailing slashes from the endpoint as this could
	// create problems when making requests.
	c.httpEndpoint = strings.TrimSuffix(c.httpEndpoint, "/")

	// Attempt to read the existing settings.
	gotSettings, err := settings.ReadSettings(nil)
	if err != nil {
		if err == settings.ErrNoSettings {
			c.UI.Infof("\nNo existing settings file. Generating a new one.")
			// Continue on and create the empty structure.
		} else if !os.IsNotExist(err) {
			c.UI.ErrorWithSummary(err, "failed to read pre-existing credentials")
			return 1
		}
		// It's okay if the file does not exist.  Create a new empty structure.
		gotSettings = &settings.Settings{}
	}

	return c.updateOneProfile(gotSettings)
}

func (c *configureCommand) updateOneProfile(oldSettings *settings.Settings) int {
	if oldSettings.Profiles == nil {
		// Create the profiles map for the first time.
		oldSettings.Profiles = map[string]settings.Profile{}
	}

	profile, ok := oldSettings.Profiles[c.profileName]
	if !ok {
		// Create a new profile.
		oldSettings.Profiles[c.profileName] = settings.Profile{}
	}

	if !strings.HasPrefix(c.httpEndpoint, "http://") && !strings.HasPrefix(c.httpEndpoint, "https://") {
		// Handle case where only hostname was entered by prepending HTTPS to it.
		c.httpEndpoint = fmt.Sprintf("https://%s", c.httpEndpoint)
	}

	// Show the values before they're written.

	c.UI.Output("Setting profile:")
	c.UI.NamedValues([]terminal.NamedValue{
		{Name: "Profile", Value: c.profileName},
		{Name: "HTTP endpoint", Value: c.httpEndpoint},
		{Name: "TLS Skip Verify", Value: c.tlsSkipVerify},
	})

	// Set the endpoints on the settings.
	profile.Endpoint = c.httpEndpoint
	profile.TLSSkipVerify = c.tlsSkipVerify
	oldSettings.Profiles[c.profileName] = profile

	// Write the file.
	if err := oldSettings.WriteSettingsFile(nil); err != nil {
		c.UI.ErrorWithSummary(err, "failed to write settings file")
		return 1
	}

	return 0
}

func (c *configureCommand) Synopsis() string {
	return "Create or update a profile."
}

func (c *configureCommand) Usage() string {
	return "tharsis configure [options]"
}

func (c *configureCommand) Description() string {
	return `
   The configure command creates or updates a profile. If no
   options are specified, the command prompts for values.
`
}

func (c *configureCommand) Example() string {
	return `
tharsis configure \
  --http-endpoint https://api.tharsis.example.com \
  --profile prod-example
`
}

func (c *configureCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("command options", flag.ContinueOnError)
	f.StringVar(
		&c.profileName,
		"profile",
		"",
		"The name of the profile to set.",
	)
	f.StringVar(
		&c.httpEndpoint,
		"http-endpoint",
		c.DefaultHTTPEndpoint,
		"The Tharsis HTTP API endpoint (in URL format).",
	)
	f.BoolVar(
		&c.tlsSkipVerify,
		"insecure-tls-skip-verify",
		false,
		"Allow TLS but disable verification of the gRPC server's certificate chain and hostname. "+
			"This should ONLY be true for testing as it could allow the CLI to connect to an impersonated server.",
	)

	return f
}

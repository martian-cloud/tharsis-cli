package command

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
)

// configureCommand is the top-level structure for the configure command.
type configureCommand struct {
	meta *Metadata
}

// NewConfigureCommandFactory returns a configureCommand struct.
func NewConfigureCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return configureCommand{
			meta: meta,
		}, nil
	}
}

func (cc configureCommand) Run(opts []string) int {
	cc.meta.Logger.Debugf("Starting the 'configure' command with %d arguments:", len(opts))
	for ix, arg := range opts {
		cc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	defs := buildConfigureDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(cc.meta.BinaryName+" configure", defs, opts)
	if err != nil {
		cc.meta.Logger.Errorf(output.FormatError("failed to parse configure options", err))
		return 1
	}
	if len(cmdArgs) > 0 {
		msg := fmt.Sprintf("excessive configure arguments: %s", cmdArgs)
		cc.meta.Logger.Errorf(output.FormatError(msg, nil), cc.HelpConfigure())
		return 1
	}

	profileName := getOption("profile", "", cmdOpts)[0]
	endpointURL := getOption("endpoint-url", "", cmdOpts)[0]
	if (profileName == "") && (endpointURL == "") {
		// If neither option is specified, prompt for values.

		profileName, err = cc.meta.UI.Ask("\n  Enter the profile name: ")
		if err != nil {
			cc.meta.Logger.Error(output.FormatError("failed to request profile name", err))
			return 1
		}
		if profileName == "" {
			// If nothing entered manually, default to default.
			profileName = "default"
		}

		var endpointAskPrompt string
		if cc.meta.DefaultEndpointURL != "" {
			endpointAskPrompt = fmt.Sprintf("  Enter the endpoint URL (default = %s): ", cc.meta.DefaultEndpointURL)
		} else {
			endpointAskPrompt = "  Enter the endpoint URL: "
		}
		endpointURL, err = cc.meta.UI.Ask(endpointAskPrompt)
		if err != nil {
			cc.meta.Logger.Error(output.FormatError("failed to request endpoint URL", err))
			return 1
		}
		if endpointURL == "" {
			// If nothing entered manually, default to the value passed into the
			// main package at build time.
			endpointURL = cc.meta.DefaultEndpointURL
		}
	}
	if (profileName == "") || (endpointURL == "") {
		// If only one option is specified, error out.
		// This can happen if the only one option is supplied or if the
		// interactive response leaves the endpoint URL blank and
		// the default value from build time is blank.
		cc.meta.Logger.Error(output.FormatError("Please specify both the --profile=... and --endpoint-url=... options.", nil))
		return 1
	}

	// Attempt to read the existing settings.
	gotSettings, err := settings.ReadSettings(nil)
	if err != nil {
		if err == settings.ErrNoSettings {
			cc.meta.UI.Output("No existing settings file. Generating a new one.")
			// Continue on and create the empty structure.
		} else if !os.IsNotExist(err) {
			cc.meta.UI.Output(fmt.Sprintf("Failed to read pre-existing credentials: %s", err))
			return 1
		}
		// It's okay if the file does not exist.  Create a new empty structure.
		gotSettings = &settings.Settings{}
	}

	return cc.updateOneProfile(gotSettings, profileName, strings.TrimSuffix(endpointURL, "/"))
}

func (cc configureCommand) updateOneProfile(oldSettings *settings.Settings, name, tharsisURL string) int {
	cc.meta.UI.Output("Setting profile URL:\n")
	cc.meta.UI.Output(fmt.Sprintf("Profile: %s", name))
	cc.meta.UI.Output(fmt.Sprintf("    URL: %s", tharsisURL))

	// Make the change:
	if oldSettings == nil {
		// Create the settings struct for the first time.
		oldSettings = &settings.Settings{}
	}

	if oldSettings.Profiles == nil {
		// Create the profiles map for the first time.
		oldSettings.Profiles = map[string]settings.Profile{}
	}

	profile, ok := oldSettings.Profiles[name]
	if !ok {
		// Create a new profile.
		oldSettings.Profiles[name] = settings.Profile{}
	}

	// Set the profile, even if the profile will later be deleted.
	parsedURL, err := url.Parse(tharsisURL)
	if err != nil {
		cc.meta.Logger.Error(output.FormatError(fmt.Sprintf("failed to parse URL: %s", tharsisURL), nil))
		return 1
	}
	fixParsedURL(parsedURL)
	profile.TharsisURL = parsedURL.String()
	oldSettings.Profiles[name] = profile

	// Special case to remove a profile.
	if tharsisURL == "-" {
		delete(oldSettings.Profiles, name)
	}

	// Write the file.
	err = oldSettings.WriteSettingsFile(nil)
	if err != nil {
		cc.meta.Logger.Error(output.FormatError("failed to write settings file", nil))
		return 1
	}

	return 0
}

// fixParsedURL fixes up an edge case where the user specifies only the hostname,
// in which case url.Parse() treats it as a path rather than as a hostname.
func fixParsedURL(u *url.URL) {
	if (u.Scheme == "") && (u.Host == "") && (u.Path != "") {
		u.Scheme = "https" // Tharsis server ignores HTTP, so default to HTTP
		u.Host = u.Path
		u.Path = ""
	}
}

// buildConfigureDefs returns the defs used by the configure command.
func buildConfigureDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"profile": {
			Arguments: []string{"Profile"},
			Synopsis:  "The name of the profile to set.",
		},
		"endpoint-url": {
			Arguments: []string{"Endpoint_URL"},
			Synopsis:  "Endpoint URL for this profile.",
		},
	}
}

func (cc configureCommand) Synopsis() string {
	return "Set, update, or remove a profile."
}

func (cc configureCommand) Help() string {
	return cc.HelpConfigure()
}

// HelpConfigure produces the long (many lines) help string for the 'configure' command.
func (cc configureCommand) HelpConfigure() string {
	return fmt.Sprintf(`
Usage: %s configure [options]

   The configure command sets, updates, or removes a profile.

%s


If neither option is specified, the command prompts for values.

If the URL value is "-", it deletes that profile.
`, cc.meta.BinaryName, buildHelpText(buildConfigureDefs()))
}

// The End.

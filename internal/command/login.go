package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/auth"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
)

// loginCommand is the top-level structure for the login command.
type loginCommand struct {
	meta *Metadata
}

// NewLoginCommandFactory returns a loginCommand struct.
func NewLoginCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return loginCommand{
			meta: meta,
		}, nil
	}
}

func (lc loginCommand) Run(args []string) int {
	lc.meta.Logger.Debugf("Starting the 'login' command with %d arguments:", len(args))
	for ix, arg := range args {
		lc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	_, cmdArgs, err := optparser.ParseCommandOptions(lc.meta.BinaryName+" login",
		optparser.OptionDefinitions{}, args)
	if err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to parse login options", err))
		return 1
	}
	if len(cmdArgs) > 0 {
		msg := fmt.Sprintf("excessive login arguments: %s", cmdArgs)
		lc.meta.Logger.Error(output.FormatError(msg, nil), lc.HelpLogin())
		return 1
	}

	// Cannot delay reading settings past this point.
	currentSettings, err := lc.meta.ReadSettings()
	if err != nil {
		if err != settings.ErrNoSettings || lc.meta.DefaultEndpointURL == "" {
			lc.meta.Logger.Error(output.FormatError("Failed to read settings file", err))
			return 1
		}
		// Build settings when they don't exist.
		profile := settings.Profile{TharsisURL: lc.meta.DefaultEndpointURL}
		profiles := map[string]settings.Profile{"default": profile}
		currentSettings = &settings.Settings{Profiles: profiles, CurrentProfile: &profile}
	}

	settingsFilePath, err := settings.DefaultSettingsFilename()
	if err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to get default settings file name", err))
		return 1
	}
	lc.meta.UI.Output("Tharsis has opened your default browser to complete the SSO login.")
	lc.meta.UI.Output("")
	lc.meta.UI.Output("If the login is successful, Tharsis will store the authentication token in:")
	lc.meta.UI.Output(settingsFilePath)

	// Perform SSO login using shared auth module
	ssoClient, err := auth.NewSSOClient(
		currentSettings.CurrentProfile.TharsisURL,
		auth.WithLogger(lc.meta.Logger),
		auth.WithUI(lc.meta.UI),
	)
	if err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to create SSO client", err))
		return 1
	}

	token, err := ssoClient.PerformLogin(context.Background())
	if err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to complete SSO login", err))
		return 1
	}

	// Store token
	if err := ssoClient.StoreToken(token); err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to save token", err))
		return 1
	}

	lc.meta.UI.Output(fmt.Sprintf("\n%sTharsis SSO login was successful using the %s profile.%s\n",
		green, lc.meta.CurrentProfileName, reset))
	return 0
}

func (lc loginCommand) Synopsis() string {
	return "Log in to the OAuth2 provider and return an authentication token."
}

func (lc loginCommand) Help() string {
	return lc.HelpLogin()
}

// HelpLogin produces the long (many lines) help string for the 'login' command.
func (lc loginCommand) HelpLogin() string {
	return fmt.Sprintf(`
Usage: %s [global options] sso login [options]

   The login command starts an embedded web server and opens
   a web browser page or tab pointed at said web server.
   That redirects to the OAuth2 provider's login page, where
   the user can sign in. If there is an SSO scheme active,
   that will sign in the user. The login command captures
   the authentication token for use in subsequent commands.

`, lc.meta.BinaryName)
}

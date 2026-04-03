package command

import (
	"errors"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/auth"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
)

// loginCommand is the top-level structure for the login command.
type loginCommand struct {
	*BaseCommand
}

var _ Command = (*loginCommand)(nil)

func (c *loginCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

// NewLoginCommandFactory returns a loginCommand struct.
func NewLoginCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &loginCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *loginCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("sso login"),
		WithInputValidator(c.validate),
		WithClient(false), // No auth is required for this module.
	); code != 0 {
		return code
	}

	// Cannot delay reading settings past this point.
	currentSettings, err := c.getCurrentSettings()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to read settings file")
		return 1
	}

	credsFilepath, err := settings.DefaultCredentialsFilepath()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get default credentials file name")
		return 1
	}
	c.UI.Output("Tharsis has opened your default browser to complete the SSO login.")
	c.UI.Output("")
	c.UI.Output("If the login is successful, Tharsis will store the authentication token in:")
	c.UI.Output(credsFilepath)

	// Perform SSO login using shared auth module
	ssoClient, err := auth.NewSSOAuthenticator(
		currentSettings.CurrentProfile.Endpoint,
		auth.WithLogger(c.Logger),
		auth.WithUI(c.UI),
		auth.WithGRPCClient(c.grpcClient),
	)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create SSO client")
		return 1
	}

	token, err := ssoClient.Authenticate(c.Context)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to complete SSO login")
		return 1
	}

	if err := ssoClient.StoreToken(token); err != nil {
		c.UI.ErrorWithSummary(err, "failed to save token")
		return 1
	}

	c.UI.Successf("Tharsis SSO login was successful using the %q profile!", c.CurrentProfileName)
	return 0
}

func (c *loginCommand) Synopsis() string {
	return "Log in to the OAuth2 provider and return an authentication token."
}

func (c *loginCommand) Usage() string {
	return "tharsis [global options] sso login"
}

func (c *loginCommand) Description() string {
	return `
   Starts an embedded web server and opens a browser to the
   OAuth2 provider's login page. If SSO is active, the user
   is signed in automatically. The authentication token is
   captured and stored for use in subsequent commands.
`
}

func (c *loginCommand) Example() string {
	return `
tharsis sso login
`
}

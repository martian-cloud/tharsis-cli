package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/auth"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
)

// loginCommand is the top-level structure for the login command.
type loginCommand struct {
	*BaseCommand
}

var _ Command = (*loginCommand)(nil)

func (c *loginCommand) validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Empty),
	)
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

	token, err := ssoClient.PerformLogin(c.Context)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to complete SSO login")
		return 1
	}

	if err := ssoClient.StoreToken(token); err != nil {
		c.UI.ErrorWithSummary(err, "failed to save token")
		return 1
	}

	c.UI.Successf("\nTharsis SSO login was successful using the %s profile!\n", c.CurrentProfileName)
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
   The login command starts an embedded web server and opens
   a web browser page or tab pointed at said web server.
   That redirects to the OAuth2 provider's login page, where
   the user can sign in. If there is an SSO scheme active,
   that will sign in the user. The login command captures
   the authentication token for use in subsequent commands.
`
}

func (c *loginCommand) Example() string {
	return `
tharsis sso login
`
}

func (c *loginCommand) Flags() *flag.FlagSet {
	return nil
}

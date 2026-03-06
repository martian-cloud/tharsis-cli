package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type serviceAccountCreateOIDCTokenCommand struct {
	*BaseCommand

	token  string
	toJSON bool
}

// NewServiceAccountCreateOIDCTokenCommandFactory returns a serviceAccountCreateOIDCTokenCommand struct.
func NewServiceAccountCreateOIDCTokenCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &serviceAccountCreateOIDCTokenCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *serviceAccountCreateOIDCTokenCommand) validate() error {
	const message = "service-account-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.token, validation.Required),
	)
}

func (c *serviceAccountCreateOIDCTokenCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("service-account create-oidc-token"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.CreateOIDCTokenRequest{
		ServiceAccountId: c.arguments[0],
		Token:            c.token,
	}

	c.Logger.Debug("service-account create-oidc-token input", "input", input)

	result, err := c.client.ServiceAccountsClient.CreateOIDCToken(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create token for service account")
		return 1
	}

	if c.toJSON {
		if err := c.UI.JSON(result); err != nil {
			c.UI.ErrorWithSummary(err, "failed to output JSON")
			return 1
		}
		return 0
	}

	c.UI.Output(result.Token)
	return 0
}

func (*serviceAccountCreateOIDCTokenCommand) Synopsis() string {
	return "Create a token for a service account using OIDC."
}

func (*serviceAccountCreateOIDCTokenCommand) Description() string {
	return `
   The service-account create-oidc-token command creates a token for a service account using OIDC authentication.
   The input token is issued by an identity provider specified in the service account's trust policy.
   The output token can be used to authenticate with the API.
`
}

func (*serviceAccountCreateOIDCTokenCommand) Usage() string {
	return "tharsis [global options] service-account create-oidc-token [options] <service-account-id>"
}

func (*serviceAccountCreateOIDCTokenCommand) Example() string {
	return `
tharsis service-account create-oidc-token \
  --token <oidc-token> \
  trn:service_account:ops/my-sa
`
}

func (c *serviceAccountCreateOIDCTokenCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.token,
		"token",
		"",
		"Initial authentication token from identity provider.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)

	return f
}

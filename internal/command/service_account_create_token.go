package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type serviceAccountCreateTokenCommand struct {
	*BaseCommand

	token  string
	toJSON bool
}

// NewServiceAccountCreateTokenCommandFactory returns a serviceAccountCreateTokenCommand struct.
func NewServiceAccountCreateTokenCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &serviceAccountCreateTokenCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *serviceAccountCreateTokenCommand) validate() error {
	const message = "service-account-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.token, validation.Required),
	)
}

func (c *serviceAccountCreateTokenCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("service-account create-token"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.CreateOIDCTokenRequest{
		ServiceAccountId: toTRN(trn.ResourceTypeServiceAccount, c.arguments[0]),
		Token:            c.token,
	}

	result, err := c.grpcClient.ServiceAccountsClient.CreateOIDCToken(c.Context, input)
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

func (*serviceAccountCreateTokenCommand) Synopsis() string {
	return "Create a token for a service account."
}

func (*serviceAccountCreateTokenCommand) Description() string {
	return `
   The service-account create-token command creates a token for a service account using OIDC authentication.
   The input token is issued by an identity provider specified in the service account's trust policy.
   The output token can be used to authenticate with the API.
`
}

func (*serviceAccountCreateTokenCommand) Usage() string {
	return "tharsis [global options] service-account create-token [options] <service-account-id>"
}

func (*serviceAccountCreateTokenCommand) Example() string {
	return `
tharsis service-account create-token \
  --token <oidc-token> \
  trn:service_account:<group_path>/<service_account_name>
`
}

func (c *serviceAccountCreateTokenCommand) Flags() *flag.FlagSet {
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

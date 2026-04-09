package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type serviceAccountCreateTokenCommand struct {
	*BaseCommand

	token  *string
	toJSON *bool
}

var _ Command = (*serviceAccountCreateTokenCommand)(nil)

// NewServiceAccountCreateTokenCommandFactory returns a serviceAccountCreateTokenCommand struct.
func NewServiceAccountCreateTokenCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &serviceAccountCreateTokenCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *serviceAccountCreateTokenCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: service account id")
	}

	return nil
}

func (c *serviceAccountCreateTokenCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("service-account create-token"),
		WithInputValidator(c.validate),
		WithClient(false), // Service account create token is unauthenticated.
	); code != 0 {
		return code
	}

	input := &pb.CreateOIDCTokenRequest{
		ServiceAccountId: trn.ToTRN(trn.ResourceTypeServiceAccount, c.arguments[0]),
		Token:            *c.token,
	}

	result, err := c.grpcClient.ServiceAccountsClient.CreateOIDCToken(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create token for service account")
		return 1
	}

	if *c.toJSON {
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
   Exchanges an identity provider token for a Tharsis
   API token using OIDC authentication.
`
}

func (*serviceAccountCreateTokenCommand) Usage() string {
	return "tharsis [global options] service-account create-token [options] <service-account-id>"
}

func (*serviceAccountCreateTokenCommand) Example() string {
	return `
tharsis service-account create-token \
  -token "<oidc-token>" \
  trn:service_account:<group_path>/<service_account_name>
`
}

func (c *serviceAccountCreateTokenCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.token,
		"token",
		"Initial authentication token from identity provider.",
		flag.Required(),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}

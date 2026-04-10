package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type vcsProviderResetOAuthTokenCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*vcsProviderResetOAuthTokenCommand)(nil)

func (c *vcsProviderResetOAuthTokenCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: provider id")
	}

	return nil
}

// NewVCSProviderResetOAuthTokenCommandFactory returns a vcsProviderResetOAuthTokenCommand struct.
func NewVCSProviderResetOAuthTokenCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &vcsProviderResetOAuthTokenCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *vcsProviderResetOAuthTokenCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("vcs-provider reset-oauth-token"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	result, err := c.grpcClient.VCSProvidersClient.ResetVCSProviderOAuthToken(c.Context, &pb.ResetVCSProviderOAuthTokenRequest{
		ProviderId: c.arguments[0],
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to reset VCS provider OAuth token")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*vcsProviderResetOAuthTokenCommand) Synopsis() string {
	return "Reset the OAuth token for a VCS provider."
}

func (*vcsProviderResetOAuthTokenCommand) Usage() string {
	return "tharsis [global options] vcs-provider reset-oauth-token [options] <provider-id>"
}

func (*vcsProviderResetOAuthTokenCommand) Description() string {
	return `
   Invalidates the current OAuth token for a
   VCS provider and generates a new
   authorization URL. The URL must be visited
   in a browser to reauthorize the VCS
   provider with the OAuth application.
   Useful after updating OAuth credentials or
   if the token has been compromised.
`
}

func (*vcsProviderResetOAuthTokenCommand) Example() string {
	return `
tharsis vcs-provider reset-oauth-token <provider_id>
`
}

func (c *vcsProviderResetOAuthTokenCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type vcsProviderUpdateCommand struct {
	*BaseCommand

	description       *string
	oauthClientID     *string
	oauthClientSecret *string
	version           *int64
	toJSON            *bool
}

var _ Command = (*vcsProviderUpdateCommand)(nil)

func (c *vcsProviderUpdateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewVCSProviderUpdateCommandFactory returns a vcsProviderUpdateCommand struct.
func NewVCSProviderUpdateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &vcsProviderUpdateCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *vcsProviderUpdateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("vcs-provider update"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	updatedProvider, err := c.grpcClient.VCSProvidersClient.UpdateVCSProvider(c.Context, &pb.UpdateVCSProviderRequest{
		Id:                c.arguments[0],
		Description:       c.description,
		OauthClientId:     c.oauthClientID,
		OauthClientSecret: c.oauthClientSecret,
		Version:           c.version,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update VCS provider")
		return 1
	}

	return c.Output(updatedProvider, c.toJSON)
}

func (*vcsProviderUpdateCommand) Synopsis() string {
	return "Update a VCS provider."
}

func (*vcsProviderUpdateCommand) Usage() string {
	return "tharsis [global options] vcs-provider update [options] <id>"
}

func (*vcsProviderUpdateCommand) Description() string {
	return `
   Updates a VCS provider's description and OAuth credentials
   (application ID and secret). After updating OAuth
   credentials, you may need to reset the OAuth token to
   reauthorize the connection.
`
}

func (*vcsProviderUpdateCommand) Example() string {
	return `
tharsis vcs-provider update \
  -description "<description>" \
  <vcs_provider_id>
`
}

func (c *vcsProviderUpdateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.description,
		"description",
		"Description for the VCS provider.",
	)
	f.StringVar(
		&c.oauthClientID,
		"oauth-client-id",
		"OAuth client ID.",
	)
	f.StringVar(
		&c.oauthClientSecret,
		"oauth-client-secret",
		"OAuth client secret.",
	)
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

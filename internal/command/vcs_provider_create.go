package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	"github.com/aws/smithy-go/ptr"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type vcsProviderCreateCommand struct {
	*BaseCommand

	groupID            *string
	description        *string
	providerType       *string
	url                *string
	oauthClientID      *string
	oauthClientSecret  *string
	autoCreateWebhooks *bool
	toJSON             *bool
}

var _ Command = (*vcsProviderCreateCommand)(nil)

func (c *vcsProviderCreateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: name")
	}

	return nil
}

// NewVCSProviderCreateCommandFactory returns a vcsProviderCreateCommand struct.
func NewVCSProviderCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &vcsProviderCreateCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *vcsProviderCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("vcs-provider create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	result, err := c.grpcClient.VCSProvidersClient.CreateVCSProvider(c.Context, &pb.CreateVCSProviderRequest{
		Name:               c.arguments[0],
		Description:        ptr.ToString(c.description),
		GroupId:            *c.groupID,
		Type:               pb.VCSProviderType(pb.VCSProviderType_value[*c.providerType]),
		Url:                c.url,
		OauthClientId:      *c.oauthClientID,
		OauthClientSecret:  *c.oauthClientSecret,
		AutoCreateWebhooks: *c.autoCreateWebhooks,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create VCS provider")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*vcsProviderCreateCommand) Synopsis() string {
	return "Create a new VCS provider."
}

func (*vcsProviderCreateCommand) Usage() string {
	return "tharsis [global options] vcs-provider create [options] <name>"
}

func (*vcsProviderCreateCommand) Description() string {
	return `
   Creates a new VCS provider that establishes an
   OAuth-authenticated connection between Tharsis and GitHub
   or GitLab. VCS providers are created within a group and
   inherited by child groups. Requires an OAuth application
   ID and secret from the host provider. Returns an OAuth
   authorization URL that must be visited to complete setup.
`
}

func (*vcsProviderCreateCommand) Example() string {
	return `
tharsis vcs-provider create \
  -group-id "trn:group:<group_path>" \
  -type "GITHUB" \
  -oauth-client-id "<client_id>" \
  -oauth-client-secret "<client_secret>" \
  -auto-create-webhooks \
  <name>
`
}

func (c *vcsProviderCreateCommand) Flags() *flag.Set {
	typeValues := slices.Collect(maps.Keys(pb.VCSProviderType_value))

	f := flag.NewSet("Command options")
	f.StringVar(
		&c.groupID,
		"group-id",
		"Group ID or TRN where the VCS provider will be created.",
		flag.Required(),
	)
	f.StringVar(
		&c.providerType,
		"type",
		"VCS provider type.",
		flag.Required(),
		flag.ValidValues(typeValues...),
		flag.PredictValues(typeValues...),
		flag.TransformString(strings.ToUpper),
	)
	f.StringVar(
		&c.oauthClientID,
		"oauth-client-id",
		"OAuth client ID.",
		flag.Required(),
	)
	f.StringVar(
		&c.oauthClientSecret,
		"oauth-client-secret",
		"OAuth client secret.",
		flag.Required(),
	)
	f.StringVar(
		&c.description,
		"description",
		"Description for the VCS provider.",
	)
	f.StringVar(
		&c.url,
		"url",
		"Custom URL for self-hosted VCS instances.",
	)
	f.BoolVar(
		&c.autoCreateWebhooks,
		"auto-create-webhooks",
		"Automatically create webhooks.",
		flag.Default(false),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

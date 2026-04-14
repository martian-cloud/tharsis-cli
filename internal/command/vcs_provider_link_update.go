package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type vcsProviderLinkUpdateCommand struct {
	*BaseCommand

	moduleDirectory     *string
	branch              *string
	tagRegex            *string
	globPatterns        []string
	autoSpeculativePlan *bool
	webhookDisabled     *bool
	version             *int64
	toJSON              *bool
}

var _ Command = (*vcsProviderLinkUpdateCommand)(nil)

func (c *vcsProviderLinkUpdateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewVCSProviderLinkUpdateCommandFactory returns a vcsProviderLinkUpdateCommand struct.
func NewVCSProviderLinkUpdateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &vcsProviderLinkUpdateCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *vcsProviderLinkUpdateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("vcs-provider-link update"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.UpdateWorkspaceVCSProviderLinkRequest{
		Id:                  c.arguments[0],
		ModuleDirectory:     c.moduleDirectory,
		Branch:              c.branch,
		TagRegex:            c.tagRegex,
		AutoSpeculativePlan: c.autoSpeculativePlan,
		WebhookDisabled:     c.webhookDisabled,
		Version:             c.version,
		GlobPatterns:        c.globPatterns,
	}

	result, err := c.grpcClient.VCSProvidersClient.UpdateWorkspaceVCSProviderLink(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update VCS provider link")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*vcsProviderLinkUpdateCommand) Synopsis() string {
	return "Update a workspace VCS provider link."
}

func (*vcsProviderLinkUpdateCommand) Usage() string {
	return "tharsis [global options] vcs-provider-link update [options] <id>"
}

func (*vcsProviderLinkUpdateCommand) Description() string {
	return `
   Updates an existing workspace VCS provider
   link. All fields except the repository
   path can be modified, including branch,
   module directory, tag regex, glob patterns,
   speculative plan settings, and webhook
   configuration.
`
}

func (*vcsProviderLinkUpdateCommand) Example() string {
	return `
tharsis vcs-provider-link update \
  -branch "<branch>" \
  -auto-speculative-plan \
  <link_id>
`
}

func (c *vcsProviderLinkUpdateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.moduleDirectory,
		"module-directory",
		"Subdirectory containing the Terraform module.",
	)
	f.StringVar(
		&c.branch,
		"branch",
		"Branch to track.",
	)
	f.StringVar(
		&c.tagRegex,
		"tag-regex",
		"Tag regex pattern to trigger runs.",
	)
	f.StringSliceVar(
		&c.globPatterns,
		"glob-pattern",
		"Glob pattern to filter file changes. Can be specified multiple times.",
	)
	f.BoolVar(
		&c.autoSpeculativePlan,
		"auto-speculative-plan",
		"Automatically create speculative plans for pull requests.",
	)
	f.BoolVar(
		&c.webhookDisabled,
		"webhook-disabled",
		"Disable webhook creation.",
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

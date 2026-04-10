package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type vcsProviderLinkCreateCommand struct {
	*BaseCommand

	workspaceID         *string
	providerID          *string
	repositoryPath      *string
	moduleDirectory     *string
	branch              *string
	tagRegex            *string
	globPatterns        []string
	autoSpeculativePlan *bool
	webhookDisabled     *bool
	toJSON              *bool
}

var _ Command = (*vcsProviderLinkCreateCommand)(nil)

func (c *vcsProviderLinkCreateCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

// NewVCSProviderLinkCreateCommandFactory returns a vcsProviderLinkCreateCommand struct.
func NewVCSProviderLinkCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &vcsProviderLinkCreateCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *vcsProviderLinkCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("vcs-provider-link create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.CreateWorkspaceVCSProviderLinkRequest{
		WorkspaceId:         *c.workspaceID,
		ProviderId:          *c.providerID,
		RepositoryPath:      *c.repositoryPath,
		ModuleDirectory:     c.moduleDirectory,
		Branch:              c.branch,
		TagRegex:            c.tagRegex,
		AutoSpeculativePlan: *c.autoSpeculativePlan,
		WebhookDisabled:     *c.webhookDisabled,
		GlobPatterns:        c.globPatterns,
	}

	result, err := c.grpcClient.VCSProvidersClient.CreateWorkspaceVCSProviderLink(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create VCS provider link")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*vcsProviderLinkCreateCommand) Synopsis() string {
	return "Link a workspace to a VCS provider."
}

func (*vcsProviderLinkCreateCommand) Usage() string {
	return "tharsis [global options] vcs-provider-link create [options]"
}

func (*vcsProviderLinkCreateCommand) Description() string {
	return `
   Connects a workspace to a VCS repository,
   enabling automatic runs on commits to the
   configured branch. A workspace can only be
   linked to one VCS provider. Configure glob
   patterns to trigger runs only when specific
   files change, and enable auto-speculative-plan
   for automatic plan previews on pull/merge requests.
   The repository path cannot be changed after creation.
`
}

func (*vcsProviderLinkCreateCommand) Example() string {
	return `
tharsis vcs-provider-link create \
  -workspace-id "<workspace_id>" \
  -provider-id "<provider_id>" \
  -repository-path "<repository_path>" \
  -branch "<branch>" \
  -auto-speculative-plan
`
}

func (c *vcsProviderLinkCreateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.workspaceID,
		"workspace-id",
		"Workspace ID or TRN to link.",
		flag.Required(),
	)
	f.StringVar(
		&c.providerID,
		"provider-id",
		"VCS provider ID or TRN to link.",
		flag.Required(),
	)
	f.StringVar(
		&c.repositoryPath,
		"repository-path",
		"Repository path (e.g. owner/repo).",
		flag.Required(),
	)
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
		"Glob pattern to filter file changes.",
	)
	f.BoolVar(
		&c.autoSpeculativePlan,
		"auto-speculative-plan",
		"Automatically create speculative plans for pull requests.",
		flag.Default(false),
	)
	f.BoolVar(
		&c.webhookDisabled,
		"webhook-disabled",
		"Disable webhook creation.",
		flag.Default(false),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

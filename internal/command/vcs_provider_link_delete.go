package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type vcsProviderLinkDeleteCommand struct {
	*BaseCommand

	force   *bool
	version *int64
}

var _ Command = (*vcsProviderLinkDeleteCommand)(nil)

func (c *vcsProviderLinkDeleteCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewVCSProviderLinkDeleteCommandFactory returns a vcsProviderLinkDeleteCommand struct.
func NewVCSProviderLinkDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &vcsProviderLinkDeleteCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *vcsProviderLinkDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("vcs-provider-link delete"),
		WithInputValidator(c.validate),
		WithClient(true),
		WithWarningPrompt("This will force delete the VCS provider link and remove the webhook."),
	); code != 0 {
		return code
	}

	if _, err := c.grpcClient.VCSProvidersClient.DeleteWorkspaceVCSProviderLink(c.Context, &pb.DeleteWorkspaceVCSProviderLinkRequest{
		Id:      c.arguments[0],
		Force:   c.force,
		Version: c.version,
	}); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete VCS provider link")
		return 1
	}

	c.UI.Successf("VCS provider link deleted successfully!")
	return 0
}

func (*vcsProviderLinkDeleteCommand) Synopsis() string {
	return "Delete a workspace VCS provider link."
}

func (*vcsProviderLinkDeleteCommand) Usage() string {
	return "tharsis [global options] vcs-provider-link delete [options] <id>"
}

func (*vcsProviderLinkDeleteCommand) Description() string {
	return `
   Disconnects the workspace from its VCS
   repository and removes the associated
   webhook. Use -force if the webhook cannot
   be removed from the VCS host.
`
}

func (*vcsProviderLinkDeleteCommand) Example() string {
	return `
tharsis vcs-provider-link delete -force <link_id>
`
}

func (c *vcsProviderLinkDeleteCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.force,
		"force",
		"Force delete even if the webhook cannot be removed.",
	)
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)

	return f
}

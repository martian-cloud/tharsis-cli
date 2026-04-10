package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type vcsProviderDeleteCommand struct {
	*BaseCommand

	force   *bool
	version *int64
}

var _ Command = (*vcsProviderDeleteCommand)(nil)

func (c *vcsProviderDeleteCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewVCSProviderDeleteCommandFactory returns a vcsProviderDeleteCommand struct.
func NewVCSProviderDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &vcsProviderDeleteCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *vcsProviderDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("vcs-provider delete"),
		WithInputValidator(c.validate),
		WithClient(true),
		WithWarningPrompt("This will force delete the VCS provider and unlink all connected workspaces."),
	); code != 0 {
		return code
	}

	if _, err := c.grpcClient.VCSProvidersClient.DeleteVCSProvider(c.Context, &pb.DeleteVCSProviderRequest{
		Id:      c.arguments[0],
		Force:   c.force,
		Version: c.version,
	}); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete VCS provider")
		return 1
	}

	c.UI.Successf("VCS provider deleted successfully!")
	return 0
}

func (*vcsProviderDeleteCommand) Synopsis() string {
	return "Delete a VCS provider."
}

func (*vcsProviderDeleteCommand) Usage() string {
	return "tharsis [global options] vcs-provider delete [options] <id>"
}

func (*vcsProviderDeleteCommand) Description() string {
	return `
   Permanently removes a VCS provider, severing the OAuth
   connection and unlinking all connected workspaces. This
   operation is irreversible. Use -force to delete even if
   linked to workspaces (prompts for confirmation).
`
}

func (*vcsProviderDeleteCommand) Example() string {
	return `
tharsis vcs-provider delete -force <vcs_provider_id>
`
}

func (c *vcsProviderDeleteCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.force,
		"force",
		"Force delete even if linked to workspaces.",
	)
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)

	return f
}

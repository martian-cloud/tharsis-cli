package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type vcsProviderCreateRunCommand struct {
	*BaseCommand

	referenceName *string
	isDestroy     *bool
}

var _ Command = (*vcsProviderCreateRunCommand)(nil)

func (c *vcsProviderCreateRunCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: workspace id")
	}

	return nil
}

// NewVCSProviderCreateRunCommandFactory returns a vcsProviderCreateRunCommand struct.
func NewVCSProviderCreateRunCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &vcsProviderCreateRunCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *vcsProviderCreateRunCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("vcs-provider create-run"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	if _, err := c.grpcClient.VCSProvidersClient.CreateVCSRun(c.Context, &pb.CreateVCSRunRequest{
		WorkspaceId:   c.arguments[0],
		ReferenceName: c.referenceName,
		IsDestroy:     *c.isDestroy,
	}); err != nil {
		c.UI.ErrorWithSummary(err, "failed to create VCS run")
		return 1
	}

	c.UI.Successf("VCS run created successfully!")
	return 0
}

func (*vcsProviderCreateRunCommand) Synopsis() string {
	return "Create a run from a VCS repository."
}

func (*vcsProviderCreateRunCommand) Usage() string {
	return "tharsis [global options] vcs-provider create-run [options] <workspace-id>"
}

func (*vcsProviderCreateRunCommand) Description() string {
	return `
   Manually triggers a Terraform run using
   the configuration from the workspace's
   linked VCS repository. Optionally specify
   a Git reference (branch or tag) with
   -reference-name. Use -destroy to create
   a destroy run.
`
}

func (*vcsProviderCreateRunCommand) Example() string {
	return `
tharsis vcs-provider create-run -reference-name "<reference>" <workspace_id>
`
}

func (c *vcsProviderCreateRunCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.referenceName,
		"reference-name",
		"Git reference name (e.g. refs/heads/main, refs/tags/v1.0.0).",
	)
	f.BoolVar(
		&c.isDestroy,
		"destroy",
		"Create a destroy run.",
		flag.Default(false),
	)

	return f
}

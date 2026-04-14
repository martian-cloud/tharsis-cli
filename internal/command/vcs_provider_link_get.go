package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type vcsProviderLinkGetCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*vcsProviderLinkGetCommand)(nil)

func (c *vcsProviderLinkGetCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewVCSProviderLinkGetCommandFactory returns a vcsProviderLinkGetCommand struct.
func NewVCSProviderLinkGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &vcsProviderLinkGetCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *vcsProviderLinkGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("vcs-provider-link get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	result, err := c.grpcClient.VCSProvidersClient.GetWorkspaceVCSProviderLinkByID(c.Context, &pb.GetWorkspaceVCSProviderLinkByIDRequest{
		Id: c.arguments[0],
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get VCS provider link")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*vcsProviderLinkGetCommand) Synopsis() string {
	return "Get a workspace VCS provider link."
}

func (*vcsProviderLinkGetCommand) Usage() string {
	return "tharsis [global options] vcs-provider-link get [options] <id>"
}

func (*vcsProviderLinkGetCommand) Description() string {
	return `
   Retrieves details about a workspace VCS
   provider link, including its repository
   path, branch, module directory, tag regex,
   glob patterns, and webhook settings.
`
}

func (*vcsProviderLinkGetCommand) Example() string {
	return `
tharsis vcs-provider-link get <link_id>
`
}

func (c *vcsProviderLinkGetCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

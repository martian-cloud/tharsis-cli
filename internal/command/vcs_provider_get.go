package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type vcsProviderGetCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*vcsProviderGetCommand)(nil)

func (c *vcsProviderGetCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewVCSProviderGetCommandFactory returns a vcsProviderGetCommand struct.
func NewVCSProviderGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &vcsProviderGetCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *vcsProviderGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("vcs-provider get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	provider, err := c.grpcClient.VCSProvidersClient.GetVCSProviderByID(c.Context, &pb.GetVCSProviderByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get VCS provider")
		return 1
	}

	return c.Output(provider, c.toJSON)
}

func (*vcsProviderGetCommand) Synopsis() string {
	return "Get a VCS provider."
}

func (*vcsProviderGetCommand) Usage() string {
	return "tharsis [global options] vcs-provider get [options] <id>"
}

func (*vcsProviderGetCommand) Description() string {
	return `
   Retrieves details about a VCS provider including its type
   (GitHub or GitLab), URL, auto-create webhooks setting, and
   associated group.
`
}

func (*vcsProviderGetCommand) Example() string {
	return `
tharsis vcs-provider get <vcs_provider_id>
`
}

func (c *vcsProviderGetCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

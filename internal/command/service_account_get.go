package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type serviceAccountGetCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*serviceAccountGetCommand)(nil)

func (c *serviceAccountGetCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewServiceAccountGetCommandFactory returns a serviceAccountGetCommand struct.
func NewServiceAccountGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &serviceAccountGetCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *serviceAccountGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("service-account get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	sa, err := c.grpcClient.ServiceAccountsClient.GetServiceAccountByID(c.Context, &pb.GetServiceAccountByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get service account")
		return 1
	}

	return c.Output(sa, c.toJSON)
}

func (*serviceAccountGetCommand) Synopsis() string {
	return "Get a service account."
}

func (*serviceAccountGetCommand) Usage() string {
	return "tharsis [global options] service-account get [options] <id>"
}

func (*serviceAccountGetCommand) Description() string {
	return `
   Returns a service account's details
   including its OIDC trust policies and
   associated group.
`
}

func (*serviceAccountGetCommand) Example() string {
	return `
tharsis service-account get trn:service_account:<group_path>/<service_account_name>
`
}

func (c *serviceAccountGetCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

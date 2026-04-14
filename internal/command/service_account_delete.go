package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type serviceAccountDeleteCommand struct {
	*BaseCommand

	version *int64
}

var _ Command = (*serviceAccountDeleteCommand)(nil)

func (c *serviceAccountDeleteCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewServiceAccountDeleteCommandFactory returns a serviceAccountDeleteCommand struct.
func NewServiceAccountDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &serviceAccountDeleteCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *serviceAccountDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("service-account delete"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	if _, err := c.grpcClient.ServiceAccountsClient.DeleteServiceAccount(c.Context, &pb.DeleteServiceAccountRequest{
		Id:      c.arguments[0],
		Version: c.version,
	}); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete service account")
		return 1
	}

	c.UI.Successf("Service account deleted successfully!")
	return 0
}

func (*serviceAccountDeleteCommand) Synopsis() string {
	return "Delete a service account."
}

func (*serviceAccountDeleteCommand) Usage() string {
	return "tharsis [global options] service-account delete [options] <id>"
}

func (*serviceAccountDeleteCommand) Description() string {
	return `
   Permanently deletes a service account.
   This is irreversible and revokes all
   tokens issued to the account.
`
}

func (*serviceAccountDeleteCommand) Example() string {
	return `
tharsis service-account delete trn:service_account:<group_path>/<service_account_name>
`
}

func (c *serviceAccountDeleteCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)

	return f
}

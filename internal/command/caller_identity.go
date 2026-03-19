package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// callerIdentityCommand is the top-level structure for the caller-identity command.
type callerIdentityCommand struct {
	*BaseCommand

	toJSON bool
}

var _ Command = (*callerIdentityCommand)(nil)

func (c *callerIdentityCommand) validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Empty),
	)
}

// NewCallerIdentityCommandFactory returns a callerIdentityCommand struct.
func NewCallerIdentityCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &callerIdentityCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *callerIdentityCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("caller-identity"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	resp, err := c.grpcClient.CallerIdentityClient.GetCallerIdentity(c.Context, &emptypb.Empty{})
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			c.UI.Warnf("Not authenticated. Run 'tharsis sso login' to authenticate.")
			return 1
		}

		c.UI.ErrorWithSummary(err, "failed to get caller identity")
		return 1
	}

	return c.outputCallerIdentity(resp)
}

func (*callerIdentityCommand) Synopsis() string {
	return "Get the caller's identity."
}

func (*callerIdentityCommand) Usage() string {
	return "tharsis [global options] caller-identity [options]"
}

func (*callerIdentityCommand) Description() string {
	return `
   The caller-identity command returns information about the
   authenticated caller (User or ServiceAccount).
`
}

func (*callerIdentityCommand) Example() string {
	return `
tharsis caller-identity
`
}

func (c *callerIdentityCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show output as JSON.",
	)

	return f
}

func (c *callerIdentityCommand) outputCallerIdentity(resp *pb.GetCallerIdentityResponse) int {
	if c.toJSON {
		if err := c.UI.JSON(resp); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		switch caller := resp.Caller.(type) {
		case *pb.GetCallerIdentityResponse_User:
			c.UI.NamedValues([]terminal.NamedValue{
				{Name: "Name", Value: caller.User.Username},
				{Name: "TRN", Value: caller.User.Metadata.Trn},
				{Name: "Email", Value: caller.User.Email},
				{Name: "Admin", Value: caller.User.Admin},
				{Name: "Active", Value: caller.User.Active},
			})
		case *pb.GetCallerIdentityResponse_ServiceAccount:
			c.UI.NamedValues([]terminal.NamedValue{
				{Name: "Name", Value: caller.ServiceAccount.Name},
				{Name: "TRN", Value: caller.ServiceAccount.Metadata.Trn},
				{Name: "Description", Value: caller.ServiceAccount.Description},
			})
		}
	}

	return 0
}

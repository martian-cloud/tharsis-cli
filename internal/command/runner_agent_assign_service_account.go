package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type runnerAgentAssignServiceAccountCommand struct {
	*BaseCommand
}

// NewRunnerAgentAssignServiceAccountCommandFactory returns a runnerAgentAssignServiceAccountCommand struct.
func NewRunnerAgentAssignServiceAccountCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &runnerAgentAssignServiceAccountCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *runnerAgentAssignServiceAccountCommand) validate() error {
	const message = "service account id and runner agent id are required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(2, 2).Error(message),
		),
	)
}

func (c *runnerAgentAssignServiceAccountCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("runner-agent assign-service-account"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.AssignServiceAccountToRunnerRequest{
		ServiceAccountId: toTRN(trn.ResourceTypeServiceAccount, c.arguments[0]),
		RunnerId:         toTRN(trn.ResourceTypeRunner, c.arguments[1]),
	}

	c.Logger.Debug("runner-agent assign-service-account input", "input", input)

	if _, err := c.grpcClient.RunnersClient.AssignServiceAccountToRunner(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to assign service account to runner agent")
		return 1
	}

	c.UI.Successf("Service account assigned to runner agent successfully!")
	return 0
}

func (*runnerAgentAssignServiceAccountCommand) Synopsis() string {
	return "Assign a service account to a runner agent."
}

func (*runnerAgentAssignServiceAccountCommand) Description() string {
	return `
   The runner-agent assign-service-account command assigns a service account to a runner agent.
`
}

func (*runnerAgentAssignServiceAccountCommand) Usage() string {
	return "tharsis [global options] runner-agent assign-service-account <service-account-id> <runner-id>"
}

func (*runnerAgentAssignServiceAccountCommand) Example() string {
	return `
tharsis runner-agent assign-service-account \
  trn:service_account:<group_path>/<service_account_name> \
  trn:runner:<group_path>/<runner_name>
`
}

func (c *runnerAgentAssignServiceAccountCommand) Flags() *flag.FlagSet {
	return nil
}

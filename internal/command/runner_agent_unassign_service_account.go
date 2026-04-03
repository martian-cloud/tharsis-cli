package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type runnerAgentUnassignServiceAccountCommand struct {
	*BaseCommand
}

var _ Command = (*runnerAgentUnassignServiceAccountCommand)(nil)

// NewRunnerAgentUnassignServiceAccountCommandFactory returns a runnerAgentUnassignServiceAccountCommand struct.
func NewRunnerAgentUnassignServiceAccountCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &runnerAgentUnassignServiceAccountCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *runnerAgentUnassignServiceAccountCommand) validate() error {
	if len(c.arguments) != 2 {
		return errors.New("expected exactly two arguments: service account id and runner agent id")
	}

	return nil
}

func (c *runnerAgentUnassignServiceAccountCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("runner-agent unassign-service-account"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.UnassignServiceAccountFromRunnerRequest{
		ServiceAccountId: trn.ToTRN(trn.ResourceTypeServiceAccount, c.arguments[0]),
		RunnerId:         trn.ToTRN(trn.ResourceTypeRunner, c.arguments[1]),
	}

	if _, err := c.grpcClient.RunnersClient.UnassignServiceAccountFromRunner(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to unassign service account from runner agent")
		return 1
	}

	c.UI.Successf("Service account unassigned from runner agent successfully!")
	return 0
}

func (*runnerAgentUnassignServiceAccountCommand) Synopsis() string {
	return "Unassign a service account from a runner agent."
}

func (*runnerAgentUnassignServiceAccountCommand) Description() string {
	return `
   Revokes a service account's access to a runner agent.
`
}

func (*runnerAgentUnassignServiceAccountCommand) Usage() string {
	return "tharsis [global options] runner-agent unassign-service-account <service-account-id> <runner-id>"
}

func (*runnerAgentUnassignServiceAccountCommand) Example() string {
	return `
tharsis runner-agent unassign-service-account \
  trn:service_account:<group_path>/<service_account_name> \
  trn:runner:<group_path>/<runner_name>
`
}

package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type runnerAgentUnassignServiceAccountCommand struct {
	*BaseCommand

	runnerID         string
	serviceAccountID string
}

// NewRunnerAgentUnassignServiceAccountCommandFactory returns a runnerAgentUnassignServiceAccountCommand struct.
func NewRunnerAgentUnassignServiceAccountCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &runnerAgentUnassignServiceAccountCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *runnerAgentUnassignServiceAccountCommand) validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Empty),
		validation.Field(&c.runnerID, validation.Required),
		validation.Field(&c.serviceAccountID, validation.Required),
	)
}

func (c *runnerAgentUnassignServiceAccountCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("runner-agent unassign-service-account"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.UnassignServiceAccountFromRunnerRequest{
		RunnerId:         c.runnerID,
		ServiceAccountId: c.serviceAccountID,
	}

	c.Logger.Debug("runner-agent unassign-service-account input", "input", input)

	if _, err := c.grpcClient.RunnersClient.UnassignServiceAccountFromRunner(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to unassign service account from runner agent")
		return 1
	}

	c.UI.Successf("Service account %s unassigned from runner agent successfully", c.serviceAccountID)
	return 0
}

func (*runnerAgentUnassignServiceAccountCommand) Synopsis() string {
	return "Unassign a service account from a runner agent."
}

func (*runnerAgentUnassignServiceAccountCommand) Description() string {
	return `
   The runner-agent unassign-service-account command removes a service account from a runner agent.
`
}

func (*runnerAgentUnassignServiceAccountCommand) Usage() string {
	return "tharsis [global options] runner-agent unassign-service-account [options]"
}

func (*runnerAgentUnassignServiceAccountCommand) Example() string {
	return `
tharsis runner-agent unassign-service-account \
  --runner-id trn:runner:ops/my-runner \
  --service-account-id trn:service_account:ops/my-sa
`
}

func (c *runnerAgentUnassignServiceAccountCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.runnerID,
		"runner-id",
		"",
		"The ID of the runner agent.",
	)
	f.StringVar(
		&c.serviceAccountID,
		"service-account-id",
		"",
		"The ID of the service account to unassign.",
	)

	return f
}

package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type runnerAgentAssignServiceAccountCommand struct {
	*BaseCommand

	runnerID         string
	serviceAccountID string
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
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Empty),
		validation.Field(&c.runnerID, validation.Required),
		validation.Field(&c.serviceAccountID, validation.Required),
	)
}

func (c *runnerAgentAssignServiceAccountCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("runner-agent assign-service-account"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.AssignServiceAccountToRunnerRequest{
		RunnerId:         c.runnerID,
		ServiceAccountId: c.serviceAccountID,
	}

	c.Logger.Debug("runner-agent assign-service-account input", "input", input)

	if _, err := c.client.RunnersClient.AssignServiceAccountToRunner(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to assign service account to runner agent")
		return 1
	}

	c.UI.Successf("Service account %s assigned to runner agent successfully", c.serviceAccountID)
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
	return "tharsis [global options] runner-agent assign-service-account [options]"
}

func (*runnerAgentAssignServiceAccountCommand) Example() string {
	return `
tharsis runner-agent assign-service-account \
  --runner-id trn:runner:ops/my-runner \
  --service-account-id trn:service_account:ops/my-sa
`
}

func (c *runnerAgentAssignServiceAccountCommand) Flags() *flag.FlagSet {
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
		"The ID of the service account to assign.",
	)

	return f
}

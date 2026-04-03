package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

// runnerAgentDeleteCommand is the top-level structure for the runner agent delete command.
type runnerAgentDeleteCommand struct {
	*BaseCommand

	version *int64
}

var _ Command = (*runnerAgentDeleteCommand)(nil)

func (c *runnerAgentDeleteCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewRunnerAgentDeleteCommandFactory returns a runnerAgentDeleteCommand struct.
func NewRunnerAgentDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &runnerAgentDeleteCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *runnerAgentDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("runner-agent delete"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.DeleteRunnerRequest{
		Id:      c.arguments[0],
		Version: c.version,
	}

	if _, err := c.grpcClient.RunnersClient.DeleteRunner(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete runner agent")
		return 1
	}

	c.UI.Successf("Runner agent deleted successfully!")
	return 0
}

func (*runnerAgentDeleteCommand) Synopsis() string {
	return "Delete a runner agent."
}

func (*runnerAgentDeleteCommand) Usage() string {
	return "tharsis [global options] runner-agent delete [options] <id>"
}

func (*runnerAgentDeleteCommand) Description() string {
	return `
   Permanently removes a runner agent.
`
}

func (*runnerAgentDeleteCommand) Example() string {
	return `
tharsis runner-agent delete trn:runner:<group_path>/<runner_name>
`
}

func (c *runnerAgentDeleteCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)

	return f
}

package command

import (
	"flag"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// runnerAgentDeleteCommand is the top-level structure for the runner agent delete command.
type runnerAgentDeleteCommand struct {
	*BaseCommand

	version *int64
}

var _ Command = (*runnerAgentDeleteCommand)(nil)

func (c *runnerAgentDeleteCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
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

	c.Logger.Debug("runner agent delete input", "input", input)

	if _, err := c.client.RunnersClient.DeleteRunner(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete runner agent")
		return 1
	}

	c.UI.Output("Runner agent deleted successfully!")

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
   The runner-agent delete command deletes a runner agent.
`
}

func (*runnerAgentDeleteCommand) Example() string {
	return `
tharsis runner-agent delete trn:runner:ops/prod-runner
`
}

func (c *runnerAgentDeleteCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"version",
		"Metadata version of the resource to be deleted. "+
			"In most cases, this is not required.",
		func(s string) error {
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			c.version = &v
			return nil
		},
	)

	return f
}

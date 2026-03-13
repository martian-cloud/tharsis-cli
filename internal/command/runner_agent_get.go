package command

import (
	"flag"
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// runnerAgentGetCommand is the top-level structure for the runner agent get command.
type runnerAgentGetCommand struct {
	*BaseCommand

	toJSON bool
}

var _ Command = (*runnerAgentGetCommand)(nil)

func (c *runnerAgentGetCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewRunnerAgentGetCommandFactory returns a runnerAgentGetCommand struct.
func NewRunnerAgentGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &runnerAgentGetCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *runnerAgentGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("runner-agent get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	runner, err := c.grpcClient.RunnersClient.GetRunnerByID(c.Context, &pb.GetRunnerByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get runner agent")
		return 1
	}

	return outputRunnerAgent(c.UI, c.toJSON, runner)
}

func (*runnerAgentGetCommand) Synopsis() string {
	return "Get a runner agent."
}

func (*runnerAgentGetCommand) Usage() string {
	return "tharsis [global options] runner-agent get [options] <id>"
}

func (*runnerAgentGetCommand) Description() string {
	return `
   The runner-agent get command gets a runner agent by ID.
`
}

func (*runnerAgentGetCommand) Example() string {
	return `
tharsis runner-agent get trn:runner:<group_path>/<runner_name>
`
}

func (c *runnerAgentGetCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}

func outputRunnerAgent(ui terminal.UI, toJSON bool, runner *pb.Runner) int {
	if toJSON {
		if err := ui.JSON(runner); err != nil {
			ui.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		t := terminal.NewTable("id", "name", "type", "disabled", "run_untagged_jobs", "tags")
		t.Rich([]string{
			runner.Metadata.Id,
			runner.Name,
			runner.Type,
			fmt.Sprintf("%t", runner.Disabled),
			fmt.Sprintf("%t", runner.RunUntaggedJobs),
			strings.Join(runner.Tags, ", "),
		}, nil)

		ui.Table(t)
	}

	return 0
}

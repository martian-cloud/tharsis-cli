package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

// runnerAgentUpdateCommand is the top-level structure for the runner agent update command.
type runnerAgentUpdateCommand struct {
	*BaseCommand

	description     *string
	version         *int64
	tags            []string
	disabled        *bool
	runUntaggedJobs *bool
	toJSON          *bool
}

var _ Command = (*runnerAgentUpdateCommand)(nil)

func (c *runnerAgentUpdateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewRunnerAgentUpdateCommandFactory returns a runnerAgentUpdateCommand struct.
func NewRunnerAgentUpdateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &runnerAgentUpdateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *runnerAgentUpdateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("runner-agent update"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.UpdateRunnerRequest{
		Id:              c.arguments[0],
		Description:     c.description,
		Disabled:        c.disabled,
		RunUntaggedJobs: c.runUntaggedJobs,
		Tags:            c.tags,
		Version:         c.version,
	}

	updatedRunner, err := c.grpcClient.RunnersClient.UpdateRunner(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update runner agent")
		return 1
	}

	return c.Output(updatedRunner, c.toJSON)
}

func (*runnerAgentUpdateCommand) Synopsis() string {
	return "Update a runner agent."
}

func (*runnerAgentUpdateCommand) Usage() string {
	return "tharsis [global options] runner-agent update [options] <id>"
}

func (*runnerAgentUpdateCommand) Description() string {
	return `
   Modifies an existing runner agent's configuration.
`
}

func (*runnerAgentUpdateCommand) Example() string {
	return `
tharsis runner-agent update \
  -description "Updated description" \
  -disabled true \
  -tag "prod" \
  -tag "us-west-2" \
  trn:runner:<group_path>/<runner_name>
`
}

func (c *runnerAgentUpdateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.description,
		"description",
		"Description for the runner agent.",
	)
	f.BoolVar(
		&c.disabled,
		"disabled",
		"Enable or disable the runner agent.",
	)
	f.BoolVar(
		&c.runUntaggedJobs,
		"run-untagged-jobs",
		"Allow the runner agent to execute jobs without tags.",
	)
	f.StringSliceVar(
		&c.tags,
		"tag",
		"Tag for the runner agent.",
	)
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

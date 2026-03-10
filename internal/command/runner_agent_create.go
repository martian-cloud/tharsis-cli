package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// runnerAgentCreateCommand is the top-level structure for the runner agent create command.
type runnerAgentCreateCommand struct {
	*BaseCommand

	groupID         string
	description     string
	runUntaggedJobs bool
	tags            []string
	toJSON          bool
}

var _ Command = (*runnerAgentCreateCommand)(nil)

func (c *runnerAgentCreateCommand) validate() error {
	const message = "name is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.groupID, validation.Required),
	)
}

// NewRunnerAgentCreateCommandFactory returns a runnerAgentCreateCommand struct.
func NewRunnerAgentCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &runnerAgentCreateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *runnerAgentCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("runner-agent create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.CreateRunnerRequest{
		Name:            c.arguments[0],
		Description:     c.description,
		GroupId:         c.groupID,
		RunUntaggedJobs: c.runUntaggedJobs,
		Tags:            c.tags,
	}

	c.Logger.Debug("runner agent create input", "input", input)

	createdRunner, err := c.grpcClient.RunnersClient.CreateRunner(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create a runner agent")
		return 1
	}

	return outputRunnerAgent(c.UI, c.toJSON, createdRunner)
}

func (*runnerAgentCreateCommand) Synopsis() string {
	return "Create a new runner agent."
}

func (*runnerAgentCreateCommand) Usage() string {
	return "tharsis [global options] runner-agent create [options] <name>"
}

func (*runnerAgentCreateCommand) Description() string {
	return `
   The runner-agent create command creates a new runner agent.
`
}

func (*runnerAgentCreateCommand) Example() string {
	return `
tharsis runner-agent create \
  --group-id trn:group:ops/my-group \
  --description "Production runner" \
  --run-untagged-jobs \
  --tag prod \
  --tag us-east-1 \
  prod-runner
`
}

func (c *runnerAgentCreateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.groupID,
		"group-id",
		"",
		"Group ID or TRN where the runner agent will be created.",
	)
	f.StringVar(
		&c.description,
		"description",
		"",
		"Description for the runner agent.",
	)
	f.BoolVar(
		&c.runUntaggedJobs,
		"run-untagged-jobs",
		false,
		"Allow the runner agent to execute jobs without tags.",
	)
	f.Func(
		"tag",
		"Tag for the runner agent. (This flag may be repeated)",
		func(s string) error {
			c.tags = append(c.tags, s)
			return nil
		},
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}

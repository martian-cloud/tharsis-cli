package command

import (
	"flag"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// runnerAgentUpdateCommand is the top-level structure for the runner agent update command.
type runnerAgentUpdateCommand struct {
	*BaseCommand

	description     *string
	disabled        *bool
	runUntaggedJobs *bool
	tags            []string
	version         *int64
	toJSON          bool
}

var _ Command = (*runnerAgentUpdateCommand)(nil)

func (c *runnerAgentUpdateCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
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

	c.Logger.Debug("runner agent update input", "input", input)

	updatedRunner, err := c.client.RunnersClient.UpdateRunner(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update runner agent")
		return 1
	}

	return outputRunnerAgent(c.UI, c.toJSON, updatedRunner)
}

func (*runnerAgentUpdateCommand) Synopsis() string {
	return "Update a runner agent."
}

func (*runnerAgentUpdateCommand) Usage() string {
	return "tharsis [global options] runner-agent update [options] <id>"
}

func (*runnerAgentUpdateCommand) Description() string {
	return `
   The runner-agent update command updates an existing runner agent.
`
}

func (*runnerAgentUpdateCommand) Example() string {
	return `
tharsis runner-agent update \
  --description "Updated description" \
  --disabled true \
  --tag prod \
  --tag us-west-2 \
  trn:runner:abc123
`
}

func (c *runnerAgentUpdateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"description",
		"Description for the runner agent.",
		func(s string) error {
			c.description = &s
			return nil
		},
	)
	f.Func(
		"disabled",
		"Enable or disable the runner agent (true or false).",
		func(s string) error {
			val, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}
			c.disabled = &val
			return nil
		},
	)
	f.Func(
		"run-untagged-jobs",
		"Allow the runner agent to execute jobs without tags (true or false).",
		func(s string) error {
			val, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}
			c.runUntaggedJobs = &val
			return nil
		},
	)
	f.Func(
		"tag",
		"Tag for the runner agent. (This flag may be repeated)",
		func(s string) error {
			c.tags = append(c.tags, s)
			return nil
		},
	)
	f.Func(
		"version",
		"Metadata version of the resource to be updated. "+
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
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}

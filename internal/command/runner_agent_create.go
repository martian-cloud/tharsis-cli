package command

import (
	"errors"
	"flag"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// runnerAgentCreateCommand is the top-level structure for the runner agent create command.
type runnerAgentCreateCommand struct {
	*BaseCommand

	disabled        *bool
	runnerName      string
	groupID         string
	description     string
	tags            []string
	runUntaggedJobs bool
	toJSON          bool
}

var _ Command = (*runnerAgentCreateCommand)(nil)

func (c *runnerAgentCreateCommand) validate() error {
	if len(c.arguments) > 0 && c.runnerName != "" {
		return errors.New("must supply only one of runner name argument or option, not both")
	}

	if len(c.arguments) == 0 && c.runnerName == "" {
		return errors.New("runner name argument or option is required")
	}

	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Length(0, 1)),
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

	if c.runnerName != "" {
		c.arguments = append(c.arguments, c.runnerName)
	}

	input := &pb.CreateRunnerRequest{
		Name:            c.arguments[0],
		Description:     c.description,
		GroupId:         c.groupID,
		RunUntaggedJobs: c.runUntaggedJobs,
		Tags:            c.tags,
		Disabled:        c.disabled,
	}

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
  --group-id trn:group:<group_path> \
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
	f.Func(
		"group-path",
		"Full path of group where runner will be created. Deprecated.",
		func(s string) error {
			c.groupID = trn.NewResourceTRN(trn.ResourceTypeGroup, s)
			return nil
		},
	)
	f.BoolFunc(
		"disabled",
		"Whether the runner is disabled.",
		func(s string) error {
			v, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}

			c.disabled = &v
			return nil
		})
	f.StringVar(
		&c.runnerName,
		"runner-name",
		"",
		"Name of the new runner agent. Deprecated.",
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

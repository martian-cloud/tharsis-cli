package command

import (
	"errors"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// runnerAgentCreateCommand is the top-level structure for the runner agent create command.
type runnerAgentCreateCommand struct {
	*BaseCommand

	runnerName      *string
	groupID         *string
	description     *string
	tags            []string
	disabled        *bool
	runUntaggedJobs *bool
	toJSON          *bool
}

var _ Command = (*runnerAgentCreateCommand)(nil)

func (c *runnerAgentCreateCommand) validate() error {
	if c.runnerName != nil && len(c.arguments) > 0 {
		return errors.New("must supply only one of runner name argument or -runner-name flag, not both")
	}

	if c.runnerName == nil && len(c.arguments) == 0 {
		return errors.New("runner name argument or -runner-name flag is required")
	}

	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Length(0, 1)),
		validation.Field(&c.groupID, validation.Required, validation.NotNil),
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

	// Deprecated -runner-name flag support.
	if c.runnerName != nil {
		c.arguments = append(c.arguments, *c.runnerName)
	}

	input := &pb.CreateRunnerRequest{
		Name:            c.arguments[0],
		Description:     ptr.ToString(c.description),
		GroupId:         *c.groupID,
		RunUntaggedJobs: *c.runUntaggedJobs,
		Tags:            c.tags,
		Disabled:        c.disabled,
	}

	createdRunner, err := c.grpcClient.RunnersClient.CreateRunner(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create a runner agent")
		return 1
	}

	return c.Output(createdRunner, c.toJSON)
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
  -group-id "trn:group:<group_path>" \
  -description "Production runner" \
  -run-untagged-jobs \
  -tag "prod" \
  -tag "us-east-1" \
  prod-runner
`
}

func (c *runnerAgentCreateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.groupID,
		"group-id",
		"Group ID or TRN where the runner agent will be created.",
	)
	f.StringVar(
		&c.groupID,
		"group-path",
		"Full path of group where runner will be created.",
		flag.Deprecated("use -group-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeGroup, s)
		}),
	)
	f.BoolVar(
		&c.disabled,
		"disabled",
		"Whether the runner is disabled.",
	)
	f.StringVar(
		&c.runnerName,
		"runner-name",
		"Name of the new runner agent.",
		flag.Deprecated("pass name as an argument"),
	)
	f.StringVar(
		&c.description,
		"description",
		"Description for the runner agent.",
	)
	f.BoolVar(
		&c.runUntaggedJobs,
		"run-untagged-jobs",
		"Allow the runner agent to execute jobs without tags.",
		flag.Default(false),
	)
	f.StringSliceVar(
		&c.tags,
		"tag",
		"Tag for the runner agent.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}

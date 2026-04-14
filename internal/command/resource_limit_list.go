package command

import (
	"errors"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"google.golang.org/protobuf/types/known/emptypb"
)

type resourceLimitListCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*resourceLimitListCommand)(nil)

// NewResourceLimitListCommandFactory returns a resourceLimitListCommand struct.
func NewResourceLimitListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &resourceLimitListCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *resourceLimitListCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

func (c *resourceLimitListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("resource-limit list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	result, err := c.grpcClient.ResourceLimitsClient.GetResourceLimits(c.Context, &emptypb.Empty{})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get resource limits")
		return 1
	}

	return c.OutputList(result.ResourceLimits, c.toJSON, "name", "value")
}

func (*resourceLimitListCommand) Synopsis() string {
	return "List all resource limits."
}

func (*resourceLimitListCommand) Usage() string {
	return "tharsis [global options] resource-limit list [options]"
}

func (*resourceLimitListCommand) Description() string {
	return `
   Lists all configured resource limits.
   Resource limits control the maximum
   number of resources (e.g. workspaces,
   webhooks) allowed per namespace.
`
}

func (*resourceLimitListCommand) Example() string {
	return `
tharsis resource-limit list -json
`
}

func (c *resourceLimitListCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

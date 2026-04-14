package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type resourceLimitUpdateCommand struct {
	*BaseCommand

	value   *int32
	version *int64
	toJSON  *bool
}

var _ Command = (*resourceLimitUpdateCommand)(nil)

func (c *resourceLimitUpdateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: name")
	}

	return nil
}

// NewResourceLimitUpdateCommandFactory returns a resourceLimitUpdateCommand struct.
func NewResourceLimitUpdateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &resourceLimitUpdateCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *resourceLimitUpdateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("resource-limit update"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	updatedLimit, err := c.grpcClient.ResourceLimitsClient.UpdateResourceLimit(c.Context, &pb.UpdateResourceLimitRequest{
		Name:    c.arguments[0],
		Value:   *c.value,
		Version: c.version,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update resource limit")
		return 1
	}

	return c.Output(updatedLimit, c.toJSON)
}

func (*resourceLimitUpdateCommand) Synopsis() string {
	return "Update a resource limit."
}

func (*resourceLimitUpdateCommand) Usage() string {
	return "tharsis [global options] resource-limit update [options] <name>"
}

func (*resourceLimitUpdateCommand) Description() string {
	return `
   Changes the maximum allowed count for a
   resource limit. Requires the exact limit
   name (e.g. ResourceLimitWebhooksPerNamespace).
`
}

func (*resourceLimitUpdateCommand) Example() string {
	return `
tharsis resource-limit update -value 200 ResourceLimitWebhooksPerNamespace
`
}

func (c *resourceLimitUpdateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.Int32Var(
		&c.value,
		"value",
		"New value for the resource limit.",
		flag.Required(),
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

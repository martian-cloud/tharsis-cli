package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type runnerAgentListCommand struct {
	*BaseCommand

	limit            *int32
	cursor           *string
	namespacePath    *string
	sortBy           *string
	includeInherited *bool
	toJSON           *bool
}

var _ Command = (*runnerAgentListCommand)(nil)

// NewRunnerAgentListCommandFactory returns a runnerAgentListCommand struct.
func NewRunnerAgentListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &runnerAgentListCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *runnerAgentListCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

func (c *runnerAgentListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("runner-agent list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetRunnersRequest{
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		NamespacePath:    c.namespacePath,
		IncludeInherited: *c.includeInherited,
	}

	if c.sortBy != nil {
		input.Sort = pb.RunnerSortableField(pb.RunnerSortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.RunnersClient.GetRunners(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of runner agents")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "name", "type", "group_id")
}

func (*runnerAgentListCommand) Synopsis() string {
	return "Retrieve a paginated list of runner agents."
}

func (*runnerAgentListCommand) Usage() string {
	return "tharsis [global options] runner-agent list [options]"
}

func (*runnerAgentListCommand) Description() string {
	return `
   Lists runner agents with pagination and sorting.
   Filter by namespace and use -include-inherited
   for parent group runners.
`
}

func (*runnerAgentListCommand) Example() string {
	return `
tharsis runner-agent list -namespace-path "<group_path>" -include-inherited -json
`
}

func (c *runnerAgentListCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.RunnerSortableField_value))

	f := flag.NewSet("Command options")
	f.StringVar(
		&c.namespacePath,
		"namespace-path",
		"Namespace path to list runner agents for.",
	)
	f.BoolVar(
		&c.includeInherited,
		"include-inherited",
		"Include runner agents inherited from parent groups.",
		flag.Default(false),
	)
	f.StringVar(
		&c.cursor,
		"cursor",
		"The cursor string for manual pagination.",
	)
	f.Int32Var(
		&c.limit,
		"limit",
		"Maximum number of result elements to return.",
		flag.Default(maxPaginationLimit),
		flag.ValidRange(0, int(maxPaginationLimit)),
	)
	f.StringVar(
		&c.sortBy,
		"sort-by",
		"Sort by this field.",
		flag.ValidValues(sortValues...),
		flag.PredictValues(sortValues...),
		flag.TransformString(strings.ToUpper),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

package command

import (
	"fmt"
	"maps"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceListCommand struct {
	*BaseCommand

	limit        *int32
	cursor       *string
	search       *string
	groupID      *string
	sortBy       *string
	sortOrder    *string
	labelFilters map[string]string
	toJSON       *bool
}

var _ Command = (*workspaceListCommand)(nil)

// NewWorkspaceListCommandFactory returns a workspaceListCommand struct.
func NewWorkspaceListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceListCommand{
			BaseCommand:  baseCommand,
			labelFilters: make(map[string]string),
		}, nil
	}
}

func (c *workspaceListCommand) validate() error {
	if c.sortBy != nil && c.sortOrder != nil {
		return fmt.Errorf("cannot use both -sort-by and -sort-order")
	}

	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Empty),
	)
}

func (c *workspaceListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	sortByEnum, err := parseSortField[pb.WorkspaceSortableField](
		c.sortBy,
		c.sortOrder,
		pb.WorkspaceSortableField_value,
	)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to parse sort field")
		return 1
	}

	input := &pb.GetWorkspacesRequest{
		Sort: sortByEnum,
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Search:       c.search,
		GroupId:      c.groupID,
		LabelFilters: c.labelFilters,
	}

	result, err := c.grpcClient.WorkspacesClient.GetWorkspaces(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of workspaces")
		return 1
	}

	return c.OutputList(result, c.toJSON)
}

func (*workspaceListCommand) Synopsis() string {
	return "Retrieve a paginated list of workspaces."
}

func (*workspaceListCommand) Description() string {
	return `
   The workspace list command prints information about (likely
   multiple) workspaces. Supports pagination, filtering and
   sorting the output.
`
}

func (*workspaceListCommand) Usage() string {
	return "tharsis [global options] workspace list [options]"
}

func (*workspaceListCommand) Example() string {
	return `
tharsis workspace list \
  -group-id trn:group:<group_path> \
  -label env=prod \
  -label team=platform \
  -sort-by FULL_PATH_ASC \
  -limit 5 \
  -json
`
}

func (c *workspaceListCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.WorkspaceSortableField_value))

	f := flag.NewSet("Command options")
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
		flag.PredictValues(sortValues...),
	)
	f.StringVar(
		&c.search,
		"search",
		"Filter to only workspaces containing this substring in their path.",
	)
	f.StringVar(
		&c.groupID,
		"group-id",
		"Filter to only workspaces in this group.",
	)
	f.StringVar(
		&c.groupID,
		"group-path",
		"Filter to only workspaces in this group path.",
		flag.Deprecated("use -group-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeGroup, s)
		}),
	)
	f.MapVar(
		&c.labelFilters,
		"label",
		"Filter by label (key=value).",
	)
	f.StringVar(
		&c.sortOrder,
		"sort-order",
		"Sort in this direction.",
		flag.Deprecated("use -sort-by"),
		flag.ValidValues("ASC", "DESC"),
		flag.PredictValues("ASC", "DESC"),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}

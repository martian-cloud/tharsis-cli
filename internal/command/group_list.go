package command

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupListCommand struct {
	*BaseCommand

	limit     *int32
	cursor    *string
	search    *string
	parentID  *string
	sortBy    *string
	sortOrder *string
	toJSON    *bool
}

// NewGroupListCommandFactory returns a groupListCommand struct.
func NewGroupListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupListCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupListCommand) validate() error {
	if c.sortBy != nil && c.sortOrder != nil {
		return fmt.Errorf("cannot use both -sort-by and -sort-order")
	}

	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Empty),
	)
}

func (c *groupListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetGroupsRequest{
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Search:   c.search,
		ParentId: c.parentID,
	}

	if c.sortOrder != nil {
		switch strings.ToUpper(*c.sortOrder) {
		case "ASC":
			input.Sort = pb.GroupSortableField_FULL_PATH_ASC.Enum()
		case "DESC":
			input.Sort = pb.GroupSortableField_FULL_PATH_DESC.Enum()
		}
	}

	if c.sortBy != nil {
		value := pb.GroupSortableField(pb.GroupSortableField_value[*c.sortBy])
		input.Sort = value.Enum()
	}

	result, err := c.grpcClient.GroupsClient.GetGroups(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of groups")
		return 1
	}

	return c.OutputProtoList(result, c.toJSON)
}

func (*groupListCommand) Synopsis() string {
	return "Retrieve a paginated list of groups."
}

func (*groupListCommand) Description() string {
	return `
   The group list command prints information about (likely
   multiple) groups. Supports pagination, filtering and
   sorting the output.
`
}

func (*groupListCommand) Usage() string {
	return "tharsis [global options] group list [options]"
}

func (*groupListCommand) Example() string {
	return `
tharsis group list \
  -parent-id trn:group:<parent_group_path> \
  -sort-by FULL_PATH_ASC \
  -limit 5 \
  -json
`
}

func (c *groupListCommand) Flags() *flag.Set {
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
	sortValues := slices.Collect(maps.Keys(pb.GroupSortableField_value))
	f.StringVar(
		&c.sortBy,
		"sort-by",
		"Sort by this field.",
		flag.ValidValues(sortValues...),
		flag.PredictValues(sortValues...),
	)
	f.StringVar(
		&c.search,
		"search",
		"Filter to only groups containing this substring in their path.",
	)
	f.StringVar(
		&c.parentID,
		"parent-id",
		"Filter to only direct sub-groups of this parent group.",
	)
	f.StringVar(
		&c.parentID,
		"parent-path",
		"Filter to only direct sub-groups of this parent group.",
		flag.Deprecated("use -parent-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeGroup, s)
		}),
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

package command

import (
	"flag"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupListCommand struct {
	*BaseCommand

	limit    int
	cursor   *string
	search   *string
	parentID *string
	sortBy   *pb.GroupSortableField
	toJSON   bool
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
	return validation.ValidateStruct(c,
		validation.Field(&c.limit, validation.Min(0), validation.Max(maxPaginationLimit)),
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
		Sort: c.sortBy,
		PaginationOptions: &pb.PaginationOptions{
			First: ptr.Int32(int32(c.limit)),
			After: c.cursor,
		},
		Search:   c.search,
		ParentId: c.parentID,
	}

	result, err := c.grpcClient.GroupsClient.GetGroups(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of groups")
		return 1
	}

	if c.toJSON {
		if err := c.UI.JSON(result); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		t := terminal.NewTable("id", "name", "description", "full_path")

		for _, group := range result.Groups {
			t.Rich([]string{
				group.Metadata.Id,
				group.Name,
				group.Description,
				group.FullPath,
			}, nil)
		}

		c.UI.Table(t)
		namedValues := []terminal.NamedValue{
			{Name: "Total count", Value: result.GetPageInfo().TotalCount},
			{Name: "Has Next Page", Value: result.GetPageInfo().HasNextPage},
		}
		if result.GetPageInfo().EndCursor != nil {
			namedValues = append(namedValues, terminal.NamedValue{
				Name:  "Next cursor",
				Value: result.GetPageInfo().GetEndCursor(),
			})
		}

		c.UI.NamedValues(namedValues)
	}

	return 0
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
  --parent-id trn:group:<parent_group_path> \
  --sort-by FULL_PATH_ASC \
  --limit 5 \
  --json
`
}

func (c *groupListCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"cursor",
		"The cursor string for manual pagination.",
		func(s string) error {
			c.cursor = &s
			return nil
		},
	)
	f.IntVar(
		&c.limit,
		"limit",
		maxPaginationLimit,
		"Maximum number of result elements to return.",
	)
	f.Func(
		"sort-by",
		"Sort by this field (e.g., UPDATED_AT_ASC, UPDATED_AT_DESC, FULL_PATH_ASC, FULL_PATH_DESC).",
		func(s string) error {
			value, ok := pb.GroupSortableField_value[strings.ToUpper(s)]
			if !ok {
				return fmt.Errorf("invalid sort-by value: %s (valid values: %v)", s, slices.Collect(maps.Keys(pb.GroupSortableField_value)))
			}
			c.sortBy = pb.GroupSortableField(value).Enum()
			return nil
		},
	)
	f.Func(
		"search",
		"Filter to only groups containing this substring in their path.",
		func(s string) error {
			c.search = &s
			return nil
		},
	)
	f.Func(
		"parent-id",
		"Filter to only direct sub-groups of this parent group.",
		func(s string) error {
			c.parentID = &s
			return nil
		},
	)
	f.Func(
		"parent-path",
		"Filter to only direct sub-groups of this parent group. Deprecated",
		func(s string) error {
			c.parentID = ptr.String(trn.NewResourceTRN(trn.ResourceTypeGroup, s))
			return nil
		},
	)
	f.Func(
		"sort-order",
		"Sort in this direction, ASC or DESC. Deprecated",
		func(s string) error {
			switch strings.ToUpper(s) {
			case "ASC":
				c.sortBy = pb.GroupSortableField_FULL_PATH_ASC.Enum()
			case "DESC":
				c.sortBy = pb.GroupSortableField_FULL_PATH_DESC.Enum()
			default:
				return fmt.Errorf("unknown sort order %s", s)
			}

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

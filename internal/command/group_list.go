package command

import (
	"flag"
	"fmt"
	"maps"
	"strconv"
	"strings"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type groupListCommand struct {
	*BaseCommand

	limit    *int32
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
		validation.Field(&c.limit, validation.Min(0), validation.Max(100), validation.When(c.limit != nil)),
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

	if c.limit == nil {
		c.limit = ptr.Int32(defaultPaginationLimit)
	}

	input := &pb.GetGroupsRequest{
		Sort: c.sortBy,
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Search:   c.search,
		ParentId: c.parentID,
	}

	c.Logger.Debug("group list input", "input", input)

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
  --parent-id trn:group:top-level/bottom-level \
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
	f.Func(
		"limit",
		"Maximum number of result elements to return. Defaults to 100.",
		func(s string) error {
			i, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return err
			}
			c.limit = ptr.Int32(int32(i))
			return nil
		},
	)
	f.Func(
		"sort-by",
		"Sort by this field (e.g., UPDATED_AT_ASC, UPDATED_AT_DESC, FULL_PATH_ASC, FULL_PATH_DESC).",
		func(s string) error {
			value, ok := pb.GroupSortableField_value[strings.ToUpper(s)]
			if !ok {
				return fmt.Errorf("invalid sort-by value: %s (valid values: %v)", s, maps.Keys(pb.GroupSortableField_value))
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
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}

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

type workspaceListCommand struct {
	*BaseCommand

	limit        *int32
	cursor       *string
	search       *string
	groupID      *string
	sortBy       *pb.WorkspaceSortableField
	labelFilters map[string]string
	toJSON       bool
}

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
	return validation.ValidateStruct(c,
		validation.Field(&c.limit, validation.Min(0), validation.Max(maxPaginationLimit), validation.When(c.limit != nil)),
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

	if c.limit == nil {
		c.limit = ptr.Int32(maxPaginationLimit)
	}

	input := &pb.GetWorkspacesRequest{
		Sort: c.sortBy,
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Search:       c.search,
		GroupId:      c.groupID,
		LabelFilters: c.labelFilters,
	}

	c.Logger.Debug("workspace list input", "input", input)

	result, err := c.grpcClient.WorkspacesClient.GetWorkspaces(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of workspaces")
		return 1
	}

	if c.toJSON {
		if err := c.UI.JSON(result); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		t := terminal.NewTable("id", "name", "description", "full_path")

		for _, workspace := range result.Workspaces {
			t.Rich([]string{
				workspace.Metadata.Id,
				workspace.Name,
				workspace.Description,
				workspace.FullPath,
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
  --group-id trn:group:top-level \
  --label env=prod \
  --label team=platform \
  --sort-by FULL_PATH_ASC \
  --limit 5 \
  --json
`
}

func (c *workspaceListCommand) Flags() *flag.FlagSet {
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
			value, ok := pb.WorkspaceSortableField_value[strings.ToUpper(s)]
			if !ok {
				return fmt.Errorf("invalid sort-by value: %s (valid values: %v)", s, maps.Keys(pb.WorkspaceSortableField_value))
			}
			c.sortBy = pb.WorkspaceSortableField(value).Enum()
			return nil
		},
	)
	f.Func(
		"search",
		"Filter to only workspaces containing this substring in their path.",
		func(s string) error {
			c.search = &s
			return nil
		},
	)
	f.Func(
		"group-id",
		"Filter to only workspaces in this group.",
		func(s string) error {
			c.groupID = &s
			return nil
		},
	)
	f.Func(
		"label",
		"Filter by label (key=value). This flag may be repeated.",
		func(s string) error {
			parts := strings.SplitN(s, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid label filter format: %s (must be key=value)", s)
			}
			c.labelFilters[parts[0]] = parts[1]
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

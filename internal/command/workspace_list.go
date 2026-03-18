package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceListCommand struct {
	*BaseCommand

	limit        int
	sortOrder    *string
	cursor       *string
	search       *string
	groupID      *string
	sortBy       *string
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
		validation.Field(&c.limit, validation.Min(0), validation.Max(maxPaginationLimit)),
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
			First: ptr.Int32(int32(c.limit)),
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
  --group-id trn:group:<group_path> \
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
			// TODO: Update to use PB types and validate with PB map once deprecation is done.
			switch v := strings.ToUpper(s); v {
			case pb.WorkspaceSortableField_FULL_PATH_ASC.String(),
				pb.WorkspaceSortableField_FULL_PATH_DESC.String(),
				pb.WorkspaceSortableField_UPDATED_AT_ASC.String(),
				pb.WorkspaceSortableField_UPDATED_AT_DESC.String():
				c.sortBy = &v
			case "UPDATED": // Deprecated.
				c.sortBy = ptr.String("UPDATED_AT")
			case "PATH": // Deprecated.
				c.sortBy = ptr.String("FULL_PATH")
			default:
				return fmt.Errorf("unknown sort by option %s", s)
			}

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
		"group-path",
		"Filter to only workspaces in this group path. Deprecated.",
		func(s string) error {
			c.groupID = ptr.String(trn.NewResourceTRN(trn.ResourceTypeGroup, s))
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
	f.Func(
		"sort-order",
		"Sort in this direction, ASC or DESC. Deprecated",
		func(s string) error {
			switch strings.ToUpper(s) {
			case "ASC", "DESC":
				c.sortOrder = &s
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

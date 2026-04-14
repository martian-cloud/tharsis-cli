package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

// roleListCommand is the top-level structure for the role list command.
type roleListCommand struct {
	*BaseCommand

	limit  *int32
	cursor *string
	search *string
	sortBy *string
	toJSON *bool
}

var _ Command = (*roleListCommand)(nil)

// NewRoleListCommandFactory returns a roleListCommand struct.
func NewRoleListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &roleListCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *roleListCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

func (c *roleListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("role list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetRolesRequest{
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Search: c.search,
	}

	if c.sortBy != nil {
		input.Sort = pb.RoleSortableField(pb.RoleSortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.RolesClient.GetRoles(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of roles")
		return 1
	}

	return c.OutputList(result, c.toJSON, "name", "description", "created_by")
}

func (*roleListCommand) Synopsis() string {
	return "Retrieve a paginated list of roles."
}

func (*roleListCommand) Description() string {
	return `
   Returns a paginated list of roles with
   sorting support. Use -search to filter
   roles by name.
`
}

func (*roleListCommand) Usage() string {
	return "tharsis [global options] role list [options]"
}

func (*roleListCommand) Example() string {
	return `
tharsis role list \
  -sort-by "NAME_ASC" \
  -limit 5 \
  -json
`
}

func (c *roleListCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.RoleSortableField_value))

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
		flag.ValidValues(sortValues...),
		flag.PredictValues(sortValues...),
		flag.TransformString(strings.ToUpper),
	)
	f.StringVar(
		&c.search,
		"search",
		"Filter to only roles containing this substring in their name.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

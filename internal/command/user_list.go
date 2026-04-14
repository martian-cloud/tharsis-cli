package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type userListCommand struct {
	*BaseCommand

	limit  *int32
	cursor *string
	search *string
	sortBy *string
	toJSON *bool
}

var _ Command = (*userListCommand)(nil)

// NewUserListCommandFactory returns a userListCommand struct.
func NewUserListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &userListCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *userListCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

func (c *userListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("user list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetUsersRequest{
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Search: c.search,
	}

	if c.sortBy != nil {
		input.Sort = pb.UserSortableField(pb.UserSortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.UsersClient.GetUsers(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of users")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "username", "email", "admin", "active")
}

func (*userListCommand) Synopsis() string {
	return "Retrieve a paginated list of users."
}

func (*userListCommand) Usage() string {
	return "tharsis [global options] user list [options]"
}

func (*userListCommand) Description() string {
	return `
   Returns a paginated list of users with
   sorting support. Use -search to filter
   by username or email.
`
}

func (*userListCommand) Example() string {
	return `
tharsis user list -search "<name>" -json
`
}

func (c *userListCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.UserSortableField_value))

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
		"Filter to users containing this substring.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

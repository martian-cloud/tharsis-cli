package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type teamListMembersCommand struct {
	*BaseCommand

	teamID *string
	limit  *int32
	cursor *string
	sortBy *string
	toJSON *bool
}

var _ Command = (*teamListMembersCommand)(nil)

func (c *teamListMembersCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

// NewTeamListMembersCommandFactory returns a teamListMembersCommand struct.
func NewTeamListMembersCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &teamListMembersCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *teamListMembersCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("team list-members"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetTeamMembersRequest{
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		TeamId: c.teamID,
	}

	if c.sortBy != nil {
		input.Sort = pb.TeamMemberSortableField(pb.TeamMemberSortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.TeamsClient.GetTeamMembers(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to list team members")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "user_id", "is_maintainer")
}

func (*teamListMembersCommand) Synopsis() string {
	return "List members of a team."
}

func (*teamListMembersCommand) Usage() string {
	return "tharsis [global options] team list-members [options]"
}

func (*teamListMembersCommand) Description() string {
	return `
   Returns a paginated list of team members with their roles.
   Use -sort-by to order results by username.
`
}

func (*teamListMembersCommand) Example() string {
	return `
tharsis team list-members -team-id "<team_id>" -sort-by "USERNAME_ASC" -limit 5
`
}

func (c *teamListMembersCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.TeamMemberSortableField_value))

	f := flag.NewSet("Command options")
	f.StringVar(
		&c.teamID,
		"team-id",
		"ID of the team.",
		flag.Required(),
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

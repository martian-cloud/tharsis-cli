package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type teamListCommand struct {
	*BaseCommand

	limit          *int32
	cursor         *string
	teamNamePrefix *string
	userID         *string
	sortBy         *string
	toJSON         *bool
}

var _ Command = (*teamListCommand)(nil)

// NewTeamListCommandFactory returns a teamListCommand struct.
func NewTeamListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &teamListCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *teamListCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

func (c *teamListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("team list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetTeamsRequest{
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		TeamNamePrefix: c.teamNamePrefix,
		UserId:         c.userID,
	}

	if c.sortBy != nil {
		input.Sort = pb.TeamSortableField(pb.TeamSortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.TeamsClient.GetTeams(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of teams")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "name", "description")
}

func (*teamListCommand) Synopsis() string {
	return "Retrieve a paginated list of teams."
}

func (*teamListCommand) Usage() string {
	return "tharsis [global options] team list [options]"
}

func (*teamListCommand) Description() string {
	return `
   Returns a paginated list of teams. Filter by name prefix
   using -name-prefix or by teams containing a specific user
   using -user-id.
`
}

func (*teamListCommand) Example() string {
	return `
tharsis team list -sort-by "NAME_ASC" -limit 5 -json
`
}

func (c *teamListCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.TeamSortableField_value))

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
		&c.teamNamePrefix,
		"name-prefix",
		"Filter to teams whose name starts with this prefix.",
	)
	f.StringVar(
		&c.userID,
		"user-id",
		"Filter to teams that contain this user.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

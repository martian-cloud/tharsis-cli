package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type stateVersionListCommand struct {
	*BaseCommand

	limit       *int32
	cursor      *string
	workspaceID *string
	sortBy      *string
	toJSON      *bool
}

var _ Command = (*stateVersionListCommand)(nil)

// NewStateVersionListCommandFactory returns a stateVersionListCommand struct.
func NewStateVersionListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &stateVersionListCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *stateVersionListCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

func (c *stateVersionListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("state-version list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetStateVersionsRequest{
		WorkspaceId: *c.workspaceID,
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
	}

	if c.sortBy != nil {
		input.Sort = pb.StateVersionSortableField(pb.StateVersionSortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.StateVersionsClient.GetStateVersions(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of state versions")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "run_id", "created_by")
}

func (*stateVersionListCommand) Synopsis() string {
	return "Retrieve a paginated list of state versions."
}

func (*stateVersionListCommand) Usage() string {
	return "tharsis [global options] state-version list [options]"
}

func (*stateVersionListCommand) Description() string {
	return `
   Lists state versions for a workspace with pagination and sorting.
`
}

func (*stateVersionListCommand) Example() string {
	return `
tharsis state-version list -workspace-id "trn:workspace:<group_path>/<workspace_name>" -json
`
}

func (c *stateVersionListCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.StateVersionSortableField_value))

	f := flag.NewSet("Command options")
	f.StringVar(
		&c.workspaceID,
		"workspace-id",
		"Workspace ID or TRN to list state versions for.",
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

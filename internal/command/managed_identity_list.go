package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type managedIdentityListCommand struct {
	*BaseCommand

	limit            *int32
	cursor           *string
	search           *string
	namespacePath    *string
	sortBy           *string
	includeInherited *bool
	toJSON           *bool
}

var _ Command = (*managedIdentityListCommand)(nil)

// NewManagedIdentityListCommandFactory returns a managedIdentityListCommand struct.
func NewManagedIdentityListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &managedIdentityListCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *managedIdentityListCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

func (c *managedIdentityListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("managed-identity list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetManagedIdentitiesRequest{
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Search:           c.search,
		NamespacePath:    *c.namespacePath,
		IncludeInherited: *c.includeInherited,
	}

	if c.sortBy != nil {
		input.Sort = pb.ManagedIdentitySortableField(pb.ManagedIdentitySortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.ManagedIdentitiesClient.GetManagedIdentities(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of managed identities")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "name", "type", "group_id")
}

func (*managedIdentityListCommand) Synopsis() string {
	return "Retrieve a paginated list of managed identities."
}

func (*managedIdentityListCommand) Usage() string {
	return "tharsis [global options] managed-identity list [options]"
}

func (*managedIdentityListCommand) Description() string {
	return `
   Lists managed identities within a namespace.
   Identities are inherited from parent groups and
   can be filtered with -include-inherited.
`
}

func (*managedIdentityListCommand) Example() string {
	return `
tharsis managed-identity list -namespace-path "<group_path>" -include-inherited -json
`
}

func (c *managedIdentityListCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.ManagedIdentitySortableField_value))

	f := flag.NewSet("Command options")
	f.StringVar(
		&c.namespacePath,
		"namespace-path",
		"Namespace path to list managed identities for.",
		flag.Required(),
	)
	f.BoolVar(
		&c.includeInherited,
		"include-inherited",
		"Include managed identities inherited from parent groups.",
		flag.Default(false),
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
	f.StringVar(
		&c.search,
		"search",
		"Filter to managed identities containing this substring.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

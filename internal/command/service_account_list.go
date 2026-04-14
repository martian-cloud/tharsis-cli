package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type serviceAccountListCommand struct {
	*BaseCommand

	limit            *int32
	cursor           *string
	search           *string
	namespacePath    *string
	runnerID         *string
	sortBy           *string
	includeInherited *bool
	toJSON           *bool
}

var _ Command = (*serviceAccountListCommand)(nil)

// NewServiceAccountListCommandFactory returns a serviceAccountListCommand struct.
func NewServiceAccountListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &serviceAccountListCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *serviceAccountListCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

func (c *serviceAccountListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("service-account list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetServiceAccountsRequest{
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Search:           c.search,
		NamespacePath:    *c.namespacePath,
		RunnerId:         c.runnerID,
		IncludeInherited: *c.includeInherited,
	}

	if c.sortBy != nil {
		input.Sort = pb.ServiceAccountSortableField(pb.ServiceAccountSortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.ServiceAccountsClient.GetServiceAccounts(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of service accounts")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "name", "description", "client_credentials_enabled", "created_by")
}

func (*serviceAccountListCommand) Synopsis() string {
	return "Retrieve a paginated list of service accounts."
}

func (*serviceAccountListCommand) Usage() string {
	return "tharsis [global options] service-account list [options]"
}

func (*serviceAccountListCommand) Description() string {
	return `
   Lists service accounts within a namespace
   with pagination and sorting.
`
}

func (*serviceAccountListCommand) Example() string {
	return `
tharsis service-account list -namespace-path "<group_path>" -include-inherited -json
`
}

func (c *serviceAccountListCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.ServiceAccountSortableField_value))

	f := flag.NewSet("Command options")
	f.StringVar(
		&c.namespacePath,
		"namespace-path",
		"Namespace path to list service accounts for.",
		flag.Required(),
	)
	f.BoolVar(
		&c.includeInherited,
		"include-inherited",
		"Include service accounts inherited from parent groups.",
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
		"Filter to service accounts containing this substring.",
	)
	f.StringVar(
		&c.runnerID,
		"runner-id",
		"Filter to service accounts assigned to this runner.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type membershipListCommand struct {
	*BaseCommand

	limit            *int32
	cursor           *string
	sortBy           *string
	userID           *string
	serviceAccountID *string
	toJSON           *bool
}

var _ Command = (*membershipListCommand)(nil)

// NewMembershipListCommandFactory returns a membershipListCommand struct.
func NewMembershipListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &membershipListCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *membershipListCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	if c.userID == nil && c.serviceAccountID == nil {
		return errors.New("one of -user-id or -service-account-id is required")
	}

	return nil
}

func (c *membershipListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("membership list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetNamespaceMembershipsForSubjectRequest{
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		UserId:           c.userID,
		ServiceAccountId: c.serviceAccountID,
	}

	if c.sortBy != nil {
		input.Sort = pb.NamespaceMembershipSortableField(pb.NamespaceMembershipSortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.NamespaceMembershipsClient.GetNamespaceMembershipsForSubject(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to list memberships")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "role_id", "namespace_path")
}

func (*membershipListCommand) Synopsis() string {
	return "List namespace memberships for a user or service account."
}

func (*membershipListCommand) Description() string {
	return `
   Lists all namespace memberships for a user or
   service account, showing which namespaces the
   subject has access to and their assigned role.
   Specify exactly one of -user-id or
   -service-account-id.
`
}

func (*membershipListCommand) Usage() string {
	return "tharsis [global options] membership list [options]"
}

func (*membershipListCommand) Example() string {
	return `
tharsis membership list -user-id <user_id>
`
}

func (c *membershipListCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.NamespaceMembershipSortableField_value))

	f := flag.NewSet("Command options")
	f.StringVar(
		&c.userID,
		"user-id",
		"List memberships for this user.",
	)
	f.StringVar(
		&c.serviceAccountID,
		"service-account-id",
		"List memberships for this service account.",
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

	f.MutuallyExclusive("user-id", "service-account-id")

	return f
}

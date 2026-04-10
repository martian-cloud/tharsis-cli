package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type gpgKeyListCommand struct {
	*BaseCommand

	limit            *int32
	cursor           *string
	namespacePath    *string
	sortBy           *string
	includeInherited *bool
	toJSON           *bool
}

var _ Command = (*gpgKeyListCommand)(nil)

// NewGPGKeyListCommandFactory returns a gpgKeyListCommand struct.
func NewGPGKeyListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &gpgKeyListCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *gpgKeyListCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

func (c *gpgKeyListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("gpg-key list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetGPGKeysRequest{
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		NamespacePath:    *c.namespacePath,
		IncludeInherited: *c.includeInherited,
	}

	if c.sortBy != nil {
		input.Sort = pb.GPGKeySortableField(pb.GPGKeySortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.GPGKeysClient.GetGPGKeys(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of GPG keys")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "gpg_key_id", "fingerprint", "created_by")
}

func (*gpgKeyListCommand) Synopsis() string {
	return "Retrieve a paginated list of GPG keys."
}

func (*gpgKeyListCommand) Usage() string {
	return "tharsis [global options] gpg-key list [options]"
}

func (*gpgKeyListCommand) Description() string {
	return `
   Lists GPG keys scoped to a namespace.
   Use -include-inherited to also show keys
   from parent groups. Supports pagination
   and sorting.
`
}

func (*gpgKeyListCommand) Example() string {
	return `
tharsis gpg-key list -namespace-path "<group_path>" -include-inherited -json
`
}

func (c *gpgKeyListCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.GPGKeySortableField_value))

	f := flag.NewSet("Command options")
	f.StringVar(
		&c.namespacePath,
		"namespace-path",
		"Namespace path to list GPG keys for.",
		flag.Required(),
	)
	f.BoolVar(
		&c.includeInherited,
		"include-inherited",
		"Include GPG keys inherited from parent groups.",
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
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type terraformProviderMirrorListPlatformMirrorsCommand struct {
	*BaseCommand

	limit        *int32
	cursor       *string
	sortBy       *string
	os           *string
	architecture *string
	toJSON       *bool
}

var _ Command = (*terraformProviderMirrorListPlatformMirrorsCommand)(nil)

// NewTerraformProviderMirrorListPlatformMirrorsCommandFactory returns a terraformProviderMirrorListPlatformMirrorsCommand struct.
func NewTerraformProviderMirrorListPlatformMirrorsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderMirrorListPlatformMirrorsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderMirrorListPlatformMirrorsCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: version mirror id")
	}

	return nil
}

func (c *terraformProviderMirrorListPlatformMirrorsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider-mirror list-platform-mirrors"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetTerraformProviderPlatformMirrorsRequest{
		VersionMirrorId: c.arguments[0],
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Os:           c.os,
		Architecture: c.architecture,
	}

	if c.sortBy != nil {
		input.Sort = pb.TerraformProviderPlatformMirrorSortableField(pb.TerraformProviderPlatformMirrorSortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.TerraformProviderMirrorsClient.GetTerraformProviderPlatformMirrors(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of provider platform mirrors")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "os", "architecture")
}

func (*terraformProviderMirrorListPlatformMirrorsCommand) Synopsis() string {
	return "List platform mirrors for a provider version mirror."
}

func (*terraformProviderMirrorListPlatformMirrorsCommand) Description() string {
	return `
   Lists mirrored platforms for a provider version.
   Filter by OS or architecture.
`
}

func (*terraformProviderMirrorListPlatformMirrorsCommand) Usage() string {
	return "tharsis [global options] terraform-provider-mirror list-platform-mirrors [options] <version-mirror-id>"
}

func (*terraformProviderMirrorListPlatformMirrorsCommand) Example() string {
	return `
tharsis terraform-provider-mirror list-platform-mirrors <version_mirror_id>
`
}

func (c *terraformProviderMirrorListPlatformMirrorsCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.TerraformProviderPlatformMirrorSortableField_value))

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
		&c.os,
		"os",
		"Filter to platforms with this OS.",
	)
	f.StringVar(
		&c.architecture,
		"architecture",
		"Filter to platforms with this architecture.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}

package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupListMembershipsCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*groupListMembershipsCommand)(nil)

// NewGroupListMembershipsCommandFactory returns a groupListMembershipsCommand struct.
func NewGroupListMembershipsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupListMembershipsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupListMembershipsCommand) validate() error {
	const message = "group-path is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *groupListMembershipsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group list-memberships"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: trn.ToTRN(trn.ResourceTypeGroup, c.arguments[0])})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group")
		return 1
	}

	input := &pb.GetNamespaceMembershipsForNamespaceRequest{
		NamespacePath: group.FullPath,
	}

	result, err := c.grpcClient.NamespaceMembershipsClient.GetNamespaceMembershipsForNamespace(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of group memberships")
		return 1
	}

	return c.OutputProtoList(result, c.toJSON)
}

func (*groupListMembershipsCommand) Synopsis() string {
	return "Retrieve a list of group memberships."
}

func (*groupListMembershipsCommand) Description() string {
	return `
   The group list-memberships command prints information about
   memberships for a specific group.
`
}

func (*groupListMembershipsCommand) Usage() string {
	return "tharsis [global options] group list-memberships [options] <group-id>"
}

func (*groupListMembershipsCommand) Example() string {
	return `
tharsis group list-memberships trn:group:<group_path>
`
}

func (c *groupListMembershipsCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}

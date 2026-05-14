package command

import (
	"errors"
	"strings"

	"github.com/aws/smithy-go/ptr"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// groupCreateCommand is the top-level structure for the group create command.
type groupCreateCommand struct {
	*BaseCommand

	parentGroupID *string
	description   *string
	toJSON        *bool
	ifNotExists   *bool
	parents       *bool
}

var _ Command = (*groupCreateCommand)(nil)

func (c *groupCreateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: name")
	}

	if *c.parents && c.parentGroupID == nil {
		return errors.New("-parent-group-id is required when -parents is set")
	}

	return nil
}

// NewGroupCreateCommandFactory returns a groupCreateCommand struct.
func NewGroupCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupCreateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	name := c.arguments[0]

	if *c.ifNotExists {
		var checkID string
		if c.parentGroupID != nil {
			c.Logger.Debug("getting parent group", "value", *c.parentGroupID)

			group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: *c.parentGroupID})
			if err != nil {
				// If parent not found and -parents is set, the target can't
				// exist either — skip the check and proceed to create parents.
				if !(*c.parents && status.Code(err) == codes.NotFound) {
					c.UI.ErrorWithSummary(err, "failed to get parent group")
					return 1
				}
			} else {
				checkID = trn.TypeGroup.Build(group.FullPath, name)
			}
		} else {
			checkID = trn.TypeGroup.Build(name)
		}

		if checkID != "" {
			c.Logger.Debug("checking if group exists", "value", checkID)

			existingGroup, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: checkID})
			if err != nil && status.Code(err) != codes.NotFound {
				c.UI.ErrorWithSummary(err, "failed to check group")
				return 1
			}

			if existingGroup != nil {
				c.Logger.Debug("group already exists, returning existing group")
				return c.Output(existingGroup, c.toJSON)
			}
		}
	}

	// When -parents is set, ensure all groups in the -parent-group-id path
	// exist before creating the target group. The -parent-group-id must be
	// a TRN so the path can be extracted without requiring the group to exist.
	if *c.parents {
		if !trn.IsTRN(*c.parentGroupID) {
			c.UI.ErrorWithSummary(
				errors.New("-parent-group-id must be a TRN when -parents is set"),
				"failed to validate group create input",
			)
			return 1
		}

		parsed, err := trn.TypeGroup.Parse(*c.parentGroupID)
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to parse parent group ID")
			return 1
		}

		// Ensure parent groups exist.
		if code := c.ensureParentGroups(parsed.Path()); code != 0 {
			return code
		}
	}

	if c.parentGroupID == nil {
		// Deprecated. Attempt to extract parent path from the argument.
		if parent, child := extractParentPath(name); parent != "" {
			c.parentGroupID = new(trn.TypeGroup.Build(parent))
			name = child
		}
	}

	input := &pb.CreateGroupRequest{
		Name:        name,
		ParentId:    c.parentGroupID,
		Description: ptr.ToString(c.description),
	}

	createdGroup, err := c.grpcClient.GroupsClient.CreateGroup(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create a group")
		return 1
	}

	return c.Output(createdGroup, c.toJSON)
}

// ensureParentGroups creates any missing groups in the given path.
// Existing groups are skipped.
func (c *groupCreateCommand) ensureParentGroups(path string) int {
	segments := strings.Split(path, "/")

	var lastGroup *pb.Group
	for i := range segments {
		currentPath := strings.Join(segments[:i+1], "/")
		currentTRN := trn.TypeGroup.Build(currentPath)

		c.Logger.Debug("checking if parent group exists", "path", currentPath)

		existing, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: currentTRN})
		if err != nil && status.Code(err) != codes.NotFound {
			c.UI.ErrorWithSummary(err, "failed to check group")
			return 1
		}

		if existing != nil {
			lastGroup = existing
			continue
		}

		// Create the missing parent group.
		var parentID *string
		if lastGroup != nil {
			parentTRN := trn.TypeGroup.Build(lastGroup.FullPath)
			parentID = &parentTRN
		}

		c.Logger.Debug("creating parent group", "name", segments[i], "parentID", parentID)

		created, err := c.grpcClient.GroupsClient.CreateGroup(c.Context, &pb.CreateGroupRequest{
			Name:     segments[i],
			ParentId: parentID,
		})
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to create a group")
			return 1
		}

		lastGroup = created
	}

	return 0
}

func (*groupCreateCommand) Synopsis() string {
	return "Create a new group."
}

func (*groupCreateCommand) Usage() string {
	return "tharsis [global options] group create [options] <name>"
}

func (*groupCreateCommand) Description() string {
	return `
   Creates a new group under a parent group with an
   optional description. Use -parents to automatically
   create any missing intermediate groups specified
   by -parent-group-id.
`
}

func (*groupCreateCommand) Example() string {
	return `
tharsis group create \
  -parent-group-id "trn:group:<group_path>" \
  -description "Operations group" \
  <name>

tharsis group create -parents \
  -parent-group-id "trn:group:xyz/team-a/dev" \
  api
`
}

func (c *groupCreateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.parentGroupID,
		"parent-group-id",
		"Parent group ID.",
	)
	f.StringVar(
		&c.description,
		"description",
		"Description for the new group.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)
	f.BoolVar(
		&c.ifNotExists,
		"if-not-exists",
		"Do not error if the group already exists; return the existing group instead.",
		flag.Default(false),
	)
	f.BoolVar(
		&c.parents,
		"parents",
		"Create missing intermediate groups specified by -parent-group-id.",
		flag.Default(false),
	)

	return f
}

package command

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// workspaceCreateCommand is the top-level structure for the workspace create command.
type workspaceCreateCommand struct {
	*BaseCommand

	parentGroupID      string
	description        string
	terraformVersion   string
	maxJobDuration     *int32
	preventDestroyPlan bool
	managedIdentityID  *string
	labels             map[string]string
	toJSON             bool
	ifNotExists        bool
}

var _ Command = (*workspaceCreateCommand)(nil)

func (c *workspaceCreateCommand) validate() error {
	const message = "name is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewWorkspaceCreateCommandFactory returns a workspaceCreateCommand struct.
func NewWorkspaceCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceCreateCommand{
			BaseCommand: baseCommand,
			labels:      make(map[string]string),
		}, nil
	}
}

func (c *workspaceCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	name := c.arguments[0]

	if c.parentGroupID == "" {
		// Handle deprecated syntax where full path of the workspace is passed into the argument.
		parent, child := extractParentPath(name)
		c.parentGroupID = trn.NewResourceTRN(trn.ResourceTypeGroup, parent)
		name = child
	}

	if c.ifNotExists {
		c.Logger.Debug("getting parent group", "value", c.parentGroupID)

		group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: c.parentGroupID})
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to get parent group")
			return 1
		}

		checkID := trn.NewResourceTRN(trn.ResourceTypeWorkspace, group.FullPath, name)
		c.Logger.Debug("checking if workspace exists", "value", checkID)

		workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{Id: checkID})
		if err != nil && status.Code(err) != codes.NotFound {
			c.UI.ErrorWithSummary(err, "failed to check workspace")
			return 1
		}

		if workspace != nil {
			c.Logger.Debug("workspace already exists, returning existing workspace")
			return outputWorkspace(c.UI, c.toJSON, workspace)
		}
	}

	input := &pb.CreateWorkspaceRequest{
		Name:               name,
		GroupId:            c.parentGroupID,
		Description:        c.description,
		TerraformVersion:   c.terraformVersion,
		MaxJobDuration:     c.maxJobDuration,
		PreventDestroyPlan: c.preventDestroyPlan,
		Labels:             c.labels,
	}

	createdWorkspace, err := c.grpcClient.WorkspacesClient.CreateWorkspace(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create a workspace")
		return 1
	}

	if c.managedIdentityID != nil {
		assignInput := &pb.AssignManagedIdentityToWorkspaceRequest{
			ManagedIdentityId: toTRN(trn.ResourceTypeManagedIdentity, *c.managedIdentityID),
			WorkspaceId:       createdWorkspace.Metadata.Id,
		}

		if _, err := c.grpcClient.ManagedIdentitiesClient.AssignManagedIdentityToWorkspace(c.Context, assignInput); err != nil {
			c.UI.ErrorWithSummary(err, "failed to assign managed identity to workspace")
			return 1
		}
	}

	return outputWorkspace(c.UI, c.toJSON, createdWorkspace)
}

func (*workspaceCreateCommand) Synopsis() string {
	return "Create a new workspace."
}

func (*workspaceCreateCommand) Usage() string {
	return "tharsis [global options] workspace create [options] <name>"
}

func (*workspaceCreateCommand) Description() string {
	return `
   The workspace create command creates a new workspace. It
   allows setting a workspace's description (optional),
   maximum job duration and managed identity. Shows final
   output as JSON, if specified. Idempotent when used with
   --if-not-exists option.
`
}

func (*workspaceCreateCommand) Example() string {
	return `
tharsis workspace create \
  --parent-group-id trn:group:<group_path> \
  --description "Production workspace" \
  --terraform-version "1.5.0" \
  --max-job-duration 60 \
  --prevent-destroy-plan \
  --managed-identity trn:managed_identity:<group_path>/<identity_name> \
  --label env=prod \
  --label team=platform \
  <name>
`
}

func (c *workspaceCreateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.parentGroupID,
		"parent-group-id",
		"",
		"Parent group ID.",
	)
	f.StringVar(
		&c.description,
		"description",
		"",
		"Description for the new workspace.",
	)
	f.StringVar(
		&c.terraformVersion,
		"terraform-version",
		"",
		"The default Terraform CLI version for the new workspace.",
	)
	f.Func(
		"max-job-duration",
		"The amount of minutes before a job is gracefully canceled (Default 720).",
		func(s string) error {
			v, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return err
			}
			c.maxJobDuration = ptr.Int32(int32(v))
			return nil
		},
	)
	f.BoolVar(
		&c.preventDestroyPlan,
		"prevent-destroy-plan",
		false,
		"Whether a run/plan will be prevented from destroying deployed resources.",
	)
	f.Func(
		"label",
		"Labels for the new workspace (key=value). Can be specified multiple times.",
		func(s string) error {
			parts := strings.Split(s, "=")
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("label key and value cannot be empty")
			}
			c.labels[parts[0]] = parts[1]
			return nil
		},
	)
	f.Func(
		"managed-identity",
		"The ID of a managed identity to assign.",
		func(s string) error {
			c.managedIdentityID = &s
			return nil
		},
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)
	f.BoolVar(
		&c.ifNotExists,
		"if-not-exists",
		false,
		"Create a workspace if it does not already exist.",
	)

	return f
}

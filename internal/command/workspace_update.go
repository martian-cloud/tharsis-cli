package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// workspaceUpdateCommand is the top-level structure for the workspace update command.
type workspaceUpdateCommand struct {
	*BaseCommand

	description        *string
	terraformVersion   *string
	maxJobDuration     *int32
	version            *int64
	labels             map[string]string
	preventDestroyPlan *bool
	toJSON             *bool
}

var _ Command = (*workspaceUpdateCommand)(nil)

func (c *workspaceUpdateCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewWorkspaceUpdateCommandFactory returns a workspaceUpdateCommand struct.
func NewWorkspaceUpdateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceUpdateCommand{
			BaseCommand: baseCommand,
			labels:      make(map[string]string),
		}, nil
	}
}

func (c *workspaceUpdateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace update"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.UpdateWorkspaceRequest{
		Id:                 trn.ToTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
		Description:        c.description,
		TerraformVersion:   c.terraformVersion,
		MaxJobDuration:     c.maxJobDuration,
		PreventDestroyPlan: c.preventDestroyPlan,
		Version:            c.version,
		Labels:             c.labels,
	}

	updatedWorkspace, err := c.grpcClient.WorkspacesClient.UpdateWorkspace(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update a workspace")
		return 1
	}

	return outputWorkspace(c.UI, *c.toJSON, updatedWorkspace)
}

func (*workspaceUpdateCommand) Synopsis() string {
	return "Update a workspace."
}

func (*workspaceUpdateCommand) Usage() string {
	return "tharsis [global options] workspace update [options] <id>"
}

func (*workspaceUpdateCommand) Description() string {
	return `
   The workspace update command updates a workspace.
   Currently, it supports updating the description and the
   maximum job duration. Shows final output as JSON, if
   specified.
`
}

func (*workspaceUpdateCommand) Example() string {
	return `
tharsis workspace update \
  -description "Updated production workspace" \
  -terraform-version "1.6.0" \
  -max-job-duration 120 \
  -prevent-destroy-plan true \
  trn:workspace:<workspace_path>
`
}

func (c *workspaceUpdateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.description,
		"description",
		"Description for the workspace.",
	)
	f.StringVar(
		&c.terraformVersion,
		"terraform-version",
		"The default Terraform CLI version for the workspace.",
	)
	f.Int32Var(
		&c.maxJobDuration,
		"max-job-duration",
		"The amount of minutes before a job is gracefully canceled.",
	)
	f.BoolVar(
		&c.preventDestroyPlan,
		"prevent-destroy-plan",
		"Whether a run/plan will be prevented from destroying deployed resources.",
	)
	f.Int64Var(
		&c.version,
		"version",
		"Metadata version of the resource to be updated. In most cases, this is not required.",
	)
	f.MapVar(
		&c.labels,
		"label",
		"Labels for the workspace (key=value).",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}

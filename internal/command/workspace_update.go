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
)

// workspaceUpdateCommand is the top-level structure for the workspace update command.
type workspaceUpdateCommand struct {
	*BaseCommand

	description        *string
	terraformVersion   *string
	maxJobDuration     *int32
	preventDestroyPlan *bool
	version            *int64
	labels             map[string]string
	toJSON             bool
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
		Id:                 toTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
		Description:        c.description,
		TerraformVersion:   c.terraformVersion,
		MaxJobDuration:     c.maxJobDuration,
		PreventDestroyPlan: c.preventDestroyPlan,
		Version:            c.version,
		Labels:             c.labels,
	}

	c.Logger.Debug("workspace update input", "input", input)

	updatedWorkspace, err := c.grpcClient.WorkspacesClient.UpdateWorkspace(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update a workspace")
		return 1
	}

	return outputWorkspace(c.UI, c.toJSON, updatedWorkspace)
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
  --description "Updated production workspace" \
  --terraform-version "1.6.0" \
  --max-job-duration 120 \
  --prevent-destroy-plan true \
  trn:workspace:<workspace_path>
`
}

func (c *workspaceUpdateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"description",
		"Description for the workspace.",
		func(s string) error {
			c.description = &s
			return nil
		},
	)
	f.Func(
		"terraform-version",
		"The default Terraform CLI version for the workspace.",
		func(s string) error {
			c.terraformVersion = &s
			return nil
		},
	)
	f.Func(
		"max-job-duration",
		"The amount of minutes before a job is gracefully canceled.",
		func(s string) error {
			v, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return err
			}
			c.maxJobDuration = ptr.Int32(int32(v))
			return nil
		},
	)
	f.Func(
		"prevent-destroy-plan",
		"Whether a run/plan will be prevented from destroying deployed resources.",
		func(s string) error {
			v, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}
			c.preventDestroyPlan = &v
			return nil
		},
	)
	f.Func(
		"version",
		"Metadata version of the resource to be updated. "+
			"In most cases, this is not required.",
		func(s string) error {
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			c.version = &v
			return nil
		},
	)
	f.Func(
		"label",
		"Labels for the workspace (key=value). Can be specified multiple times.",
		func(s string) error {
			parts := strings.Split(s, "=")
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("label key and value cannot be empty")
			}
			c.labels[parts[0]] = parts[1]
			return nil
		},
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}

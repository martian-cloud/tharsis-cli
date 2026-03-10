package command

import (
	"flag"
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type workspaceLabelCommand struct {
	*BaseCommand

	labels    map[string]*string // nil value means remove
	overwrite bool
	toJSON    bool
}

// NewWorkspaceLabelCommandFactory returns a workspaceLabelCommand struct.
func NewWorkspaceLabelCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceLabelCommand{
			BaseCommand: baseCommand,
			labels:      make(map[string]*string),
		}, nil
	}
}

func (c *workspaceLabelCommand) validate() error {
	const message = "workspace-id is required"

	for key, value := range c.labels {
		if key == "" {
			return fmt.Errorf("invalid label format: key cannot be empty")
		}
		if value != nil && *value == "" {
			return fmt.Errorf("invalid label format: value cannot be empty for key %s", key)
		}
	}

	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *workspaceLabelCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace label"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	workspaceID := c.arguments[0]

	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{
		Id: workspaceID,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
		return 1
	}

	newLabels := c.applyLabelOperations(workspace.Labels)

	input := &pb.UpdateWorkspaceRequest{
		Id:     workspaceID,
		Labels: newLabels,
	}

	c.Logger.Debug("workspace label input", "input", input)

	updatedWorkspace, err := c.grpcClient.WorkspacesClient.UpdateWorkspace(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update workspace")
		return 1
	}

	return outputWorkspace(c.UI, c.toJSON, updatedWorkspace)
}

func (c *workspaceLabelCommand) applyLabelOperations(existingLabels map[string]string) map[string]string {
	var workingLabels map[string]string
	if c.overwrite {
		workingLabels = make(map[string]string)
	} else {
		workingLabels = make(map[string]string, len(existingLabels))
		for k, v := range existingLabels {
			workingLabels[k] = v
		}
	}

	for key, value := range c.labels {
		if value == nil {
			delete(workingLabels, key)
		} else {
			workingLabels[key] = *value
		}
	}

	return workingLabels
}

func (*workspaceLabelCommand) Synopsis() string {
	return "Manage labels on a workspace."
}

func (*workspaceLabelCommand) Description() string {
	return `
   The workspace label command manages labels on a workspace.
   It supports adding, updating, removing, and overwriting labels.

   Label operations:
     key=value  Add or update a label
     key-       Remove a label (not allowed with --overwrite)
`
}

func (*workspaceLabelCommand) Usage() string {
	return "tharsis [global options] workspace label [options] <workspace-id>"
}

func (*workspaceLabelCommand) Example() string {
	return `
tharsis workspace label \
  --label env=prod \
  --label tier=frontend \
  trn:workspace:ops/my-workspace
`
}

func (c *workspaceLabelCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"label",
		"Label operation (key=value to add/update, key- to remove). Can be specified multiple times.",
		func(s string) error {
			if key, ok := strings.CutSuffix(s, "-"); ok {
				c.labels[key] = nil
			} else if strings.Contains(s, "=") {
				parts := strings.SplitN(s, "=", 2)
				c.labels[parts[0]] = &parts[1]
			} else {
				return fmt.Errorf("invalid label format: %s (expected key=value or key-)", s)
			}

			return nil
		},
	)
	f.BoolVar(
		&c.overwrite,
		"overwrite",
		false,
		"Replace all existing labels with the specified labels.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)

	return f
}

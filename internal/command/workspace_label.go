package command

import (
	"flag"
	"fmt"
	"maps"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
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
	if len(c.arguments) < 2 {
		return fmt.Errorf("workspace-id and at least one label operation are required")
	}

	// Validate label operations
	for _, arg := range c.arguments[1:] {
		if key, ok := strings.CutSuffix(arg, "-"); ok {
			if key == "" {
				return fmt.Errorf("invalid label format: key cannot be empty")
			}
			if c.overwrite {
				return fmt.Errorf("label removal syntax (key-) cannot be used with --overwrite flag")
			}
		} else if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			if parts[0] == "" {
				return fmt.Errorf("invalid label format: key cannot be empty")
			}
			if parts[1] == "" {
				return fmt.Errorf("invalid label format: value cannot be empty for key %s", parts[0])
			}
		} else {
			return fmt.Errorf("invalid label format: %s (expected key=value or key-)", arg)
		}
	}

	return nil
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

	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{
		Id: trn.ToTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
		return 1
	}

	// Parse label operations from remaining arguments
	for _, arg := range c.arguments[1:] {
		if key, ok := strings.CutSuffix(arg, "-"); ok {
			c.labels[key] = nil
		} else {
			parts := strings.SplitN(arg, "=", 2)
			c.labels[parts[0]] = &parts[1]
		}
	}

	newLabels := c.applyLabelOperations(workspace.Labels)

	input := &pb.UpdateWorkspaceRequest{
		Id:     workspace.Metadata.Id,
		Labels: newLabels,
	}

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
		maps.Copy(workingLabels, existingLabels)
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
	return "tharsis [global options] workspace label [options] <workspace-id> <label-operation>..."
}

func (*workspaceLabelCommand) Example() string {
	return `
tharsis workspace label \
  --overwrite \
  trn:workspace:<workspace_path> \
  env=prod \
  tier=frontend
`
}

func (c *workspaceLabelCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
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

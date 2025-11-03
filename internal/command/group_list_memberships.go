package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// groupListMembershipsCommand is the top-level structure for the group list-memberships command.
type groupListMembershipsCommand struct {
	meta *Metadata
}

// NewGroupListMembershipsCommandFactory returns a groupListMembershipsCommand struct.
func NewGroupListMembershipsCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupListMembershipsCommand{
			meta: meta,
		}, nil
	}
}

func (glm groupListMembershipsCommand) Run(args []string) int {
	glm.meta.Logger.Debugf("Starting the 'group list-memberships' command with %d arguments:", len(args))
	for ix, arg := range args {
		glm.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := glm.meta.GetSDKClient()
	if err != nil {
		glm.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return glm.doGroupListMemberships(ctx, client, args)
}

func (glm groupListMembershipsCommand) doGroupListMemberships(ctx context.Context, client *tharsis.Client, opts []string) int {
	glm.meta.Logger.Debugf("will do group list-memberships, %d opts", len(opts))

	defs := glm.buildGroupListMembershipsOptionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(glm.meta.BinaryName+" group list-memberships", defs, opts)
	if err != nil {
		glm.meta.Logger.Error(output.FormatError("failed to parse group list-memberships argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		glm.meta.Logger.Error(output.FormatError("missing group list-memberships full path", nil), glm.HelpGroupListMemberships())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group list-memberships arguments: %s", cmdArgs)
		glm.meta.Logger.Error(output.FormatError(msg, nil), glm.HelpGroupListMemberships())
		return 1
	}

	path := cmdArgs[0]

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		glm.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Error is already logged.
	if !isNamespacePathValid(glm.meta, path) {
		return 1
	}

	// Query for the group to make sure it exists and is a group.
	_, err = client.Group.GetGroup(ctx, &sdktypes.GetGroupInput{Path: &path})
	if err != nil {
		glm.meta.UI.Error(output.FormatError("failed to find group", err))
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.GetNamespaceMembershipsInput{
		NamespacePath: path,
	}
	glm.meta.Logger.Debugf("group list-memberships input: %#v", input)

	// Get the group's memberships.
	foundMemberships, err := client.NamespaceMembership.GetMemberships(ctx, input)
	if err != nil {
		glm.meta.Logger.Error(output.FormatError("failed to list a group's memberships", err))
		return 1
	}

	return outputNamespaceMemberships(glm.meta, toJSON, foundMemberships)
}

// buildGroupListMembershipsOptionDefs returns the defs used by
// group list-memberships command.
func (glm groupListMembershipsCommand) buildGroupListMembershipsOptionDefs() optparser.OptionDefinitions {
	return buildJSONOptionDefs(optparser.OptionDefinitions{})
}

func (glm groupListMembershipsCommand) Synopsis() string {
	return "List a group's memberships."
}

func (glm groupListMembershipsCommand) Help() string {
	return glm.HelpGroupListMemberships()
}

// HelpGroupListMemberships prints the help string for the 'group list-memberships' command.
func (glm groupListMembershipsCommand) HelpGroupListMemberships() string {
	return fmt.Sprintf(`
Usage: %s [global options] group list-memberships <full_path>

   The group list-memberships command lists a group's memberships.

%s

`, glm.meta.BinaryName, buildHelpText(glm.buildGroupListMembershipsOptionDefs()))
}

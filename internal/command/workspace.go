package command

import (
	"fmt"

	"github.com/mitchellh/cli"
)

// workspaceCommand is the top-level structure for the workspace command.
type workspaceCommand struct {
	meta *Metadata
}

// NewWorkspaceCommandFactory returns a workspaceCommand struct.
func NewWorkspaceCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceCommand{
			meta: meta,
		}, nil
	}
}

func (wc workspaceCommand) Run(args []string) int {
	wc.meta.Logger.Debugf("Starting the 'workspace' command with %d arguments:", len(args))
	for ix, arg := range args {
		wc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Show the help text.
	wc.meta.UI.Output(wc.HelpWorkspace(true))
	return 1
}

func (wc workspaceCommand) Synopsis() string {
	return "Do operations on workspaces."
}

func (wc workspaceCommand) Help() string {
	return wc.HelpWorkspace(false)
}

// HelpWorkspace produces the help string for the 'workspace' command.
func (wc workspaceCommand) HelpWorkspace(subCommands bool) string {
	usage := fmt.Sprintf(`
Usage: %s [global options] workspace ...

   The workspace commands do operations on workspaces.
   Subcommands allow designating / revoking managed
   identities to / from a workspace, creating /
   updating / deleting workspaces, setting
   Terraform / environment variables and more.
`, wc.meta.BinaryName)
	sc := `

Subcommands:
    assign-managed-identity            Assign a managed identity to a workspace.
    create                             Create a new workspace.
    delete                             Delete a workspace.
    get                                Get a single workspace.
    get-assigned-managed-identities    Get assigned managed identities for a workspace.
    list                               List workspaces.
    outputs                            Get the state version outputs for a workspace.
    set-environment-vars               Set environment variables for a workspace.
    set-terraform-vars                 Set terraform variables for a workspace.
    unassign-managed-identity          Unassign a managed identity from a workspace.
    update                             Update a workspace.`

	// Avoid duplicate subcommands when -h is used.
	if subCommands {
		return usage + sc
	}

	return usage
}

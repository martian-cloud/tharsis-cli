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

	return cli.RunResultHelp
}

func (wc workspaceCommand) Synopsis() string {
	return "Do operations on workspaces."
}

func (wc workspaceCommand) Help() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace ...

   The workspace commands do operations on workspaces.
   Subcommands allow designating / revoking managed
   identities to / from a workspace, creating /
   updating / deleting workspaces, setting
   Terraform / environment variables and more.
`, wc.meta.BinaryName)
}

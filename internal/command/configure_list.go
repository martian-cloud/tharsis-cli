package command

import (
	"fmt"
	"sort"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tableformatter"
)

// configureListCommand is the top-level structure for the configure list command.
type configureListCommand struct {
	meta *Metadata
}

// NewConfigureListCommandFactory returns a configureListCommand struct.
func NewConfigureListCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return configureListCommand{
			meta: meta,
		}, nil
	}
}

func (clc configureListCommand) Run(args []string) int {
	clc.meta.Logger.Debugf("Starting the 'configure list' command with %d arguments:", len(args))
	for ix, arg := range args {
		clc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Attempt to read the existing settings.
	gotSettings, err := settings.ReadSettings(nil)
	if err != nil {
		clc.meta.UI.Output(output.FormatError("Failed to read pre-existing settings", err))
		return 1
	}

	return clc.showAllProfiles(gotSettings)
}

func (clc configureListCommand) showAllProfiles(settings *settings.Settings) int {

	// First sort the profile names.
	profileNames := []string{}
	for profileName := range settings.Profiles {
		profileNames = append(profileNames, profileName)
	}
	sort.Strings(profileNames)

	// Format and print the output.
	tableInput := make([][]string, len(settings.Profiles)+1)
	tableInput[0] = []string{"Profile", "URL"}
	for ix, profileName := range profileNames {
		tableInput[ix+1] = []string{profileName, settings.Profiles[profileName].TharsisURL}
	}
	clc.meta.UI.Output(tableformatter.FormatTable(tableInput))

	return 0
}

func (clc configureListCommand) Synopsis() string {
	return "Show all profiles"
}

func (clc configureListCommand) Help() string {
	return clc.HelpConfigureList()
}

// HelpConfigureList produces the help string for the 'configure list' command.
func (clc configureListCommand) HelpConfigureList() string {
	return fmt.Sprintf(`
Usage: %s configure list

   The configure list command prints information about all profiles.

`, clc.meta.BinaryName)
}

// The End.

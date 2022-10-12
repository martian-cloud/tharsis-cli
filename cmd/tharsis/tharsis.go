package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/command"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/logger"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
)

const (
	// Logger level environment variable
	logLevelEnvVar = "THARSIS_CLI_LOG"

	// Profile (-p) global option
	profileOption = "p"

	// Short versions of help and version options
	helpOptionShort    = "h"
	versionOptionShort = "v"
)

var (
	// Binary name and display title
	binaryName   string
	displayTitle string

	// Global option names and how many value arguments each requires.
	// Will be properly initialized shortly.
	globalOptionNames = optparser.OptionDefinitions{}

	// Version is passed in via ldflags at build time
	Version = "1.0.0"

	// DefaultEndpointURL is passed in via ldflags at build time.
	DefaultEndpointURL string
)

func main() {
	os.Exit(realMain())
}

// Facilitate testing the main function by wrapping it.
// Now, a test can call realMain without having the os.Exit call getting in the way.
func realMain() int {

	// Binary name and raw arguments.
	binaryName = filepath.Base(os.Args[0])
	displayTitle = capitalizeFirst(binaryName)

	// Build the global options.
	// The CLI library automatically handles short and long flavors of help and version,
	// but the flag parser requires every flavor of every option to be defined up front.
	globalOptionNames = optparser.OptionDefinitions{
		helpOptionShort: {
			Synopsis: "Show this usage message.",
		},
		profileOption: {
			Arguments: []string{"PROFILE"},
			Synopsis:  "profile name from config file, defaults to \"default\"",
		},
		versionOptionShort: {
			Synopsis: fmt.Sprintf("show the %s version information", displayTitle),
		},
	}

	// Binary name and raw arguments.
	rawArgs := os.Args[1:]

	// Log the startup.
	log := logger.NewAtLevel(os.Getenv(logLevelEnvVar))
	log.Debugf("Tharsis CLI version %s", Version)
	log.Debugf("binary name: %s", binaryName)
	log.Debugf("display title: %s", displayTitle)
	log.Debugf("raw arguments: %#v", rawArgs)

	// Read any global options.
	globalOptions, commandArgs, err := optparser.ParseCommandOptions(binaryName+" global options", globalOptionNames, rawArgs)
	if err != nil {
		// Logging the error here produces redundant output since flag library already does it.
		return 1
	}

	// Do some manual fix-up to feed -h and -v back to the CLI library for it to handle.
	commandArgs = fixUpHelpVersionOptions(commandArgs, globalOptions)

	log.Debugf("global options: %#v", globalOptions)
	log.Debugf("   commandArgs: %#v", commandArgs)

	// Prepare the command metadata struct.
	meta := &command.Metadata{
		BinaryName:   binaryName,
		DisplayTitle: displayTitle,
		Version:      Version,
		Logger:       log,
		UI: &cli.BasicUi{ // text UI for the CLI platform
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
		// CurrentProfileName will be set later in this module.
		// Settings will be set in individual command modules.
		DefaultEndpointURL: DefaultEndpointURL,
	}

	// Commands.
	commands := map[string]cli.CommandFactory{
		// There is no explicit 'help' command, only the '-help' global option.
		"apply":                               command.NewApplyCommandFactory(meta),
		"configure":                           command.NewConfigureCommandFactory(meta),
		"configure list":                      command.NewConfigureListCommandFactory(meta),
		"group":                               command.NewGroupCommandFactory(meta),
		"group set-terraform-vars":            command.NewGroupSetTerraformVarsCommandFactory(meta),
		"group set-environment-vars":          command.NewGroupSetEnvironmentVarsCommandFactory(meta),
		"group list":                          command.NewGroupListCommandFactory(meta),
		"group get":                           command.NewGroupGetCommandFactory(meta),
		"group create":                        command.NewGroupCreateCommandFactory(meta),
		"group update":                        command.NewGroupUpdateCommandFactory(meta),
		"group delete":                        command.NewGroupDeleteCommandFactory(meta),
		"plan":                                command.NewPlanCommandFactory(meta),
		"destroy":                             command.NewDestroyCommandFactory(meta),
		"provider":                            command.NewProviderCommandFactory(meta),
		"provider create":                     command.NewProviderCreateCommandFactory(meta),
		"provider upload-version":             command.NewProviderUploadVersionCommandFactory(meta),
		"run":                                 command.NewRunCommandFactory(meta),
		"run cancel":                          command.NewRunCancelCommandFactory(meta),
		"sso":                                 command.NewSSOCommandFactory(meta),
		"sso login":                           command.NewLoginCommandFactory(meta),
		"workspace":                           command.NewWorkspaceCommandFactory(meta),
		"workspace assign-managed-identity":   command.NewWorkspaceAssignManagedIdentityCommandFactory(meta),
		"workspace unassign-managed-identity": command.NewWorkspaceUnassignManagedIdentityCommandFactory(meta),
		"workspace get-assigned-managed-identities": command.NewWorkspaceGetAssignedManagedIdentitiesCommandFactory(meta),
		"workspace set-terraform-vars":              command.NewWorkspaceSetTerraformVarsCommandFactory(meta),
		"workspace set-environment-vars":            command.NewWorkspaceSetEnvironmentVarsCommandFactory(meta),
		"workspace list":                            command.NewWorkspaceListCommandFactory(meta),
		"workspace get":                             command.NewWorkspaceGetCommandFactory(meta),
		"workspace create":                          command.NewWorkspaceCreateCommandFactory(meta),
		"workspace update":                          command.NewWorkspaceUpdateCommandFactory(meta),
		"workspace delete":                          command.NewWorkspaceDeleteCommandFactory(meta),
		"workspace outputs":                         command.NewWorkspaceOutputsCommandFactory(meta),
	}

	// Set the profile name.
	profileNames, ok := globalOptions[profileOption]
	var profileName string
	if ok {
		profileName = profileNames[0] // already know a value was processed
	} else {
		profileName = settings.DefaultProfileName
	}
	log.Debugf("profile name: %s", profileName)
	meta.CurrentProfileName = profileName

	// Prepare to launch the CLI platform.
	c := cli.CLI{
		Name:       binaryName,
		Version:    Version,
		Args:       commandArgs,
		Commands:   commands,
		HelpFunc:   helpFunc,
		HelpWriter: os.Stdout,
	}

	// Launch the CLI platform.
	exitStatus, err := c.Run()
	if err != nil {
		log.Error(err.Error())
		return 1
	}

	log.Debugf("exit status: %d", exitStatus)
	return exitStatus
}

func capitalizeFirst(arg string) string {
	if arg == "" {
		return arg // for pathological input, return pathological output
	}
	return strings.ToUpper(arg[0:1]) + arg[1:]
}

// Manual fix-up to put -h and -v options back into command arguments so the CLI library can handle them.
func fixUpHelpVersionOptions(commandArgs []string, globalOptions map[string][]string) []string {

	// First, capture the -h and -v options to the result slice.
	hasDashH := false
	hasDashV := false
	if _, ok := globalOptions[helpOptionShort]; ok {
		hasDashH = true
	}
	if _, ok := globalOptions[versionOptionShort]; ok {
		hasDashV = true
	}

	// Start building the result.
	result := []string{}
	if hasDashH {
		result = append(result, "-h")
	}
	if hasDashV {
		result = append(result, "-v")
	}

	// Last, copy the other command args.
	result = append(result, commandArgs...)

	return result
}

// The End.

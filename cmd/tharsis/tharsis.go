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
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
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
			Synopsis:  "Profile name from config file, defaults to \"default\".",
		},
		versionOptionShort: {
			Synopsis: fmt.Sprintf("Show the %s version information.", displayTitle),
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

	// For any variation of "-h" or "-help", simply use "-h".
	// Since help option can be used for any command, we must
	// handle it the same anywhere.
	for ix, arg := range rawArgs {
		if arg == "--h" || arg == "--help" || arg == "-help" {
			rawArgs[ix] = "-h"
		}
	}

	// Only replace "--version" and "--v" at the global level i.e. the first argument.
	// Allows using the same argument in commands and subcommands.
	if len(rawArgs) > 0 && (rawArgs[0] == "--version" || rawArgs[0] == "-version" || rawArgs[0] == "--v") {
		rawArgs[0] = "-v"
	}

	// Read any global options.
	globalOptions, commandArgs, err := optparser.ParseCommandOptions(binaryName+" global options", globalOptionNames, rawArgs)
	if err != nil {
		log.Info(output.FormatError("failed to parse global options", err))
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
		UI: &output.UI{
			BasicUI: &cli.BasicUi{
				Reader:      os.Stdin,
				Writer:      os.Stdout,
				ErrorWriter: os.Stderr,
			},
		},
		// CurrentProfileName will be set later in this module.
		// Settings will be set in individual command modules.
		DefaultEndpointURL: DefaultEndpointURL,
	}

	// Commands.
	commands := map[string]cli.CommandFactory{
		// There is no explicit 'help' command, only the '-help' global option.
		"apply":                                     command.NewApplyCommandFactory(meta),
		"configure":                                 command.NewConfigureCommandFactory(meta),
		"configure list":                            command.NewConfigureListCommandFactory(meta),
		"group":                                     command.NewGroupCommandFactory(meta),
		"group set-terraform-vars":                  command.NewGroupSetTerraformVarsCommandFactory(meta),
		"group set-terraform-var":                   command.NewGroupSetTerraformVarCommandFactory(meta),
		"group delete-terraform-var":                command.NewGroupDeleteTerraformVarCommandFactory(meta),
		"group set-environment-vars":                command.NewGroupSetEnvironmentVarsCommandFactory(meta),
		"group list":                                command.NewGroupListCommandFactory(meta),
		"group get":                                 command.NewGroupGetCommandFactory(meta),
		"group create":                              command.NewGroupCreateCommandFactory(meta),
		"group migrate":                             command.NewGroupMigrateCommandFactory(meta),
		"group update":                              command.NewGroupUpdateCommandFactory(meta),
		"group delete":                              command.NewGroupDeleteCommandFactory(meta),
		"group get-membership":                      command.NewGroupGetMembershipCommandFactory(meta),
		"group list-memberships":                    command.NewGroupListMembershipsCommandFactory(meta),
		"group add-membership":                      command.NewGroupAddMembershipCommandFactory(meta),
		"group remove-membership":                   command.NewGroupRemoveMembershipCommandFactory(meta),
		"group update-membership":                   command.NewGroupUpdateMembershipCommandFactory(meta),
		"plan":                                      command.NewPlanCommandFactory(meta),
		"destroy":                                   command.NewDestroyCommandFactory(meta),
		"managed-identity":                          command.NewManagedIdentityCommandFactory(meta),
		"managed-identity create":                   command.NewManagedIdentityCreateCommandFactory(meta),
		"managed-identity delete":                   command.NewManagedIdentityDeleteCommandFactory(meta),
		"managed-identity get":                      command.NewManagedIdentityGetCommandFactory(meta),
		"managed-identity update":                   command.NewManagedIdentityUpdateCommandFactory(meta),
		"managed-identity-access-rule":              command.NewManagedIdentityAccessRuleCommandFactory(meta),
		"managed-identity-access-rule create":       command.NewManagedIdentityAccessRuleCreateCommandFactory(meta),
		"managed-identity-access-rule delete":       command.NewManagedIdentityAccessRuleDeleteCommandFactory(meta),
		"managed-identity-access-rule get":          command.NewManagedIdentityAccessRuleGetCommandFactory(meta),
		"managed-identity-access-rule list":         command.NewManagedIdentityAccessRuleListCommandFactory(meta),
		"managed-identity-access-rule update":       command.NewManagedIdentityAccessRuleUpdateCommandFactory(meta),
		"managed-identity-alias":                    command.NewManagedIdentityAliasCommandFactory(meta),
		"managed-identity-alias create":             command.NewManagedIdentityAliasCreateCommandFactory(meta),
		"managed-identity-alias delete":             command.NewManagedIdentityAliasDeleteCommandFactory(meta),
		"module":                                    command.NewModuleCommandFactory(meta),
		"module get":                                command.NewModuleGetCommandFactory(meta),
		"module list":                               command.NewModuleListCommandFactory(meta),
		"module create":                             command.NewModuleCreateCommandFactory(meta),
		"module update":                             command.NewModuleUpdateCommandFactory(meta),
		"module delete":                             command.NewModuleDeleteCommandFactory(meta),
		"module create-attestation":                 command.NewModuleCreateAttestationCommandFactory(meta),
		"module update-attestation":                 command.NewModuleUpdateAttestationCommandFactory(meta),
		"module delete-attestation":                 command.NewModuleDeleteAttestationCommandFactory(meta),
		"module list-attestations":                  command.NewModuleListAttestationsCommandFactory(meta),
		"module get-version":                        command.NewModuleGetVersionCommandFactory(meta),
		"module list-versions":                      command.NewModuleListVersionsCommandFactory(meta),
		"module delete-version":                     command.NewModuleDeleteVersionCommandFactory(meta),
		"module upload-version":                     command.NewModuleUploadVersionCommandFactory(meta),
		"run":                                       command.NewRunCommandFactory(meta),
		"run cancel":                                command.NewRunCancelCommandFactory(meta),
		"service-account":                           command.NewServiceAccountCommandFactory(meta),
		"service-account create-token":              command.NewServiceAccountCreateTokenCommandFactory(meta),
		"sso":                                       command.NewSSOCommandFactory(meta),
		"sso login":                                 command.NewLoginCommandFactory(meta),
		"terraform-provider":                        command.NewTerraformProviderCommandFactory(meta),
		"terraform-provider create":                 command.NewTerraformProviderCreateCommandFactory(meta),
		"terraform-provider upload-version":         command.NewTerraformProviderUploadVersionCommandFactory(meta),
		"terraform-provider-mirror":                 command.NewTerraformProviderMirrorCommandFactory(meta),
		"terraform-provider-mirror sync":            command.NewTerraformProviderMirrorSyncCommandFactory(meta),
		"terraform-provider-mirror get-version":     command.NewTerraformProviderMirrorGetVersionCommandFactory(meta),
		"terraform-provider-mirror list-versions":   command.NewTerraformProviderMirrorListVersionsCommandFactory(meta),
		"terraform-provider-mirror delete-version":  command.NewTerraformProviderMirrorDeleteVersionCommandFactory(meta),
		"terraform-provider-mirror list-platforms":  command.NewTerraformProviderMirrorListPlatformsCommandFactory(meta),
		"terraform-provider-mirror delete-platform": command.NewTerraformProviderMirrorDeletePlatformCommandFactory(meta),
		"workspace":                                 command.NewWorkspaceCommandFactory(meta),
		"workspace get-membership":                  command.NewWorkspaceGetMembershipCommandFactory(meta),
		"workspace list-memberships":                command.NewWorkspaceListMembershipsCommandFactory(meta),
		"workspace assign-managed-identity":         command.NewWorkspaceAssignManagedIdentityCommandFactory(meta),
		"workspace unassign-managed-identity":       command.NewWorkspaceUnassignManagedIdentityCommandFactory(meta),
		"workspace get-assigned-managed-identities": command.NewWorkspaceGetAssignedManagedIdentitiesCommandFactory(meta),
		"workspace set-terraform-vars":              command.NewWorkspaceSetTerraformVarsCommandFactory(meta),
		"workspace set-terraform-var":               command.NewWorkspaceSetTerraformVarCommandFactory(meta),
		"workspace delete-terraform-var":            command.NewWorkspaceDeleteTerraformVarCommandFactory(meta),
		"workspace set-environment-vars":            command.NewWorkspaceSetEnvironmentVarsCommandFactory(meta),
		"workspace list":                            command.NewWorkspaceListCommandFactory(meta),
		"workspace get":                             command.NewWorkspaceGetCommandFactory(meta),
		"workspace create":                          command.NewWorkspaceCreateCommandFactory(meta),
		"workspace update":                          command.NewWorkspaceUpdateCommandFactory(meta),
		"workspace delete":                          command.NewWorkspaceDeleteCommandFactory(meta),
		"workspace outputs":                         command.NewWorkspaceOutputsCommandFactory(meta),
		"workspace add-membership":                  command.NewWorkspaceAddMembershipCommandFactory(meta),
		"workspace remove-membership":               command.NewWorkspaceRemoveMembershipCommandFactory(meta),
		"workspace update-membership":               command.NewWorkspaceUpdateMembershipCommandFactory(meta),
		"runner-agent":                              command.NewRunnerAgentCommandFactory(meta),
		"runner-agent get":                          command.NewRunnerAgentGetCommandFactory(meta),
		"runner-agent create":                       command.NewRunnerAgentCreateCommandFactory(meta),
		"runner-agent update":                       command.NewRunnerAgentUpdateCommandFactory(meta),
		"runner-agent delete":                       command.NewRunnerAgentDeleteCommandFactory(meta),
		"runner-agent assign-service-account":       command.NewRunnerAgentAssignServiceAccountCommandFactory(meta),
		"runner-agent unassign-service-account":     command.NewRunnerAgentUnassignServiceAccountCommandFactory(meta),
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

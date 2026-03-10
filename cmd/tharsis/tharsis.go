// Package main contains the necessary functions for
// building the help menu and configuring the CLI
// library with all subcommand routes.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mitchellh/cli"
	"github.com/posener/complete"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/command"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/useragent"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	// Logger level environment variable
	logLevelEnvVar = "THARSIS_CLI_LOG"
	// Profile environment variable
	profileEnvVar = "THARSIS_PROFILE"
	// User-Agent header name
	userAgentHeader = "User-Agent"
)

var (
	// Version is passed in via ldflags at build time
	Version = "1.0.0"

	// DefaultHTTPEndpoint is passed in via ldflags at build time.
	DefaultHTTPEndpoint string

	// DefaultTLSSkipVerify is passed in via ldflags at build time.
	// Indicates if client should skip verifying the server's
	// certificate chain and domain name.
	DefaultTLSSkipVerify bool
)

// userAgentTransport wraps an http.RoundTripper to add User-Agent header
type userAgentTransport struct {
	userAgent string
	transport http.RoundTripper
}

// RoundTrip implements http.RoundTripper interface by adding User-Agent header to requests
func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(userAgentHeader, t.userAgent)
	return t.transport.RoundTrip(req)
}

func main() {
	os.Exit(realMain())
}

// Facilitate testing the main function by wrapping it.
// Now, a test can call realMain without having the os.Exit call getting in the way.
func realMain() int {
	var (
		// binaryName is the name of the binary.
		binaryName = filepath.Base(os.Args[0])
		// displayTitle is the name of the binary in title case.
		displayTitle = cases.Title(language.English, cases.Compact).String(binaryName)
		// rawArgs are the arguments passed to the binary.
		rawArgs = os.Args[1:]
		// profileName is the name of the profile to use.
		profileName string

		// Autocomplete flag names.
		autocompleteFlagInstall   = "enable-autocomplete"
		autocompleteFlagUninstall = "disable-autocomplete"
	)

	// Create a global flagSet.
	f := flag.NewFlagSet("global options", flag.ContinueOnError)
	f.SetOutput(io.Discard)

	// Default profile from env var, then fall back to default.
	defaultProfile := os.Getenv(profileEnvVar)
	if defaultProfile == "" {
		defaultProfile = settings.DefaultProfileName
	}

	f.StringVar(
		&profileName,
		"p",
		defaultProfile,
		"Profile name from config file. Overrides THARSIS_PROFILE env var.",
	)
	f.BoolVar(&color.NoColor, "no-color", os.Getenv("NO_COLOR") != "", "Disable colored output. Also respects NO_COLOR env var.")

	// Values are never used since CLI framework can handle them,
	// these are simply meant to facilitate the help output for
	// available global flags.
	_ = f.String("v", "", "Show the version information.")
	_ = f.String("h", "", "Show this usage message.")
	_ = f.Bool(autocompleteFlagInstall, false, "Install shell autocompletion.")
	_ = f.Bool(autocompleteFlagUninstall, false, "Uninstall shell autocompletion.")

	// Check for autocomplete flags - pass directly to CLI without parsing global flags
	isAutocomplete := false
	for _, arg := range rawArgs {
		if arg == "-"+autocompleteFlagInstall || arg == "-"+autocompleteFlagUninstall ||
			arg == "--"+autocompleteFlagInstall || arg == "--"+autocompleteFlagUninstall {
			isAutocomplete = true
			break
		}
	}

	// Only parse global flags if not an autocomplete request
	if !isAutocomplete {
		// Ignore errors since CLI framework will handle them.
		_ = f.Parse(rawArgs)
	}

	logLevel := os.Getenv(logLevelEnvVar)
	if logLevel == "" {
		// if log level is not set, keep it off by default.
		logLevel = "off"
	}

	// Log the startup.
	log := hclog.New(&hclog.LoggerOptions{
		Name:              binaryName,
		Level:             hclog.LevelFromString(logLevel),
		Output:            os.Stderr, // Send logs to stderr
		Color:             hclog.AutoColor,
		DisableTime:       true,
		IndependentLevels: true,
	})

	hclog.SetDefault(log)

	log.Debug("",
		"version", Version,
		"binary_name", binaryName,
		"display_title", displayTitle,
		"arguments", rawArgs,
		"profile_name", profileName,
	)

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

	// Get the CLI context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Build user agent
	userAgent := useragent.BuildUserAgent(Version)

	// Create HTTP client with retry logic
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RequestLogHook = func(_ retryablehttp.Logger, r *http.Request, i int) {
		if i > 0 {
			log.Debug("HTTP request retry", "method", r.Method, "url", r.URL, "attempt", i)
		}
	}
	retryClient.Logger = nil
	retryClient.RetryWaitMin = 10 * time.Second
	retryClient.RetryWaitMax = 60 * time.Second
	httpClient := retryClient.StandardClient()

	// Add User-Agent header to all requests
	httpClient.Transport = &userAgentTransport{
		userAgent: userAgent,
		transport: httpClient.Transport,
	}

	// Prepare the command metadata struct.
	baseCommand := &command.BaseCommand{
		Context:              ctx,
		BinaryName:           binaryName,
		DisplayTitle:         displayTitle,
		Version:              Version,
		Logger:               log,
		UI:                   terminal.ConsoleUI(ctx),
		CurrentProfileName:   profileName,
		DefaultHTTPEndpoint:  DefaultHTTPEndpoint,
		DefaultTLSSkipVerify: DefaultTLSSkipVerify,
		UserAgent:            userAgent,
		HTTPClient:           httpClient,
	}

	// Defer closing the base command, so the UI output could finish rendering.
	defer baseCommand.Close()

	commandArgs := rawArgs // No global flag.
	if f.NFlag() > 0 {
		// A global option was set, so pass remaining arguments to commands,
		// since it will automatically be handled.
		commandArgs = f.Args()
	}

	availableCommands, err := commands(baseCommand)
	if err != nil {
		log.Error(err.Error())
		return 1
	}

	c := cli.CLI{
		Name:                  binaryName,
		Version:               Version,
		Args:                  commandArgs,
		Commands:              availableCommands,
		HelpFunc:              helpFunc(cli.BasicHelpFunc(binaryName), f),
		HelpWriter:            os.Stdout,
		ErrorWriter:           os.Stderr,
		Autocomplete:          true,
		AutocompleteInstall:   autocompleteFlagInstall,
		AutocompleteUninstall: autocompleteFlagUninstall,
		AutocompleteGlobalFlags: complete.Flags{
			"-p": complete.PredictFunc(predictProfiles),
		},
	}

	// Run the CLI.
	exitStatus, err := c.Run()
	if err != nil {
		log.Error(err.Error())
		return 1
	}

	return exitStatus
}

// commands returns all the available commands.
func commands(baseCommand *command.BaseCommand) (map[string]cli.CommandFactory, error) {
	// The map of all commands except documentation.
	commandMap := map[string]command.Factory{
		"apply":                                     command.NewApplyCommandFactory(baseCommand),
		"configure":                                 command.NewConfigureCommandFactory(baseCommand),
		"configure delete":                          command.NewConfigureDeleteCommandFactory(baseCommand),
		"configure list":                            command.NewConfigureListCommandFactory(baseCommand),
		"destroy":                                   command.NewDestroyCommandFactory(baseCommand),
		"group":                                     command.NewHelpCommandFactory(getHelpText("group")),
		"group get":                                 command.NewGroupGetCommandFactory(baseCommand),
		"group create":                              command.NewGroupCreateCommandFactory(baseCommand),
		"group update":                              command.NewGroupUpdateCommandFactory(baseCommand),
		"group delete":                              command.NewGroupDeleteCommandFactory(baseCommand),
		"group list":                                command.NewGroupListCommandFactory(baseCommand),
		"group list-memberships":                    command.NewGroupListMembershipsCommandFactory(baseCommand),
		"group get-terraform-var":                   command.NewGroupGetTerraformVarCommandFactory(baseCommand),
		"group set-terraform-var":                   command.NewGroupSetTerraformVarCommandFactory(baseCommand),
		"group delete-terraform-var":                command.NewGroupDeleteTerraformVarCommandFactory(baseCommand),
		"group list-terraform-vars":                 command.NewGroupListTerraformVarsCommandFactory(baseCommand),
		"group set-terraform-vars":                  command.NewGroupSetTerraformVarsCommandFactory(baseCommand),
		"group list-environment-vars":               command.NewGroupListEnvironmentVarsCommandFactory(baseCommand),
		"group set-environment-vars":                command.NewGroupSetEnvironmentVarsCommandFactory(baseCommand),
		"managed-identity":                          command.NewHelpCommandFactory(getHelpText("managed-identity")),
		"managed-identity get":                      command.NewManagedIdentityGetCommandFactory(baseCommand),
		"managed-identity create":                   command.NewManagedIdentityCreateCommandFactory(baseCommand),
		"managed-identity update":                   command.NewManagedIdentityUpdateCommandFactory(baseCommand),
		"managed-identity delete":                   command.NewManagedIdentityDeleteCommandFactory(baseCommand),
		"managed-identity-access-rule":              command.NewHelpCommandFactory(getHelpText("managed-identity-access-rule")),
		"managed-identity-access-rule get":          command.NewManagedIdentityAccessRuleGetCommandFactory(baseCommand),
		"managed-identity-access-rule list":         command.NewManagedIdentityAccessRuleListCommandFactory(baseCommand),
		"managed-identity-access-rule create":       command.NewManagedIdentityAccessRuleCreateCommandFactory(baseCommand),
		"managed-identity-access-rule update":       command.NewManagedIdentityAccessRuleUpdateCommandFactory(baseCommand),
		"managed-identity-access-rule delete":       command.NewManagedIdentityAccessRuleDeleteCommandFactory(baseCommand),
		"managed-identity-alias":                    command.NewHelpCommandFactory(getHelpText("managed-identity-alias")),
		"managed-identity-alias create":             command.NewManagedIdentityAliasCreateCommandFactory(baseCommand),
		"managed-identity-alias delete":             command.NewManagedIdentityAliasDeleteCommandFactory(baseCommand),
		"module":                                    command.NewHelpCommandFactory(getHelpText("module")),
		"module get":                                command.NewModuleGetCommandFactory(baseCommand),
		"module create":                             command.NewModuleCreateCommandFactory(baseCommand),
		"module update":                             command.NewModuleUpdateCommandFactory(baseCommand),
		"module delete":                             command.NewModuleDeleteCommandFactory(baseCommand),
		"module list":                               command.NewModuleListCommandFactory(baseCommand),
		"module list-versions":                      command.NewModuleListVersionsCommandFactory(baseCommand),
		"module list-attestations":                  command.NewModuleListAttestationsCommandFactory(baseCommand),
		"module create-attestation":                 command.NewModuleCreateAttestationCommandFactory(baseCommand),
		"module update-attestation":                 command.NewModuleUpdateAttestationCommandFactory(baseCommand),
		"module delete-attestation":                 command.NewModuleDeleteAttestationCommandFactory(baseCommand),
		"module get-version":                        command.NewModuleGetVersionCommandFactory(baseCommand),
		"module delete-version":                     command.NewModuleDeleteVersionCommandFactory(baseCommand),
		"module upload-version":                     command.NewModuleUploadVersionCommandFactory(baseCommand),
		"plan":                                      command.NewPlanCommandFactory(baseCommand),
		"run":                                       command.NewHelpCommandFactory(getHelpText("run")),
		"run cancel":                                command.NewRunCancelCommandFactory(baseCommand),
		"runner-agent":                              command.NewHelpCommandFactory(getHelpText("runner-agent")),
		"runner-agent get":                          command.NewRunnerAgentGetCommandFactory(baseCommand),
		"runner-agent create":                       command.NewRunnerAgentCreateCommandFactory(baseCommand),
		"runner-agent assign-service-account":       command.NewRunnerAgentAssignServiceAccountCommandFactory(baseCommand),
		"runner-agent unassign-service-account":     command.NewRunnerAgentUnassignServiceAccountCommandFactory(baseCommand),
		"runner-agent update":                       command.NewRunnerAgentUpdateCommandFactory(baseCommand),
		"runner-agent delete":                       command.NewRunnerAgentDeleteCommandFactory(baseCommand),
		"service-account":                           command.NewHelpCommandFactory(getHelpText("service-account")),
		"service-account create-oidc-token":         command.NewServiceAccountCreateOIDCTokenCommandFactory(baseCommand),
		"sso":                                       command.NewHelpCommandFactory(getHelpText("sso")),
		"sso login":                                 command.NewLoginCommandFactory(baseCommand),
		"terraform-provider":                        command.NewHelpCommandFactory(getHelpText("terraform-provider")),
		"terraform-provider create":                 command.NewTerraformProviderCreateCommandFactory(baseCommand),
		"terraform-provider upload-version":         command.NewTerraformProviderUploadVersionCommandFactory(baseCommand),
		"terraform-provider-mirror":                 command.NewHelpCommandFactory(getHelpText("terraform-provider-mirror")),
		"terraform-provider-mirror get-version":     command.NewTerraformProviderMirrorGetVersionCommandFactory(baseCommand),
		"terraform-provider-mirror list-versions":   command.NewTerraformProviderMirrorListVersionsCommandFactory(baseCommand),
		"terraform-provider-mirror list-platforms":  command.NewTerraformProviderMirrorListPlatformsCommandFactory(baseCommand),
		"terraform-provider-mirror delete-version":  command.NewTerraformProviderMirrorDeleteVersionCommandFactory(baseCommand),
		"terraform-provider-mirror delete-platform": command.NewTerraformProviderMirrorDeletePlatformCommandFactory(baseCommand),
		"version":                                   command.NewVersionCommandFactory(baseCommand),
		"workspace":                                 command.NewHelpCommandFactory(getHelpText("workspace")),
		"workspace get":                             command.NewWorkspaceGetCommandFactory(baseCommand),
		"workspace create":                          command.NewWorkspaceCreateCommandFactory(baseCommand),
		"workspace update":                          command.NewWorkspaceUpdateCommandFactory(baseCommand),
		"workspace delete":                          command.NewWorkspaceDeleteCommandFactory(baseCommand),
		"workspace list":                            command.NewWorkspaceListCommandFactory(baseCommand),
		"workspace list-memberships":                command.NewWorkspaceListMembershipsCommandFactory(baseCommand),
		"workspace assign-managed-identity":         command.NewWorkspaceAssignManagedIdentityCommandFactory(baseCommand),
		"workspace unassign-managed-identity":       command.NewWorkspaceUnassignManagedIdentityCommandFactory(baseCommand),
		"workspace get-assigned-managed-identities": command.NewWorkspaceGetAssignedManagedIdentitiesCommandFactory(baseCommand),
		"workspace outputs":                         command.NewWorkspaceOutputsCommandFactory(baseCommand),
		"workspace label":                           command.NewWorkspaceLabelCommandFactory(baseCommand),
		"workspace get-terraform-var":               command.NewWorkspaceGetTerraformVarCommandFactory(baseCommand),
		"workspace set-terraform-var":               command.NewWorkspaceSetTerraformVarCommandFactory(baseCommand),
		"workspace delete-terraform-var":            command.NewWorkspaceDeleteTerraformVarCommandFactory(baseCommand),
		"workspace list-terraform-vars":             command.NewWorkspaceListTerraformVarsCommandFactory(baseCommand),
		"workspace set-terraform-vars":              command.NewWorkspaceSetTerraformVarsCommandFactory(baseCommand),
		"workspace list-environment-vars":           command.NewWorkspaceListEnvironmentVarsCommandFactory(baseCommand),
		"workspace set-environment-vars":            command.NewWorkspaceSetEnvironmentVarsCommandFactory(baseCommand),
	}

	// Add the documentation commands.
	commandMap["documentation"] = command.NewHelpCommandFactory(getHelpText("documentation"))
	commandMap["documentation generate"] = command.NewDocumentationGenerateCommandFactory(baseCommand, commandMap)

	// Convert CommandFactory to cli.CommandFactory.
	returnMap := map[string]cli.CommandFactory{}
	for name, helpCommandFactory := range commandMap {
		helpCommand, err := helpCommandFactory()
		if err != nil {
			return nil, err
		}

		returnMap[name] = func() (cli.Command, error) {
			return command.NewWrapper(helpCommand), nil
		}
	}

	return returnMap, nil
}

// helpFunc adds global options to the default help function.
func helpFunc(h cli.HelpFunc, f *flag.FlagSet) cli.HelpFunc {
	return func(commands map[string]cli.CommandFactory) string {
		var headingBuf bytes.Buffer

		// Build the header with colors
		titleColor := color.New(color.Bold, color.FgMagenta)
		boldColor := color.New(color.Bold)
		greenBold := color.New(color.Bold, color.FgGreen)

		fmt.Fprint(&headingBuf, titleColor.Sprint("Welcome to Tharsis!"))
		fmt.Fprint(&headingBuf, " — ")
		fmt.Fprintln(&headingBuf, "An open-source Terraform platform.")

		fmt.Fprint(&headingBuf, boldColor.Sprint("Documentation:"))
		fmt.Fprintln(&headingBuf, " https://tharsis.martian-cloud.io")

		fmt.Fprint(&headingBuf, greenBold.Sprint("Version:"))
		fmt.Fprintln(&headingBuf, " "+Version)
		fmt.Fprintln(&headingBuf)

		// Build global flag usage.
		var usageBuf bytes.Buffer
		usageBuf.Write([]byte("\n"))
		globalFlags := f
		globalFlags.SetOutput(&usageBuf)
		globalFlags.Usage()

		return strings.TrimSpace(headingBuf.String() + h(commands) + usageBuf.String())
	}
}

// predictProfiles returns available profile names for autocompletion.
func predictProfiles(_ complete.Args) []string {
	s, err := settings.ReadSettings(nil)
	if err != nil {
		return nil
	}

	profiles := make([]string, 0, len(s.Profiles))
	for name := range s.Profiles {
		profiles = append(profiles, name)
	}
	return profiles
}

// getHelpText returns the helpText for command.
func getHelpText(commandName string) (string, string) {
	return helpText[commandName][0], helpText[commandName][1]
}

// This should be used for all parent commands that appear on the main page
// i.e., commands that are generally placeholders for subcommands.
var helpText = map[string][2]string{
	"sso": {
		"Log in to the OAuth2 provider and return an authentication token.",
		`
The sso command authenticates the CLI with the OAuth2 provider,
and allows making authenticated calls to Tharsis backend.
`,
	},
	"documentation": {
		"Perform command documentation operations.",
		`
The documentation command(s) perform operations on the documentation.
`,
	},
	"configure": {
		"Create or update a profile.",
		`
The configure command creates or updates a profile. If no
options are specified, the command prompts for values.
`,
	},
	"group": {
		"Do operations on groups.",
		`
Groups are containers for organizing workspaces hierarchically.
They can be nested and inherit variables and managed identities
to children. Use group commands to create, update, delete groups,
set Terraform and environment variables, manage memberships, and
migrate groups between parents.
`,
	},
	"workspace": {
		"Do operations on workspaces.",
		`
Workspaces contain Terraform deployments, state, runs, and variables.
Use workspace commands to create, update, delete workspaces, assign
and unassign managed identities, set Terraform and environment
variables, manage memberships, and view workspace outputs.
`,
	},
	"managed-identity": {
		"Do operations on a managed identity.",
		`
Managed identities provide OIDC-federated credentials for cloud
providers (AWS, Azure, Kubernetes) without storing secrets. Use
managed-identity commands to create, update, delete, and get
managed identities.
`,
	},
	"managed-identity-access-rule": {
		"Do operations on a managed identity access rule.",
		`
Access rules control which runs can use a managed identity based
on conditions like module source or workspace path. Use these
commands to create, update, delete, list, and get access rules.
`,
	},
	"managed-identity-alias": {
		"Do operations on a managed identity alias.",
		`
Aliases allow referencing managed identities from other groups.
Use these commands to create and delete managed identity aliases.
`,
	},
	"module": {
		"Do operations on a terraform module.",
		`
The module registry stores Terraform modules with versioning and
attestation support. Use module commands to create, update, delete
modules, upload versions, manage attestations, and list modules
and versions.
`,
	},
	"terraform-provider": {
		"Do operations on a terraform provider.",
		`
The provider registry stores Terraform providers with versioning
support. Use terraform-provider commands to create providers and
upload provider versions to the registry.
`,
	},
	"terraform-provider-mirror": {
		"Mirror Terraform providers from any Terraform registry.",
		`
The provider mirror caches Terraform providers from any registry
for use within a group hierarchy. It supports Terraform's Provider
Network Mirror Protocol and gives root group owners control over
which providers, platform packages, and registries are available.
Use these commands to sync providers, list versions and platforms,
get version details, and delete versions or platforms.
`,
	},
	"runner-agent": {
		"Do operations on runner agents.",
		`
Runner agents are distributed job executors responsible for
launching Terraform jobs that deploy infrastructure to the cloud.
Use runner-agent commands to create, update, delete, get agents,
and assign or unassign service accounts.
`,
	},
	"service-account": {
		"Create an authentication token for a service account.",
		`
Service accounts provide machine-to-machine authentication for
CI/CD pipelines and automation. Use service-account commands to
create authentication tokens.
`,
	},
	"run": {
		"Do operations on runs.",
		`
Runs are units of execution (plan or apply) that create, update,
or destroy infrastructure resources. Use run commands to cancel
runs gracefully or forcefully.
`,
	},
	"plan": {
		"Create a speculative plan",
		`
The plan command creates a speculative plan to view the changes
Terraform will make to your infrastructure without applying them.
Supports setting run-scoped Terraform and environment variables,
planning destroy runs, and using remote module sources.
`,
	},
	"apply": {
		"Apply a single run.",
		`
The apply command applies a run to create, update, or destroy
infrastructure resources. Supports setting run-scoped Terraform
and environment variables, auto-approving changes, using remote
module sources, and specifying Terraform versions.
`,
	},
	"destroy": {
		"Destroy the workspace state.",
		`
The destroy command destroys all infrastructure resources managed
by a workspace. Similar to apply, it supports setting run-scoped
Terraform and environment variables, auto-approving changes, and
using remote module sources.
`,
	},
}

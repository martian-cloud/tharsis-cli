package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
)

// helpFunc builds the full CLI help output.
func helpFunc(globalFlags *flag.Set) cli.HelpFunc {
	return func(commands map[string]cli.CommandFactory) string {
		bold := color.New(color.Bold)
		green := output.PrimaryColor()

		var sections []string

		// Logo.
		sections = append(sections, buildLogo())

		// Usage.
		sections = append(sections, fmt.Sprintf("%s [global options] <command> [options] <args>", green.Sprint("tharsis")))

		// Command list — top-level only.
		sections = append(sections, buildCommandList(bold, commands))

		// Global flags.
		sections = append(sections, output.CommandHelp(output.CommandHelpInfo{
			Flags: globalFlags,
		}))

		// Legend.
		sections = append(sections, strings.Join([]string{
			bold.Sprint("Legend:"),
			"  " + color.New(color.FgRed).Sprint("*  ") + " required flag",
			"  " + color.New(color.FgYellow).Sprint("!  ") + " deprecated flag",
			"  " + color.New(color.FgGreen).Sprint("...") + " repeatable flag",
		}, "\n"))

		return strings.Join(sections, "\n\n")
	}
}

// getHelpText returns the helpText for command.
func getHelpText(commandName string) (string, string) {
	return helpText[commandName][0], helpText[commandName][1]
}

func buildLogo() string {
	logoLines := strings.Split(strings.TrimRight(figure.NewFigure("THARSIS", "block", true).String(), "\n"), "\n")
	shades := []*color.Color{
		color.New(color.FgHiGreen, color.Bold),
		color.New(color.FgHiGreen),
		color.New(color.FgGreen, color.Bold),
		color.New(color.FgGreen),
		color.New(color.FgGreen),
	}

	var lines []string
	for i, line := range logoLines {
		lines = append(lines, shades[i%len(shades)].Sprint(line))
	}

	dim := color.New(color.Faint)
	lines = append(lines,
		dim.Sprint("  The open-source Terraform platform."),
		fmt.Sprintf("  %s %s", dim.Sprint("Version:"), Version),
		fmt.Sprintf("  %s %s", dim.Sprint("Docs:"), "https://tharsis.martian-cloud.io"),
	)

	return strings.Join(lines, "\n")
}

func buildCommandList(bold *color.Color, commands map[string]cli.CommandFactory) string {
	keys := slices.Sorted(maps.Keys(commands))

	maxLen := 0
	for _, key := range keys {
		if !strings.Contains(key, " ") && len(key) > maxLen {
			maxLen = len(key)
		}
	}

	lines := []string{bold.Sprint("Available Commands:")}
	for _, key := range keys {
		if strings.Contains(key, " ") {
			continue
		}

		cmd, err := commands[key]()
		if err != nil {
			continue
		}

		lines = append(lines, fmt.Sprintf("    %s    %s", key+strings.Repeat(" ", maxLen-len(key)), cmd.Synopsis()))
	}

	return strings.Join(lines, "\n")
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
managed-identity commands to create, update, delete, list, and
get managed identities.
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
	"membership": {
		"Do operations on namespace memberships.",
		`
Namespace memberships control access to groups and workspaces.
Use membership commands to list memberships for a user, service
account, or team.
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
support. Use terraform-provider commands to create, get, list,
update, delete providers, upload versions, manage versions and
platforms.
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
Use runner-agent commands to create, update, delete, list, get
agents, and assign or unassign service accounts.
`,
	},
	"service-account": {
		"Do operations on service accounts.",
		`
Service accounts provide machine-to-machine authentication for
CI/CD pipelines and automation. Use service-account commands to
create, update, delete, list service accounts, and create
authentication tokens.
`,
	},
	"gpg-key": {
		"Do operations on GPG keys.",
		`
GPG keys are used for module attestation verification. Use
gpg-key commands to create, delete, list, and get GPG keys
within a group hierarchy.
`,
	},
	"resource-limit": {
		"Do operations on resource limits.",
		`
Resource limits control the maximum number of resources that
can be created. Use resource-limit commands to list and update
resource limits.
`,
	},
	"state-version": {
		"Do operations on state versions.",
		`
State versions represent snapshots of Terraform state for a
workspace. Use state-version commands to list and get state
versions.
`,
	},
	"team": {
		"Do operations on teams.",
		`
Teams group users together for access management. Use team
commands to create, update, delete, list teams, and manage
team members.
`,
	},
	"user": {
		"Do operations on users.",
		`
Users represent individuals who can access Tharsis. Use user
commands to list and get user details.
`,
	},
	"vcs-provider": {
		"Do operations on VCS providers.",
		`
VCS providers integrate GitHub or GitLab for automatic run
triggering. Use vcs-provider commands to create, update,
delete, list, get, reset OAuth tokens, and create runs.
`,
	},
	"vcs-provider-link": {
		"Do operations on workspace VCS provider links.",
		`
VCS provider links connect workspaces to VCS repositories for
automatic run triggering. Use vcs-provider-link commands to
create, update, delete, and get workspace VCS provider links.
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
	"role": {
		"Do operations on roles.",
		`
Roles define sets of permissions that can be assigned to users,
service accounts, and teams via namespace memberships. Use role
commands to create, update, delete, list roles, and view
available permissions.
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

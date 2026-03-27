package main

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strings"

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

		var buf bytes.Buffer

		// Header.
		fmt.Fprint(&buf, color.New(color.Bold, color.FgHiGreen).Sprint("Welcome to Tharsis"))
		fmt.Fprintln(&buf, " - An open-source Terraform platform.")
		fmt.Fprint(&buf, bold.Sprint("Documentation:"))
		fmt.Fprintln(&buf, " https://tharsis.martian-cloud.io")
		fmt.Fprint(&buf, green.Sprint("Version:"))
		fmt.Fprintln(&buf, " "+Version)
		fmt.Fprintln(&buf)

		// Usage.
		fmt.Fprintf(&buf, "%s [global options] <command> [options] <args>\n\n", green.Sprint("tharsis"))

		// Command list — top-level only.
		keys := slices.Sorted(maps.Keys(commands))

		maxLen := 0
		for _, key := range keys {
			if !strings.Contains(key, " ") && len(key) > maxLen {
				maxLen = len(key)
			}
		}

		fmt.Fprintln(&buf, bold.Sprint("Available Commands:"))
		for _, key := range keys {
			if strings.Contains(key, " ") {
				continue
			}

			cmd, err := commands[key]()
			if err != nil {
				continue
			}

			fmt.Fprintf(&buf, "    %s    %s\n", key+strings.Repeat(" ", maxLen-len(key)), cmd.Synopsis())
		}

		// Global flags.
		globalFlagsOutput := output.CommandHelp(output.CommandHelpInfo{
			Flags: globalFlags,
		})
		buf.WriteString(strings.TrimRight(globalFlagsOutput, "\n"))
		buf.WriteString("\n\n")

		// Legend.
		fmt.Fprintln(&buf, bold.Sprint("Legend:"))
		fmt.Fprintln(&buf, "  "+color.New(color.FgRed).Sprint("*  ")+" required flag")
		fmt.Fprintln(&buf, "  "+color.New(color.FgYellow).Sprint("!  ")+" deprecated flag")
		fmt.Fprintln(&buf, "  "+color.New(color.FgGreen).Sprint("...")+" repeatable flag")

		return strings.TrimSpace(buf.String())
	}
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

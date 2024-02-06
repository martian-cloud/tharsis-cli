package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/cli"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// managedIdentityAccessRuleCommand is the top-level structure for the managed-identity-access-rule command.
type managedIdentityAccessRuleCommand struct {
	meta *Metadata
}

// NewManagedIdentityAccessRuleCommandFactory returns a managedIdentityAccessRuleCommand struct.
func NewManagedIdentityAccessRuleCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return managedIdentityAccessRuleCommand{
			meta: meta,
		}, nil
	}
}

func (m managedIdentityAccessRuleCommand) Run(args []string) int {
	m.meta.Logger.Debugf("Starting the 'managed-identity-access-rule' command with %d arguments:", len(args))
	for ix, arg := range args {
		m.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Show the help text.
	m.meta.UI.Output(m.HelpManagedIdentityAccessRule(true))
	return 1
}

func (m managedIdentityAccessRuleCommand) Synopsis() string {
	return "Do operations on a managed identity access rule."
}

func (m managedIdentityAccessRuleCommand) Help() string {
	return m.HelpManagedIdentityAccessRule(false)
}

// HelpManagedIdentityAccessRule produces the help string for the 'managed-identity-access-rule' command.
func (m managedIdentityAccessRuleCommand) HelpManagedIdentityAccessRule(subCommands bool) string {
	usage := fmt.Sprintf(`
Usage: %s [global options] managed-identity-access-rule ...

   The managed-identity-access-rule commands do operations on a managed identity access rule.
`, m.meta.BinaryName)
	sc := `

Subcommands:
    create                       Create a new managed identity access rule.
    delete                       Delete a managed identity access rule.
    get                          Get a single managed identity access rule.
    update                       Update a managed identity access rule.`

	// Avoid duplicate subcommands when -h is used.
	if subCommands {
		return usage + sc
	}

	return usage
}

// buildModuleAttestationPolicies builds a list of module attestation policies from strings.
// It is used by multiple sub-commands.
// Each string is of this form: "PredicateType=someval,PublicKeyFile=/path/to/file"
func buildModuleAttestationPolicies(args []string) ([]sdktypes.ManagedIdentityAccessRuleModuleAttestationPolicy, error) {
	result := []sdktypes.ManagedIdentityAccessRuleModuleAttestationPolicy{}

	for _, arg := range args {
		var predicateType *string
		var filename string

		for _, kv := range strings.Split(arg, ",") {
			parts := strings.Split(kv, "=")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid module attestation policy format: %s", arg)
			}

			switch parts[0] {
			case "PredicateType":
				predicateType = &parts[1]
			case "PublicKeyFile":
				filename = parts[1]
			default:
				return nil, fmt.Errorf("invalid module attestation policy key: %s", parts[0])
			}
		}

		// Make sure the filename was supplied
		if filename == "" {
			return nil, fmt.Errorf("missing PublicKeyFile in module attestation policy: %s", arg)
		}

		// Read the public key from the file.
		publicKey, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read public key from file %s: %w", filename, err)
		}

		// Add the module attestation policy to the result.
		result = append(result, sdktypes.ManagedIdentityAccessRuleModuleAttestationPolicy{
			PredicateType: predicateType,
			PublicKey:     string(publicKey),
		})

	}

	return result, nil
}

package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tableformatter"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// providerCreateCommand is the top-level structure for the provider create command.
type providerCreateCommand struct {
	meta *Metadata
}

// NewProviderCreateCommandFactory returns a providerCreateCommand struct.
func NewProviderCreateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return providerCreateCommand{
			meta: meta,
		}, nil
	}
}

func (pcc providerCreateCommand) Run(args []string) int {
	pcc.meta.Logger.Debugf("Starting the 'provider create' command with %d arguments:", len(args))
	for ix, arg := range args {
		pcc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := pcc.meta.ReadSettings()
	if err != nil {
		pcc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		pcc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return pcc.doProviderCreate(ctx, client, args)
}

func (pcc providerCreateCommand) doProviderCreate(ctx context.Context, client *tharsis.Client, opts []string) int {
	pcc.meta.Logger.Debugf("will do provider create, %d opts", len(opts))

	defs := pcc.buildProviderCreateDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(pcc.meta.BinaryName+" provider create", defs, opts)
	if err != nil {
		pcc.meta.Logger.Error(output.FormatError("failed to parse provider create options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		pcc.meta.Logger.Error(output.FormatError("missing provider create path", nil), pcc.HelpProviderCreate())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive provider create arguments: %s", cmdArgs)
		pcc.meta.Logger.Error(output.FormatError(msg, nil), pcc.HelpProviderCreate())
		return 1
	}

	providerPath := cmdArgs[0]
	toJSON := getOption("json", "", cmdOpts)[0] == "1"
	repositoryURL := getOption("repository-url", "", cmdOpts)[0]
	private, err := getBoolOptionValue("private", "true", cmdOpts)
	if err != nil {
		pcc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Error is already logged.
	if !isResourcePathValid(pcc.meta, providerPath) {
		return 1
	}

	// Create provider
	index := strings.LastIndex(providerPath, "/")
	provider, err := client.TerraformProvider.CreateProvider(ctx, &types.CreateTerraformProviderInput{
		Name:          providerPath[index+1:],
		GroupPath:     providerPath[:index],
		Private:       private,
		RepositoryURL: repositoryURL,
	})
	if err != nil {
		pcc.meta.UI.Error(output.FormatError("failed to create provider", err))
		return 1
	}

	return pcc.outputProvider(toJSON, provider)
}

func (pcc providerCreateCommand) outputProvider(toJSON bool, provider *types.TerraformProvider) int {
	if toJSON {
		buf, err := objectToJSON(provider)
		if err != nil {
			pcc.meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		pcc.meta.UI.Output(string(buf))
	} else {
		tableInput := [][]string{
			{
				"id",
				"name",
				"resource path",
				"registry namespace",
				"private",
				"repository url",
			},
			{
				provider.Metadata.ID,
				provider.Name,
				provider.ResourcePath,
				provider.RegistryNamespace,
				fmt.Sprintf("%t", provider.Private),
				provider.RepositoryURL,
			},
		}
		pcc.meta.UI.Output(tableformatter.FormatTable(tableInput))
	}

	return 0
}

// buildProviderCreateDefs returns defs used by provider create command.
func (pcc providerCreateCommand) buildProviderCreateDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"private": {
			Arguments: []string{"Private"},
			Synopsis:  "Set private to false to allow all groups to view and use the provider (default=true).",
		},
		"repository-url": {
			Arguments: []string{"Repository_URL"},
			Synopsis:  "The repository URL for this provider.",
		},
	}

	return buildJSONOptionDefs(defs)
}

func (pcc providerCreateCommand) Synopsis() string {
	return "Create a new provider."
}

func (pcc providerCreateCommand) Help() string {
	return pcc.HelpProviderCreate()
}

// HelpProviderCreate produces the help string for the 'provider create' command.
func (pcc providerCreateCommand) HelpProviderCreate() string {
	return fmt.Sprintf(`
Usage: %s [global options] provider create [options] <full_path>

   The provider create command creates a new provider.

%s

`, pcc.meta.BinaryName, buildHelpText(pcc.buildProviderCreateDefs()))
}

// The End.

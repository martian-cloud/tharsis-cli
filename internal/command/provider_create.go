package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
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

func (p providerCreateCommand) Run(args []string) int {
	p.meta.Logger.Debugf("Starting the 'provider create' command with %d arguments:", len(args))
	for ix, arg := range args {
		p.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := p.meta.ReadSettings()
	if err != nil {
		p.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		p.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return p.doProviderCreate(ctx, client, args)
}

func (p providerCreateCommand) doProviderCreate(ctx context.Context, client *tharsis.Client, opts []string) int {
	p.meta.Logger.Debugf("will do provider create, %d opts", len(opts))

	defs := buildProviderCreateDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(p.meta.BinaryName+" provider create", defs, opts)
	if err != nil {
		p.meta.Logger.Error(output.FormatError("failed to parse provider create options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		p.meta.Logger.Error(output.FormatError("missing provider create path", nil), p.HelpProviderCreate())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive provider create arguments: %s", cmdArgs)
		p.meta.Logger.Error(output.FormatError(msg, nil), p.HelpProviderCreate())
		return 1
	}

	providerPath := cmdArgs[0]
	toJSON := getOption("json", "", cmdOpts)[0] == "1"
	private := getOption("private", "1", cmdOpts)[0] == "1"
	repositoryURL := getOption("repository-url", "", cmdOpts)[0]

	// Error is already logged.
	if !isResourcePathValid(p.meta, providerPath) {
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
		p.meta.UI.Error(output.FormatError("failed to create provider", err))
		return 1
	}

	return p.outputProvider(toJSON, provider)
}

func (p providerCreateCommand) outputProvider(toJSON bool, provider *types.TerraformProvider) int {
	if toJSON {
		buf, err := objectToJSON(provider)
		if err != nil {
			p.meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		p.meta.UI.Output(string(buf))
	} else {
		// Format the output.
		p.meta.UI.Output("provider %s output:")
		p.meta.UI.Output(fmt.Sprintf("\n              name: %s", provider.Name))
		p.meta.UI.Output(fmt.Sprintf("     resource path: %s", provider.ResourcePath))
		p.meta.UI.Output(fmt.Sprintf("           private: %t", provider.Private))
		p.meta.UI.Output(fmt.Sprintf("    Repository URL: %s", provider.RepositoryURL))
		p.meta.UI.Output(fmt.Sprintf("                ID: %s", provider.Metadata.ID))
	}

	return 0
}

// buildProviderCreateDefs returns defs used by provider create command.
func buildProviderCreateDefs() optparser.OptionDefinitions {
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

func (p providerCreateCommand) Synopsis() string {
	return "Create a new provider."
}

func (p providerCreateCommand) Help() string {
	return p.HelpProviderCreate()
}

// HelpProviderCreate produces the help string for the 'provider create' command.
func (p providerCreateCommand) HelpProviderCreate() string {
	return fmt.Sprintf(`
Usage: %s [global options] provider create [options] <full_path>

   The provider create command creates a new provider.

%s

`, p.meta.BinaryName, buildHelpText(buildProviderCreateDefs()))
}

// The End.

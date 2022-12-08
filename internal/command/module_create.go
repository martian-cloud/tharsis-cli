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

// moduleCreateCommand is the top-level structure for the module create command.
type moduleCreateCommand struct {
	meta *Metadata
}

// NewModuleCreateCommandFactory returns a moduleCreateCommand struct.
func NewModuleCreateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return moduleCreateCommand{
			meta: meta,
		}, nil
	}
}

func (p moduleCreateCommand) Run(args []string) int {
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

	return p.doModuleCreate(ctx, client, args)
}

func (p moduleCreateCommand) doModuleCreate(ctx context.Context, client *tharsis.Client, opts []string) int {
	defs := buildModuleCreateDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(p.meta.BinaryName+" module create", defs, opts)
	if err != nil {
		p.meta.Logger.Error(output.FormatError("failed to parse module create options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		p.meta.Logger.Error(output.FormatError("missing module create path", nil), p.HelpModuleCreate())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive module create arguments: %s", cmdArgs)
		p.meta.Logger.Error(output.FormatError(msg, nil), p.HelpModuleCreate())
		return 1
	}

	modulePath := cmdArgs[0]
	toJSON := getOption("json", "", cmdOpts)[0] == "1"
	private := getOption("private", "1", cmdOpts)[0] == "1"
	repositoryURL := getOption("repository-url", "", cmdOpts)[0]

	// Error is already logged.
	if !isResourcePathValid(p.meta, modulePath) {
		return 1
	}

	// Create module
	pathParts := strings.Split(modulePath, "/")
	if len(pathParts) < 3 {
		p.meta.Logger.Error(output.FormatError("resource path is not valid", nil))
		return 1
	}

	module, err := client.TerraformModule.CreateModule(ctx, &types.CreateTerraformModuleInput{
		GroupPath:     strings.Join(pathParts[:len(pathParts)-2], "/"),
		Name:          pathParts[len(pathParts)-2],
		System:        pathParts[len(pathParts)-1],
		Private:       private,
		RepositoryURL: repositoryURL,
	})
	if err != nil {
		p.meta.UI.Error(output.FormatError("failed to create module", err))
		return 1
	}

	return p.outputModule(toJSON, module)
}

func (p moduleCreateCommand) outputModule(toJSON bool, module *types.TerraformModule) int {
	if toJSON {
		buf, err := objectToJSON(module)
		if err != nil {
			p.meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		p.meta.UI.Output(string(buf))
	} else {
		// TODO: Update other commands to use table formatter for resource output
		tableInput := [][]string{
			{"id", "name", "resource path", "private", "repository url"},
			{module.Metadata.ID, module.Name, module.ResourcePath, fmt.Sprintf("%t", module.Private), module.ResourcePath},
		}
		p.meta.UI.Output(tableformatter.FormatTable(tableInput))
	}

	return 0
}

// buildModuleCreateDefs returns defs used by module create command.
func buildModuleCreateDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"private": {
			Arguments: []string{"Private"},
			Synopsis:  "Set private to false to allow all groups to view and use the module (default=true).",
		},
		"repository-url": {
			Arguments: []string{"Repository_URL"},
			Synopsis:  "The repository URL for this module.",
		},
	}

	return buildJSONOptionDefs(defs)
}

func (p moduleCreateCommand) Synopsis() string {
	return "Create a new module."
}

func (p moduleCreateCommand) Help() string {
	return p.HelpModuleCreate()
}

// HelpModuleCreate produces the help string for the 'module create' command.
func (p moduleCreateCommand) HelpModuleCreate() string {
	return fmt.Sprintf(`
Usage: %s [global options] module create [options] <module_resource_path>

   The module create command creates a new module.

%s

`, p.meta.BinaryName, buildHelpText(buildModuleCreateDefs()))
}

// The End.

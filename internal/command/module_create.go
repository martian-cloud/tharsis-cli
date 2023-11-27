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
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
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

func (mcc moduleCreateCommand) Run(args []string) int {
	mcc.meta.Logger.Debugf("Starting the 'module create' command with %d arguments:", len(args))
	for ix, arg := range args {
		mcc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := mcc.meta.ReadSettings()
	if err != nil {
		mcc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		mcc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return mcc.doModuleCreate(ctx, client, args)
}

func (mcc moduleCreateCommand) doModuleCreate(ctx context.Context, client *tharsis.Client, opts []string) int {
	mcc.meta.Logger.Debugf("will do module create, %d opts", len(opts))

	defs := buildSharedModuleDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(mcc.meta.BinaryName+" module create", defs, opts)
	if err != nil {
		mcc.meta.Logger.Error(output.FormatError("failed to parse module create options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		mcc.meta.Logger.Error(output.FormatError("missing module create path", nil), mcc.HelpModuleCreate())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive module create arguments: %s", cmdArgs)
		mcc.meta.Logger.Error(output.FormatError(msg, nil), mcc.HelpModuleCreate())
		return 1
	}

	modulePath := cmdArgs[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		mcc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	repositoryURL := getOption("repository-url", "", cmdOpts)[0]
	private, err := getBoolOptionValue("private", "true", cmdOpts)
	if err != nil {
		mcc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	ifNotExists, err := getBoolOptionValue("if-not-exists", "false", cmdOpts)
	if err != nil {
		mcc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Error is already logged.
	if !isResourcePathValid(mcc.meta, modulePath) {
		return 1
	}

	if ifNotExists {
		module, gErr := client.TerraformModule.GetModule(ctx, &sdktypes.GetTerraformModuleInput{
			Path: &modulePath,
		})
		if gErr != nil && !tharsis.IsNotFoundError(gErr) {
			mcc.meta.Logger.Error(output.FormatError("failed to get module", gErr))
			return 1
		}

		if module != nil {
			return outputModule(mcc.meta, toJSON, module)
		}
	}

	// Create module
	pathParts := strings.Split(modulePath, "/")
	if len(pathParts) < 3 {
		mcc.meta.Logger.Error(output.FormatError("resource path is not valid", nil))
		return 1
	}

	input := &sdktypes.CreateTerraformModuleInput{
		GroupPath:     strings.Join(pathParts[:len(pathParts)-2], "/"),
		Name:          pathParts[len(pathParts)-2],
		System:        pathParts[len(pathParts)-1],
		Private:       private,
		RepositoryURL: repositoryURL,
	}
	mcc.meta.Logger.Debugf("module create input: %#v", input)

	module, err := client.TerraformModule.CreateModule(ctx, input)
	if err != nil {
		mcc.meta.UI.Error(output.FormatError("failed to create module", err))
		return 1
	}

	return outputModule(mcc.meta, toJSON, module)
}

// outputModule is the final output for most module operations.
func outputModule(meta *Metadata, toJSON bool, module *sdktypes.TerraformModule) int {
	if toJSON {
		buf, err := objectToJSON(module)
		if err != nil {
			meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		meta.UI.Output(string(buf))
	} else {
		tableInput := [][]string{
			{"id", "name", "resource path", "private", "repository url"},
			{module.Metadata.ID, module.Name, module.ResourcePath, fmt.Sprintf("%t", module.Private), module.RepositoryURL},
		}
		meta.UI.Output(tableformatter.FormatTable(tableInput))
	}

	return 0
}

// buildSharedModuleDefs returns defs used by module create command.
func buildSharedModuleDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"private": {
			Arguments: []string{"Private"},
			Synopsis:  "Set private to false to allow all groups to view and use the module (default=true).",
		},
		"repository-url": {
			Arguments: []string{"Repository_URL"},
			Synopsis:  "The repository URL for this module.",
		},
		"if-not-exists": {
			Arguments: []string{},
			Synopsis:  "Create a module if it does not already exist.",
		},
	}

	return buildJSONOptionDefs(defs)
}

func (mcc moduleCreateCommand) Synopsis() string {
	return "Create a new module."
}

func (mcc moduleCreateCommand) Help() string {
	return mcc.HelpModuleCreate()
}

// HelpModuleCreate produces the help string for the 'module create' command.
func (mcc moduleCreateCommand) HelpModuleCreate() string {
	return fmt.Sprintf(`
Usage: %s [global options] module create [options] <module-path>

   The module create command creates a new module.
   Idempotent when used with --if-not-exists option.

%s

`, mcc.meta.BinaryName, buildHelpText(buildSharedModuleDefs()))
}

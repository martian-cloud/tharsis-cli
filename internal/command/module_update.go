package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// moduleUpdateCommand is the top-level structure for the module update command.
type moduleUpdateCommand struct {
	meta *Metadata
}

// NewModuleUpdateCommandFactory returns a moduleUpdateCommand struct.
func NewModuleUpdateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return moduleUpdateCommand{
			meta: meta,
		}, nil
	}
}

func (muc moduleUpdateCommand) Run(args []string) int {
	muc.meta.Logger.Debugf("Starting the 'module update' command with %d arguments:", len(args))
	for ix, arg := range args {
		muc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := muc.meta.GetSDKClient()
	if err != nil {
		muc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return muc.doModuleUpdate(ctx, client, args)
}

func (muc moduleUpdateCommand) doModuleUpdate(ctx context.Context, client *tharsis.Client, opts []string) int {
	muc.meta.Logger.Debugf("will do module update, %d opts", len(opts))

	defs := buildSharedModuleDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(muc.meta.BinaryName+" module update", defs, opts)
	if err != nil {
		muc.meta.Logger.Error(output.FormatError("failed to parse module update options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		muc.meta.Logger.Error(output.FormatError("missing module update path", nil), muc.HelpModuleUpdate())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive module update arguments: %s", cmdArgs)
		muc.meta.Logger.Error(output.FormatError(msg, nil), muc.HelpModuleUpdate())
		return 1
	}

	modulePath := cmdArgs[0]
	repositoryURL := getOption("repository-url", "", cmdOpts)[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		muc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	private, err := getBoolOptionValue("private", "true", cmdOpts)
	if err != nil {
		muc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	actualPath := trn.ToPath(modulePath)
	if !isResourcePathValid(muc.meta, actualPath) {
		return 1
	}

	// Get the module so, we can find it's ID.
	module, err := client.TerraformModule.GetModule(ctx, &sdktypes.GetTerraformModuleInput{Path: &actualPath}) // Use extracted path
	if err != nil {
		muc.meta.Logger.Error(output.FormatError("failed to get module", err))
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.UpdateTerraformModuleInput{
		ID:      module.Metadata.ID,
		Private: &private,
	}

	if repositoryURL != "" {
		input.RepositoryURL = &repositoryURL
	}

	muc.meta.Logger.Debugf("module update input: %#v", input)

	// Update the module.
	updatedModule, err := client.TerraformModule.UpdateModule(ctx, input)
	if err != nil {
		muc.meta.Logger.Error(output.FormatError("failed to update module", err))
		return 1
	}

	return outputModule(muc.meta, toJSON, updatedModule)
}

func (muc moduleUpdateCommand) Synopsis() string {
	return "Update a module."
}

func (muc moduleUpdateCommand) Help() string {
	return muc.HelpModuleUpdate()
}

// HelpModuleUpdate produces the help string for the 'module update' command.
func (muc moduleUpdateCommand) HelpModuleUpdate() string {
	return fmt.Sprintf(`
Usage: %s [global options] module update [options] <module-path>

   The module update command updates a module. Shows final
   output as JSON, if specified.

%s

`, muc.meta.BinaryName, buildHelpText(buildSharedModuleDefs()))
}

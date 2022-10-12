package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
)

// destroyCommand is the top-level structure for the destroy command.
type destroyCommand struct {
	meta *Metadata
}

// NewDestroyCommandFactory returns a destroyCommand struct.
func NewDestroyCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return destroyCommand{
			meta: meta,
		}, nil
	}
}

func (dc destroyCommand) Run(args []string) int {
	dc.meta.Logger.Debugf("Starting the 'destroy' command with %d arguments:", len(args))
	for ix, arg := range args {
		dc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := dc.meta.ReadSettings()
	if err != nil {
		dc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		dc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return dc.doDestroy(ctx, client, args)
}

func (dc destroyCommand) doDestroy(ctx context.Context, client *tharsis.Client, opts []string) int {
	dc.meta.Logger.Debugf("will do destroy, %d opts", len(opts))

	// Build option definitions for this command.
	defs := buildCommonRunOptionDefs()
	buildCommonApplyOptionDefs(defs)
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(dc.meta.BinaryName+" destroy", defs, opts)
	if err != nil {
		dc.meta.Logger.Error(output.FormatError("failed to parse destroy argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		dc.meta.Logger.Error(output.FormatError("missing destroy workspace path", nil), dc.HelpDestroy())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive destroy arguments: %s", cmdArgs)
		dc.meta.Logger.Error(output.FormatError(msg, nil), dc.HelpDestroy())
		return 1
	}

	workspacePath := cmdArgs[0]
	directoryPath := getOption("directory-path", "", cmdOpts)[0]
	comment := getOption("comment", "", cmdOpts)[0]
	autoApprove := getOption("auto-approve", "", cmdOpts)[0] == "1"
	inputRequired, err := getBoolOptionValue("input", "true", cmdOpts)
	if err != nil {
		dc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// TODO remove the requirement of having to pass in the module source /
	// configuration version for destroy operations.
	// Update the API to automatically grab the last module that was deployed on the workspace.
	moduleSource := getOption("module-source", "", cmdOpts)[0]
	moduleVersion := getOption("module-version", "", cmdOpts)[0]
	tfVariables := getOption("tf-var", "", cmdOpts)
	envVariables := getOption("env-var", "", cmdOpts)
	tfVarFile := getOption("tf-var-file", "", cmdOpts)[0]
	envVarFile := getOption("env-var-file", "", cmdOpts)[0]
	terraformVersion := getOption("terraform-version", "", cmdOpts)[0]

	// Error is already logged.
	if !isNamespacePathValid(dc.meta, workspacePath) {
		return 1
	}

	// Do the inner plan. Make it _NON_-speculative.
	createdRun, exitCode := createRun(ctx, client, dc.meta, &runInput{
		workspacePath:    workspacePath,
		directoryPath:    directoryPath,
		tfVarFilePath:    tfVarFile,
		envVarFilePath:   envVarFile,
		moduleSource:     moduleSource,
		moduleVersion:    moduleVersion,
		terraformVersion: terraformVersion,
		tfVariables:      tfVariables,
		envVariables:     envVariables,
		isDestroy:        true,
		isSpeculative:    false,
	})
	if exitCode != 0 {
		// The error message has already been logged.
		return exitCode
	}

	return startApplyStage(ctx, comment, autoApprove, inputRequired, client, createdRun, dc.meta)
}

func (dc destroyCommand) Synopsis() string {
	return "Destroy the workspace state."
}

func (dc destroyCommand) Help() string {
	return dc.HelpDestroy()
}

// HelpDestroy prints the help string for the 'destroy' command.
func (dc destroyCommand) HelpDestroy() string {
	defs := buildCommonRunOptionDefs()
	buildCommonApplyOptionDefs(defs)

	return fmt.Sprintf(`
Usage: %s [global options] destroy [options] <workspace>

   The destroy command destroys a workspace state. Similar
   to the apply command, it supports setting run-scoped
   Terraform and environment variables, using remote
   modules, etc.

%s


Combining --tf-var or --env-var and --tf-var-file or --env-var-file is not allowed.

`, dc.meta.BinaryName, buildHelpText(defs))
}

// The End.

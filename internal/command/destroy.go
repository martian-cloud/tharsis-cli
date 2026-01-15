package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
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

	client, err := dc.meta.GetSDKClient()
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
	autoApprove, err := getBoolOptionValue("auto-approve", "false", cmdOpts)
	if err != nil {
		dc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
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
	tfVariables := getOptionSlice("tf-var", cmdOpts)
	envVariables := getOptionSlice("env-var", cmdOpts)
	tfVarFiles := getOptionSlice("tf-var-file", cmdOpts)
	envVarFiles := getOptionSlice("env-var-file", cmdOpts)
	terraformVersion := getOption("terraform-version", "", cmdOpts)[0]
	targetAddresses := getOptionSlice("target", cmdOpts)
	refresh, err := getBoolOptionValue("refresh", "true", cmdOpts)
	if err != nil {
		dc.meta.UI.Error(output.FormatError("failed to parse boolean value for -refresh option", err))
		return 1
	}

	// Extract path from TRN if needed, then validate path (error is already logged by validation function)
	actualPath := trn.ToPath(workspacePath)
	if !isNamespacePathValid(dc.meta, actualPath) {
		return 1
	}

	// Do the inner plan. Make it _NON_-speculative.
	createdRun, exitCode := createRun(ctx, client, dc.meta, &runInput{
		workspacePath:    workspacePath,
		directoryPath:    directoryPath,
		tfVarFilePath:    tfVarFiles,
		envVarFilePath:   envVarFiles,
		moduleSource:     moduleSource,
		moduleVersion:    moduleVersion,
		terraformVersion: terraformVersion,
		tfVariables:      tfVariables,
		envVariables:     envVariables,
		isDestroy:        true,
		isSpeculative:    false,
		targetAddresses:  targetAddresses,
		refresh:          refresh,
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

   Terraform variables may be passed in via supported
   options or from the environment with a 'TF_VAR_'
   prefix.

   Variable parsing precedence:
     1. Terraform variables from the environment.
     2. terraform.tfvars file from module's directory,
        if present.
     3. terraform.tfvars.json file from module's
        directory, if present.
     4. *.auto.tfvars, *.auto.tfvars.json files
        from the module's directory, if present.
     5. --tf-var-file option(s).
     6. --tf-var option(s).

   NOTE: If the same variable is assigned multiple
   values, the last value found will be used. A
   --tf-var option will override the values from a
   *.tfvars file which will override values from
   the environment.

%s

`, dc.meta.BinaryName, buildHelpText(defs))
}

// Package command contains the logic for processing
// all the commands and subcommands. It uses the Tharsis
// SDK to interface with Tharsis API's remote Terraform
// backend.
package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/job"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

const (

	// Run status string value for a successful apply.
	applySucceededRunValue = "applied"

	// Apply status string value for a successful apply.
	applySucceededApplyValue = "finished"
)

// applyCommand is the top-level structure for the apply command.
type applyCommand struct {
	meta *Metadata
}

// NewApplyCommandFactory returns a applyCommand struct.
func NewApplyCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return applyCommand{
			meta: meta,
		}, nil
	}
}

func (ac applyCommand) Run(args []string) int {
	ac.meta.Logger.Debugf("Starting the 'apply' command with %d arguments:", len(args))
	for ix, arg := range args {
		ac.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := ac.meta.ReadSettings()
	if err != nil {
		ac.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		ac.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return ac.doApply(ctx, client, args)
}

func (ac applyCommand) doApply(ctx context.Context, client *tharsis.Client, opts []string) int {
	ac.meta.Logger.Debugf("will do apply, %d opts", len(opts))

	// Build option definitions for this command.
	defs := buildCommonRunOptionDefs()
	buildCommonApplyOptionDefs(defs)
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(ac.meta.BinaryName+" apply", defs, opts)
	if err != nil {
		ac.meta.Logger.Error(output.FormatError("failed to parse apply argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		ac.meta.Logger.Error(output.FormatError("missing apply workspace path", nil), ac.HelpApply())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive apply arguments: %s", cmdArgs)
		ac.meta.Logger.Error(output.FormatError(msg, nil), ac.HelpApply())
		return 1
	}

	workspacePath := cmdArgs[0]
	directoryPath := getOption("directory-path", "", cmdOpts)[0]
	comment := getOption("comment", "", cmdOpts)[0]
	autoApprove := getOption("auto-approve", "", cmdOpts)[0] == "1"
	inputRequired, err := getBoolOptionValue("input", "true", cmdOpts)
	if err != nil {
		ac.meta.UI.Error(err.Error())
		return 1
	}
	moduleSource := getOption("module-source", "", cmdOpts)[0]
	moduleVersion := getOption("module-version", "", cmdOpts)[0]
	tfVariables := getOption("tf-var", "", cmdOpts)
	envVariables := getOption("env-var", "", cmdOpts)
	tfVarFile := getOption("tf-var-file", "", cmdOpts)[0]
	envVarFile := getOption("env-var-file", "", cmdOpts)[0]
	terraformVersion := getOption("terraform-version", "", cmdOpts)[0]

	// Error is already logged.
	if !isNamespacePathValid(ac.meta, workspacePath) {
		return 1
	}

	// Do the inner plan.  Make it _NON_-speculative.
	createdRun, exitCode := createRun(ctx, client, ac.meta, &runInput{
		workspacePath:    workspacePath,
		directoryPath:    directoryPath,
		tfVarFilePath:    tfVarFile,
		envVarFilePath:   envVarFile,
		moduleSource:     moduleSource,
		moduleVersion:    moduleVersion,
		terraformVersion: terraformVersion,
		tfVariables:      tfVariables,
		envVariables:     envVariables,
		isDestroy:        false,
		isSpeculative:    false,
	})
	if exitCode != 0 {
		// The error message has already been logged.
		return exitCode
	}

	return startApplyStage(ctx, comment, autoApprove, inputRequired, client, createdRun, ac.meta)
}

// startApplyStage a run after it has been planned.
func startApplyStage(ctx context.Context, comment string, autoApprove, inputRequired bool,
	client *tharsis.Client, createdRun *sdktypes.Run, meta *Metadata) int {

	// If the run has transitioned to plannedAndFinished,
	// plan contains no changes and apply does not have to be run.
	if createdRun.Status == sdktypes.RunPlannedAndFinished {
		meta.UI.Output("Stopping since plan had no changes.")
		return 0
	}

	// Return if only inputRequired is false and autoApprove is not set.
	if !inputRequired && !autoApprove {
		meta.UI.Output("Will not apply the plan since -input was false.")
		return 0
	}

	// Ask for approval unless autoApprove is true.
	if autoApprove {
		meta.UI.Output("\nAuto-approving.\n")
	} else {
		meta.UI.Output("\nDo you approve to apply the above plan?\n")
		answer, err := meta.UI.Ask("  only 'yes' will be accepted: ")
		if err != nil {
			meta.Logger.Error(output.FormatError("failed to ask for approval of the plan", err))
			return 1
		}
		if answer != "yes" {
			meta.UI.Output("Approval response was negative.  Will NOT apply the plan.")
			return 0
		}
		// Prettify the output.
		meta.UI.Output("\n\n")
	}

	// Prepare the inputs.
	input := &sdktypes.ApplyRunInput{RunID: createdRun.Metadata.ID}
	if comment != "" {
		input.Comment = &comment
	}
	meta.Logger.Debugf("run apply input: %#v", input)

	// Apply the run.
	appliedRun, err := client.Run.ApplyRun(ctx, input)
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to apply a run", err))
		return 1
	}

	// Make sure the run has an apply.
	if appliedRun.Apply == nil {
		msg := fmt.Sprintf("Created run does not have an apply: %s", appliedRun.Metadata.ID)
		meta.Logger.Error(output.FormatError(msg, nil))
		return 1
	}

	// Display the run apply job's logs.
	applyLogChannel, err := client.Job.GetJobLogs(ctx, &sdktypes.GetJobLogsInput{
		ID:          *appliedRun.Apply.CurrentJobID,
		StartOffset: 0,
		Limit:       logLimit,
	})
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to connect to read apply logs", err))
		return 1
	}
	err = job.DisplayLogs(applyLogChannel, meta.UI)
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to read apply logs", err))
		return 1
	}

	// Check whether the apply passed (vs. failed).
	finishedRun, err := client.Run.GetRun(ctx, &sdktypes.GetRunInput{ID: createdRun.Metadata.ID})
	if err != nil {
		meta.Logger.Error(output.FormatError("Failed to get post-apply run", err))
		return 1
	}

	// If an apply job succeeds, finishedRun.Status is "applied" and
	// finishedRun.Apply.Status is "finished".
	meta.Logger.Debugf("post-apply run status: %s", finishedRun.Status)
	meta.Logger.Debugf("post-apply run.apply.status: %s", finishedRun.Apply.Status)
	if finishedRun.Status != applySucceededRunValue {
		// Status is already printed in the jog logs, so no need to log it here.
		return 1
	}
	if finishedRun.Apply.Status != applySucceededApplyValue {
		meta.Logger.Errorf("Apply status: %s", finishedRun.Apply.Status)
		return 1
	}

	return 0
}

// buildCommonApplyOptionDefs assigns common option definitions
// used by both apply and destroy commands.
func buildCommonApplyOptionDefs(defs optparser.OptionDefinitions) optparser.OptionDefinitions {
	commonDefs := optparser.OptionDefinitions{
		"module-source": {
			Arguments: []string{"Module_Source"},
			Synopsis:  "Remote module source specification.",
		},
		"module-version": {
			Arguments: []string{"Module_Version"},
			Synopsis:  "Remote module version number--defaults to latest.",
		},
		"comment": {
			Arguments: []string{"Comment"},
			Synopsis:  "Comment for the action.",
		},
		"auto-approve": {
			Arguments: []string{},
			Synopsis:  "Do not ask for approval; take approval as already granted.",
		},
		"input": {
			Arguments: []string{"true/false"},
			Synopsis:  "Do ask for user input. (default true).",
		},
	}

	// Populate existing defs.
	for k, v := range commonDefs {
		defs[k] = v
	}

	return defs
}

func (ac applyCommand) Synopsis() string {
	return "Apply a single run."
}

func (ac applyCommand) Help() string {
	return ac.HelpApply()
}

// HelpApply prints the help string for the 'apply' command.
func (ac applyCommand) HelpApply() string {
	defs := buildCommonRunOptionDefs()
	buildCommonApplyOptionDefs(defs)

	return fmt.Sprintf(`
Usage: %s [global options] apply [options] <workspace>

   The run apply command applies a run. Supports setting
   run-scoped Terraform and environment variables, auto-
   approving any changes to the infrastructure, running
   from remote module sources and more.

%s


Combining --tf-var or --env-var and --tf-var-file or --env-var-file is not allowed.

`, ac.meta.BinaryName, buildHelpText(defs))
}

// The End.

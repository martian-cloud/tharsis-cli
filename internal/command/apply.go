// Package command contains the logic for processing
// all the commands and subcommands. It uses the Tharsis
// SDK to interface with Tharsis API's remote Terraform
// backend.
package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
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

	client, err := ac.meta.GetSDKClient()
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
	autoApprove, err := getBoolOptionValue("auto-approve", "false", cmdOpts)
	if err != nil {
		ac.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	inputRequired, err := getBoolOptionValue("input", "true", cmdOpts)
	if err != nil {
		ac.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
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
		ac.meta.UI.Error(output.FormatError("failed to parse boolean value for -refresh option", err))
		return 1
	}
	refreshOnly, err := getBoolOptionValue("refresh-only", "false", cmdOpts)
	if err != nil {
		ac.meta.UI.Error(output.FormatError("failed to parse boolean value for -refresh-only option", err))
		return 1
	}

	// Error is already logged.
	if !isNamespacePathValid(ac.meta, workspacePath) {
		return 1
	}

	// Do the inner plan.  Make it _NON_-speculative.
	createdRun, exitCode := createRun(ctx, client, ac.meta, &runInput{
		workspacePath:    workspacePath,
		directoryPath:    directoryPath,
		tfVarFilePath:    tfVarFiles,
		envVarFilePath:   envVarFiles,
		moduleSource:     moduleSource,
		moduleVersion:    moduleVersion,
		terraformVersion: terraformVersion,
		tfVariables:      tfVariables,
		envVariables:     envVariables,
		isDestroy:        false,
		isSpeculative:    false,
		targetAddresses:  targetAddresses,
		refresh:          refresh,
		refreshOnly:      refreshOnly,
	})
	if exitCode != 0 {
		// The error message has already been logged.
		return exitCode
	}

	return startApplyStage(ctx, comment, autoApprove, inputRequired, client, createdRun, ac.meta)
}

// startApplyStage a run after it has been planned.
func startApplyStage(ctx context.Context, comment string, autoApprove, inputRequired bool,
	client *tharsis.Client, createdRun *sdktypes.Run, meta *Metadata,
) int {
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

	lastSeenLogSize := int32(0)
	logsInput := &sdktypes.JobLogsSubscriptionInput{
		JobID:           *appliedRun.Apply.CurrentJobID,
		RunID:           appliedRun.Metadata.ID,
		WorkspacePath:   appliedRun.WorkspacePath,
		LastSeenLogSize: &lastSeenLogSize,
	}

	meta.Logger.Debugf("apply: job logs input: %#v", logsInput)

	// Subscribe to job log events so we know when to fetch new logs.
	logChannel, err := client.Job.SubscribeToJobLogs(ctx, logsInput)
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to get job logs", err))
		return 1
	}

	for {
		logEvent, ok := <-logChannel
		if !ok {
			// No more logs since channel was closed.
			break
		}

		if logEvent.Error != nil {
			// Catch any incoming errors.
			meta.Logger.Error(output.FormatError("failed to get job logs", logEvent.Error))
			return 1
		}

		meta.UI.Output(strings.TrimSpace(logEvent.Logs))
	}

	finishedRun, err := client.Run.GetRun(ctx, &sdktypes.GetRunInput{ID: createdRun.Metadata.ID})
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to get finished run", err))
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

`, ac.meta.BinaryName, buildHelpText(defs))
}

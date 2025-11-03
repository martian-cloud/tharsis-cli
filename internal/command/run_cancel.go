package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// runCancelCommand is the top-level structure for the run cancel command.
type runCancelCommand struct {
	meta *Metadata
}

// NewRunCancelCommandFactory returns a runCancelCommand struct.
func NewRunCancelCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return runCancelCommand{
			meta: meta,
		}, nil
	}
}

func (rc runCancelCommand) Run(args []string) int {
	rc.meta.Logger.Debugf("Starting the 'run cancel' command with %d arguments:", len(args))
	for ix, arg := range args {
		rc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := rc.meta.GetSDKClient()
	if err != nil {
		rc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return rc.doRunCancel(ctx, client, args)
}

func (rc runCancelCommand) doRunCancel(ctx context.Context, client *tharsis.Client, opts []string) int {
	rc.meta.Logger.Debugf("will do run cancel, %d opts", len(opts))

	defs := rc.buildRunCancelDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(rc.meta.BinaryName+" run cancel", defs, opts)
	if err != nil {
		rc.meta.Logger.Error(output.FormatError("failed to parse run cancel options", err))
		return 1
	}
	if len(cmdArgs) < 1 || cmdArgs[0] == "" {
		rc.meta.Logger.Error(output.FormatError("missing run cancel ID", nil), rc.HelpRunCancel())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive run cancel arguments: %s", cmdArgs)
		rc.meta.Logger.Error(output.FormatError(msg, nil), rc.HelpRunCancel())
		return 1
	}

	force, err := getBoolOptionValue("force", "false", cmdOpts)
	if err != nil {
		rc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	id := cmdArgs[0]

	run, err := client.Run.GetRun(ctx, &sdktypes.GetRunInput{ID: id})
	if err != nil {
		rc.meta.Logger.Error(output.FormatError("failed to get run", err))
		return 1
	}

	// Subscribe to run events and wait for event to be canceled.
	eventChan, err := client.Run.SubscribeToWorkspaceRunEvents(ctx,
		&sdktypes.RunSubscriptionInput{
			WorkspacePath: run.WorkspacePath,
			RunID:         &id,
		})
	if err != nil {
		rc.meta.Logger.Error(output.FormatError("failed subscribe to workspace run events", err))
		return 1
	}

	input := &sdktypes.CancelRunInput{RunID: id, Force: &force}
	rc.meta.Logger.Debugf("run cancel input: %#v", input)

	_, err = client.Run.CancelRun(ctx, input)
	if err != nil {
		rc.meta.Logger.Error(output.FormatError("failed to cancel a run", err))
		return 1
	}

	rc.meta.UI.Info("Run cancellation in progress...")

	// Wait for an event on eventChan.
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case eventRun := <-eventChan:
			switch eventRun.Status {
			case sdktypes.RunApplied,
				sdktypes.RunPlanned,
				sdktypes.RunPlannedAndFinished,
				sdktypes.RunErrored:
				err = fmt.Errorf("run status: %s", eventRun.Status)
			case sdktypes.RunCanceled:
				rc.meta.UI.Info("Run cancel succeeded")
				return 0
			}
		}

		if err != nil {
			rc.meta.Logger.Error(output.FormatError("failed to cancel a run", err))
			return 1
		}
	}
}

func (rc runCancelCommand) buildRunCancelDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"force": {
			Arguments: []string{},
			Synopsis:  "Force the run to cancel.",
		},
	}
}

func (rc runCancelCommand) Synopsis() string {
	return "Cancel a run."
}

func (rc runCancelCommand) Help() string {
	return rc.HelpRunCancel()
}

// HelpRunCancel produces the help string for the 'run cancel' command.
func (rc runCancelCommand) HelpRunCancel() string {
	return fmt.Sprintf(`
Usage: %s [global options] run cancel [options] <id>

   The run cancel command cancels a run. Expects the target
   run ID as an argument. Supports forced cancellation of a
   run which is useful when a graceful cancel is not enough.

%s

`, rc.meta.BinaryName, buildHelpText(rc.buildRunCancelDefs()))
}

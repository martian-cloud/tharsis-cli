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

func (wuc runCancelCommand) Run(args []string) int {
	wuc.meta.Logger.Debugf("Starting the 'run cancel' command with %d arguments:", len(args))
	for ix, arg := range args {
		wuc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := wuc.meta.ReadSettings()
	if err != nil {
		wuc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		wuc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wuc.doRunCancel(ctx, client, args)
}

func (wuc runCancelCommand) doRunCancel(ctx context.Context, client *tharsis.Client, opts []string) int {
	wuc.meta.Logger.Debugf("will do run cancel, %d opts", len(opts))

	defs := buildRunCancelDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wuc.meta.BinaryName+" run cancel", defs, opts)
	if err != nil {
		wuc.meta.Logger.Error(output.FormatError("failed to parse run cancel options", err))
		return 1
	}
	if len(cmdArgs) < 1 || cmdArgs[0] == "" {
		wuc.meta.Logger.Error(output.FormatError("missing run cancel id", nil), wuc.HelpRunCancel())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive run cancel arguments: %s", cmdArgs)
		wuc.meta.Logger.Error(output.FormatError(msg, nil), wuc.HelpRunCancel())
		return 1
	}

	force := getOption("force", "", cmdOpts)[0] == "1"
	id := cmdArgs[0]

	run, err := client.Run.GetRun(ctx, &sdktypes.GetRunInput{ID: id})
	if err != nil {
		wuc.meta.Logger.Error(output.FormatError("failed to get run", err))
		return 1
	}

	if run == nil {
		wuc.meta.Logger.Error(output.FormatError("failed to get run", nil))
		return 1
	}

	// Subscribe to run events and wait for event to be canceled.
	done := make(chan bool)
	go func() {
		defer close(done)

		eventChan, err := client.Run.SubscribeToWorkspaceRunEvents(ctx,
			&sdktypes.RunSubscriptionInput{
				WorkspacePath: run.WorkspacePath,
				RunID:         &id,
			})
		if err != nil {
			wuc.meta.Logger.Error(output.FormatError("failed subscribe to workspace run events", err))
			return
		}

		input := &sdktypes.CancelRunInput{RunID: id, Force: &force}
		wuc.meta.Logger.Debugf("run cancel input: %#v", input)

		_, err = client.Run.CancelRun(ctx, input)
		if err != nil {
			wuc.meta.Logger.Error(output.FormatError("failed to cancel a run", err))
			return
		}

		// Wait for the run to be canceled.
		for {
			eventRun := <-eventChan
			if eventRun.Status == sdktypes.RunCanceled {
				done <- true
				return
			}
		}
	}()

	// Wait for a event on channel done.
	select {
	case <-ctx.Done():
		wuc.meta.Logger.Error(output.FormatError("failed to cancel a run", ctx.Err()))
		return 1
	case completed := <-done:
		if completed {
			wuc.meta.UI.Output("run cancel succeeded.")
		}
	}

	return 0
}

func buildRunCancelDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"force": {
			Arguments: []string{},
			Synopsis:  "Force the run to cancel.",
		},
	}
}

func (wuc runCancelCommand) Synopsis() string {
	return "Cancel a run."
}

func (wuc runCancelCommand) Help() string {
	return wuc.HelpRunCancel()
}

// HelpRunCancel produces the help string for the 'run cancel' command.
func (wuc runCancelCommand) HelpRunCancel() string {
	return fmt.Sprintf(`
Usage: %s [global options] run cancel [options] <id>

   The run cancel command cancels a run. Expects the target
   run ID as an argument. Supports forced cancellation of a
   run which is useful when a graceful cancel is not enough.

%s

`, wuc.meta.BinaryName, buildHelpText(buildRunCancelDefs()))
}

// The End.

package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/run"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	runpkg "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/pkg/run"
)

// Exit codes for a process stopped by a signal: 128 + the signal number.
const (
	interruptedExitCode = 130 // SIGINT
	terminatedExitCode  = 143 // SIGTERM
)

// interruptContexts listens for Ctrl-C (SIGINT) and SIGTERM: stopCtx is canceled by the first
// signal, which stops following the run, and cancelCtx by the second, which abandons any question
// asked about the run. Each is canceled with a run.SignalError cause naming the signal, so the
// interrupt handling can tell them apart. release stops listening.
func interruptContexts(parent context.Context) (stopCtx, cancelCtx context.Context, release func()) {
	stopCtx, stop := context.WithCancelCause(parent)
	cancelCtx, cancel := context.WithCancelCause(parent)

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})

	go func() {
		for _, fire := range []context.CancelCauseFunc{stop, cancel} {
			select {
			case sig := <-signals:
				fire(&run.SignalError{Signal: sig})
			case <-done:
				return
			}
		}
	}()

	return stopCtx, cancelCtx, func() {
		signal.Stop(signals)
		close(done)
		stop(nil)
		cancel(nil)
	}
}

// signalExitCode is the exit code for following a run stopped by the signal that canceled stopCtx.
func signalExitCode(stopCtx context.Context) int {
	if run.InterruptSignal(stopCtx) == syscall.SIGTERM {
		return terminatedExitCode
	}
	return interruptedExitCode
}

// followRunAndReport follows a created run through its stages and maps how following ended to an exit code.
func followRunAndReport(stopCtx, cancelCtx context.Context, ui terminal.UI, runMgr *run.Manager, createdRun *pb.Run, opts *run.FollowOptions) int {
	outcome, err := runMgr.FollowRun(stopCtx, cancelCtx, createdRun, opts)
	if err != nil {
		ui.ErrorWithSummary(err, "failed to follow run %s", createdRun.Metadata.Id)
		return 1
	}

	code := reportOutcome(ui, outcome, opts)
	if code == interruptedExitCode {
		code = signalExitCode(stopCtx)
	}
	warnAdvisories(ui, outcome)
	return code
}

// addFollowFlags adds the flags that control how a run is followed: -input and -wait-timeout.
func addFollowFlags(f *flag.Set, input **bool, waitTimeout **time.Duration) {
	f.BoolVar(
		input,
		"input",
		"Ask for plan approval and policy gate decisions. When false, wait for them to be made in the Tharsis UI.",
		flag.Default(true),
	)
	f.DurationVar(
		waitTimeout,
		"wait-timeout",
		"How long to wait on plan approval or a policy gate decision, whether asked for at the terminal "+
			"or being made in the Tharsis UI (e.g. 30m, 2h), before leaving the run waiting and exiting. "+
			"Must be positive.",
		flag.Default(runpkg.DefaultWaitTimeout),
		flag.Validate(func(s string) error {
			d, err := time.ParseDuration(s)
			if err != nil {
				return err
			}
			if d <= 0 {
				return errors.New("must be a positive duration")
			}
			return nil
		}),
	)
}

// followHelp describes, for a command's help, how the run it creates is followed. hasApply is false
// for a speculative plan, which has no plan approval or apply phase.
func followHelp(command string, hasApply bool) string {
	var b strings.Builder
	if hasApply {
		fmt.Fprintf(&b, `   Each stage of the run is shown as it runs, including any
   pre-plan, post-plan, pre-apply and post-apply policy
   stages with their policy findings.

   Plan approval: enter 'yes' to apply, 'no' to discard the
   run, or press Enter to leave it waiting. Approval is also
   read from piped stdin, so 'yes | tharsis %s' approves.
`, command)
	} else {
		b.WriteString(`   Each stage of the run is shown as it runs, including any
   pre-plan and post-plan policy stages with their policy
   findings.
`)
	}
	b.WriteString(`
   Policy gates: when a policy soft-fails, the run waits for
   approval. Approve it (or override it, when it has no
   eligible approvers), discard the run, or press Enter to
   leave it waiting. Gates are only asked about at an
   interactive terminal.

   With -input=false, or when the CLI can't ask, it keeps
   following the run and waits for the decision to be made
   in the Tharsis UI. A decision is waited on for at most
   -wait-timeout, after which the run is left waiting.

   Ctrl-C at an interactive terminal asks whether to cancel
   the run if it is still running. Otherwise, including on
   SIGTERM, without a terminal and with -input=false, the
   CLI stops following and leaves the run going; cancel it
   in the Tharsis UI if needed. A second interrupt exits
   immediately.
`)
	return b.String()
}

// warnAdvisories ends the output with a reminder of any advisory policies that failed, since they
// don't stop the run and are easy to miss further up.
func warnAdvisories(ui terminal.UI, outcome *runpkg.Outcome) {
	n := len(outcome.Advisories)
	if n == 0 {
		return
	}
	noun := "policies"
	if n == 1 {
		noun = "policy"
	}
	ui.Output("")
	ui.Warnf("%d advisory %s failed on run %s:", n, noun, outcome.Run.Metadata.Id)
	for _, name := range outcome.Advisories {
		ui.Warnf("  %s", name)
	}
}

// reportOutcome prints how following the run ended and returns the exit code for it.
func reportOutcome(ui terminal.UI, outcome *runpkg.Outcome, opts *run.FollowOptions) int {
	runID := outcome.Run.Metadata.Id
	status := strings.ToLower(outcome.Run.Status.String())

	switch outcome.Kind {
	case runpkg.OutcomeCompleted:
		return 0
	case runpkg.OutcomeDetached:
		waitingOn := "a decision on its " + strings.ToLower(outcome.Stage.Name) + ". Approve, override or discard it"
		if outcome.Stage.Kind == runpkg.StagePlanApproval {
			waitingOn = "its plan to be approved. Apply or discard it"
		}
		if outcome.TimedOut {
			timeout := opts.WaitTimeout
			if timeout <= 0 {
				timeout = runpkg.DefaultWaitTimeout
			}
			ui.Warnf("\nStopped waiting after %s. Run %s is waiting for %s in the Tharsis UI.",
				timeout, runID, waitingOn)
		} else {
			ui.Warnf("\nRun %s is waiting for %s in the Tharsis UI.", runID, waitingOn)
		}
		return 1
	case runpkg.OutcomeDiscarded:
		ui.Output("Run discarded.")
		return 0
	case runpkg.OutcomeInterrupted:
		switch {
		case outcome.RunCanceled:
			ui.Output("\nRun %s was canceled.", runID)
		case outcome.RunDiscarded:
			ui.Output("\nRun %s was discarded.", runID)
		default:
			ui.Output("\nStopped following run %s (%s). The run continues in Tharsis.", runID, status)
		}
		return interruptedExitCode
	default:
		switch {
		case outcome.Run.Status == pb.RunStatus_DISCARDED:
			ui.Errorf("Run %s was discarded", runID)
		case outcome.Run.Status == pb.RunStatus_CANCELED:
			ui.Errorf("Run %s was canceled", runID)
		case outcome.Stage.Name != "":
			ui.Errorf("Run %s failed during %s", runID, strings.ToLower(outcome.Stage.Name))
		default:
			ui.Errorf("Run %s ended with status %s", runID, status)
		}
		return 1
	}
}

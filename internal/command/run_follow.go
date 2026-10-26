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

// interruptContexts listens for Ctrl-C (SIGINT) and SIGTERM: followCtx is canceled by the first
// signal, which stops following the run, and cleanupCtx by the second, which abandons any question
// asked about the run. Each is canceled with a run.SignalError cause naming the signal, so the
// interrupt handling can tell them apart. release stops listening.
func interruptContexts(parent context.Context) (followCtx, cleanupCtx context.Context, release func()) {
	followCtx, stop := context.WithCancelCause(parent)
	cleanupCtx, cancel := context.WithCancelCause(parent)

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

	return followCtx, cleanupCtx, func() {
		signal.Stop(signals)
		close(done)
		stop(nil)
		cancel(nil)
	}
}

// signalExitCode is the exit code for following a run stopped by the signal that canceled followCtx.
func signalExitCode(followCtx context.Context) int {
	if run.InterruptSignal(followCtx) == syscall.SIGTERM {
		return terminatedExitCode
	}
	return interruptedExitCode
}

// followRunAndReport follows a created run through its stages and maps how following ended to an exit code.
func followRunAndReport(followCtx, cleanupCtx context.Context, ui terminal.UI, runMgr *run.Manager, createdRun *pb.Run, opts *run.FollowOptions) int {
	outcome, err := runMgr.FollowRun(followCtx, cleanupCtx, createdRun, opts)
	if err != nil {
		ui.ErrorWithSummary(err, "failed to follow run %s", createdRun.Metadata.Id)
		return 1
	}

	code := reportOutcome(ui, outcome, opts, signalExitCode(followCtx))
	warnAdvisories(ui, outcome)
	return code
}

// addFollowFlags adds the flags that control how a run is followed: -input, -wait and -wait-timeout.
func addFollowFlags(f *flag.Set, input, wait **bool, waitTimeout **time.Duration) {
	f.BoolVar(
		input,
		"input",
		"Ask for plan approval and policy gate decisions. When false, they are left to be made in the Tharsis UI.",
		flag.Default(true),
	)
	f.BoolVar(
		wait,
		"wait",
		"When a plan approval or policy gate decision can't be asked for (-input=false or no terminal), "+
			"keep following the run and wait for it to be made in the Tharsis UI, for up to -wait-timeout. "+
			"Without it, the CLI leaves the run waiting and exits.",
		flag.Default(false),
	)
	f.DurationVar(
		waitTimeout,
		"wait-timeout",
		"How long to wait on plan approval or a policy gate decision, whether asked for at the terminal "+
			"or (with -wait) being made in the Tharsis UI (e.g. 30m, 2h), before leaving the run waiting "+
			"and exiting 1. Must be positive.",
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
   interactive terminal. When a stage is blocked on a check
   this CLI can't show, there is nothing to ask; at an
   interactive terminal the CLI waits for it to be decided
   in the Tharsis UI, for up to -wait-timeout.
`)
	if hasApply {
		b.WriteString(`
   With -input=false, or when the CLI can't ask, it leaves
   the run waiting for the decision to be made in the
   Tharsis UI and exits: 0 at plan approval, 1 at a policy
   gate. With -wait, it keeps following the run while the
   decision is made in the Tharsis UI instead. A decision is
   waited on for at most -wait-timeout, after which the run
   is left waiting and the CLI exits 1.
`)
	} else {
		b.WriteString(`
   With -input=false, or when the CLI can't ask, it leaves
   the run waiting at the policy gate for the decision to
   be made in the Tharsis UI and exits 1. With -wait, it
   keeps following the run while the decision is made in
   the Tharsis UI instead. A decision is waited on for at
   most -wait-timeout, after which the run is left waiting
   and the CLI exits 1.
`)
	}
	b.WriteString(`
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

// reportOutcome prints how following the run ended and returns the exit code for it; interruptedCode
// is the code for following stopped by a signal.
func reportOutcome(ui terminal.UI, outcome *runpkg.Outcome, opts *run.FollowOptions, interruptedCode int) int {
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
			return 1
		}
		if outcome.LeaveChosen {
			// The user chose to leave the run waiting, so this isn't a failure.
			ui.Warnf("\nRun %s is waiting for %s in the Tharsis UI.", runID, waitingOn)
			return 0
		}
		// The CLI couldn't ask and -wait wasn't passed. Stopping at plan approval is the expected way
		// to hand the run over to the Tharsis UI; stopping at a policy gate means the run didn't get
		// as far as the command was asked to take it.
		if outcome.LeaveReason != "" {
			ui.Output("\nNot waiting for a decision since %s; pass -wait to wait for it.", outcome.LeaveReason)
		} else {
			ui.Output("\nNot waiting for a decision; pass -wait to wait for it.")
		}
		ui.Warnf("\nRun %s is waiting for %s in the Tharsis UI.", runID, waitingOn)
		if outcome.Stage.Kind == runpkg.StagePlanApproval {
			return 0
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
		case runpkg.IsRunFinal(outcome.Run.Status):
			ui.Output("\nStopped following run %s. The run already ended (%s).", runID, status)
		default:
			ui.Output("\nStopped following run %s (%s). The run continues in Tharsis.", runID, status)
		}
		return interruptedCode
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

package run

import (
	"context"
	"errors"
	"io"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	runpkg "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/pkg/run"
)

// cliDecider asks the user at the terminal for plan approval and gate decisions. When it can't ask
// (-input=false, no terminal, or empty stdin), it leaves the run waiting for the decision to be made
// in the Tharsis UI, and stops following it unless wait is set.
type cliDecider struct {
	ui       outputUI
	prompter *Prompter
	// wait is true when -wait was passed: keep following the run while the decision is made
	// elsewhere, rather than stopping.
	wait bool
}

// outputUI is the part of the terminal UI the hooks write to.
type outputUI interface {
	Output(msg string, raw ...interface{})
}

var _ runpkg.Decider = (*cliDecider)(nil)

// DecidePlan asks whether to apply the plan. Plan approval reads piped stdin too, so
// `yes | tharsis apply` still approves.
func (d *cliDecider) DecidePlan(ctx context.Context, req *runpkg.PlanRequest) (runpkg.PlanDecision, error) {
	if !d.prompter.AcceptsInput() {
		return d.cannotAskPlan(d.prompter.cannotAskReason()), nil
	}

	question, verb := "Do you approve to apply the above plan?", "apply"
	if req.IsDestroy {
		question, verb = "Do you approve to destroy the above resources?", "destroy"
	}
	d.ui.Output("\n%s\n", question)
	prompt := "  Enter 'yes' to " + verb + ", 'no' to discard the run, or press Enter to leave it waiting: "

	for {
		answer, err := d.prompter.Ask(ctx, prompt)
		switch {
		case errors.Is(err, io.EOF):
			// Stdin is empty, as it often is in CI.
			return d.cannotAskPlan("no answer could be read from stdin"), nil
		case errors.Is(err, terminal.ErrNonInteractive):
			// The UI can't read at all (stdout isn't a terminal).
			return d.cannotAskPlan("the terminal is not interactive"), nil
		case err != nil:
			return runpkg.PlanDecision{}, err
		}

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "yes":
			return runpkg.PlanDecision{Action: runpkg.PlanApply}, nil
		case "no":
			if req.IsDestroy {
				d.ui.Output("Approval response was negative. Will not destroy resources.")
			} else {
				d.ui.Output("Approval response was negative. Will not apply the plan.")
			}
			return runpkg.PlanDecision{Action: runpkg.PlanDiscard}, nil
		case "":
			return runpkg.PlanDecision{Action: runpkg.PlanLeave}, nil
		default:
			d.ui.Output("  %q isn't one of the choices.", answer)
		}
	}
}

// DecideGates asks what to do at a blocked policy stage. Gates are only asked about at an
// interactive terminal.
func (d *cliDecider) DecideGates(ctx context.Context, req *runpkg.GateRequest) (runpkg.GateDecision, error) {
	if !d.prompter.Interactive() {
		return d.cannotAskGates(d.prompter.cannotAskReason()), nil
	}

	d.ui.Output("")
	prompt := gatePrompt(req.CanApprove, req.CanOverride)
	for {
		answer, err := d.prompter.Ask(ctx, prompt)
		switch {
		case errors.Is(err, io.EOF):
			return d.cannotAskGates("no answer could be read from stdin"), nil
		case errors.Is(err, terminal.ErrNonInteractive):
			return d.cannotAskGates("the terminal is not interactive"), nil
		case err != nil:
			return runpkg.GateDecision{}, err
		}

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "approve":
			if req.CanApprove {
				return runpkg.GateDecision{Action: runpkg.GateApprove}, nil
			}
			d.ui.Output("  You can't approve this gate.")
		case "override":
			if req.CanOverride {
				return runpkg.GateDecision{Action: runpkg.GateOverride}, nil
			}
			d.ui.Output("  You can't override this gate.")
		case "discard":
			return runpkg.GateDecision{Action: runpkg.GateDiscard}, nil
		case "":
			return runpkg.GateDecision{Action: runpkg.GateLeave}, nil
		default:
			d.ui.Output("  %q isn't one of the choices.", answer)
		}
	}
}

// cannotAskPlan is the plan decision when the user can't be asked: wait for it to be made elsewhere
// with -wait, or else leave the run waiting.
func (d *cliDecider) cannotAskPlan(reason string) runpkg.PlanDecision {
	if d.wait {
		return runpkg.PlanDecision{Action: runpkg.PlanWaitElsewhere, Reason: reason}
	}
	return runpkg.PlanDecision{Action: runpkg.PlanLeave, Reason: reason}
}

// cannotAskGates is the gate decision when the user can't be asked: wait for it to be made
// elsewhere with -wait, or else leave the run waiting.
func (d *cliDecider) cannotAskGates(reason string) runpkg.GateDecision {
	if d.wait {
		return runpkg.GateDecision{Action: runpkg.GateWaitElsewhere, Reason: reason}
	}
	return runpkg.GateDecision{Action: runpkg.GateLeave, Reason: reason}
}

// gatePrompt lists the choices at a gate, offering approve and override only where they apply.
func gatePrompt(canApprove, canOverride bool) string {
	switch {
	case canApprove && canOverride:
		return "  Enter 'approve' to approve, 'override' to override, 'discard' to discard the run, or press Enter to leave it waiting: "
	case canApprove:
		return "  Enter 'approve' to approve, 'discard' to discard the run, or press Enter to leave it waiting: "
	case canOverride:
		return "  Enter 'override' to override, 'discard' to discard the run, or press Enter to leave it waiting: "
	default:
		return "  Enter 'discard' to discard the run, or press Enter to leave it waiting: "
	}
}

// cliInterrupter decides what happens to the run on Ctrl-C.
//
// At an interactive terminal the user is asked whether to cancel a run that is still doing work; a
// run waiting on a person (plan approval or a policy gate) is left for them to decide on.
//
// Otherwise the run is always left going: the CLI never cancels or discards a run without being told
// to. SIGTERM in particular is sent for many reasons that aren't a wish to stop the run (a CI job's
// time limit, a node drain, a runner restart), and canceling an apply part-way leaves infrastructure
// half changed, while a run left going finishes on its own and can be canceled in the Tharsis UI.
type cliInterrupter struct {
	ui       outputUI
	prompter *Prompter
	// followCtx is the context the interrupt canceled; its cause says which signal it was.
	followCtx context.Context
}

var _ runpkg.Interrupter = (*cliInterrupter)(nil)

func (i *cliInterrupter) OnInterrupt(ctx context.Context, run *pb.Run) runpkg.InterruptAction {
	if !runpkg.IsRunCancelable(run.Status) {
		return runpkg.InterruptLeave
	}
	switch {
	case isTerminated(i.followCtx):
		i.ui.Output("Not canceling the run since the CLI was terminated (SIGTERM).")
		return runpkg.InterruptLeave
	case !i.prompter.Terminal():
		i.ui.Output("Not canceling the run since the terminal is not interactive.")
		return runpkg.InterruptLeave
	case !i.prompter.Interactive():
		i.ui.Output("Not asking whether to cancel the run since -input was false.")
		return runpkg.InterruptLeave
	}

	i.ui.Output("\nDo you want to cancel run %s?\n", run.Metadata.Id)
	answer, err := i.prompter.Ask(ctx, "  only 'yes' will be accepted: ")
	if err != nil {
		if ctx.Err() != nil {
			i.ui.Output("\nTwo interrupts received. Exiting immediately.")
		}
		return runpkg.InterruptLeave
	}
	if strings.ToLower(strings.TrimSpace(answer)) != "yes" {
		i.ui.Output("The run was not canceled.")
		return runpkg.InterruptLeave
	}
	return runpkg.InterruptCancelRunIfInProgress
}

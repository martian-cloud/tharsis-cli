package run

import (
	"strings"

	"github.com/fatih/color"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	runpkg "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/pkg/run"
)

// Colors for policy findings. These color only the symbol or verdict rather than using the UI's
// message styles, which prefix the whole line ("warning: ", "! ") in non-interactive output.
var (
	colorPassed = color.New(color.FgGreen).SprintFunc()
	colorFailed = color.New(color.FgRed).SprintFunc()
	colorWarned = color.New(color.FgYellow).SprintFunc()
	colorFaint  = color.New(color.Faint).SprintFunc()
	colorBold   = color.New(color.Bold).SprintFunc()
)

// renderer shows a followed run at the terminal: it is the CLI's runpkg.Observer.
type renderer struct {
	ui    terminal.UI
	runID string
}

var _ runpkg.Observer = (*renderer)(nil)

// Logs prints a chunk of a job's logs, without color when color is off.
func (r *renderer) Logs(_ runpkg.StageRef, chunk string) error {
	logs := strings.TrimSpace(chunk)
	if color.NoColor {
		logs = terminal.StripAnsi(logs)
	}
	if logs != "" {
		r.ui.Output(logs)
	}
	return nil
}

// Event prints an event.
func (r *renderer) Event(e runpkg.Event) {
	switch e := e.(type) {
	case runpkg.StageStarted:
		title := e.Stage.Name
		if e.Retry {
			title += " (retry)"
		}
		r.ui.Output(title, terminal.WithHeaderStyle())
	case runpkg.Waiting:
		if msg := waitingMessage(e); msg != "" {
			r.ui.Output(msg)
		}
	case runpkg.PolicyCheckFinished:
		r.beginCheck(e.First)
		r.policyCheck(e.Results)
	case runpkg.UnsupportedPolicyCheck:
		r.beginCheck(e.First)
		r.ui.Output("  %s Unsupported policy check type %s; waiting for it to finish.",
			colorFaint("○"), runpkg.CheckTypeLabel(e.CheckType))
	case runpkg.NoVisibleTasks:
		r.ui.Output("\n%s This stage has no tasks this version of the CLI can show; waiting for it to finish.", colorFaint("○"))
	case runpkg.GatesPending:
		for _, g := range e.Gates {
			if g.NeedsOverride {
				r.ui.Output("\nOverride required to continue.")
				continue
			}
			r.ui.Output("\nApproval required to continue:")
			for _, line := range runpkg.GateRuleLines(g.Gate, g.Results) {
				r.ui.Output("  %s", line)
			}
		}
	case runpkg.DecisionDeferred:
		what := "a decision"
		if e.Stage.Kind == runpkg.StagePlanApproval {
			what = "approval"
		}
		if e.Reason != "" {
			r.ui.Output("\nNot asking for %s since %s.", what, e.Reason)
		}
	case runpkg.DecisionAbandoned:
		r.ui.Output("\nThe run moved on before an answer was given.")
	case runpkg.GateApproved:
		if e.AllApprovalsIn {
			r.ui.Output("\nApproval recorded; all required approvals are in.")
		} else {
			r.ui.Output("\nApproval recorded.")
		}
	case runpkg.GateOverridden:
		r.ui.Output("\nGate overridden.")
	case runpkg.GateDenied:
		if e.Action == runpkg.GateOverride {
			r.ui.Output("  %s", colorFailed("You don't have permission to override this gate."))
		} else {
			r.ui.Output("  %s", colorFailed("You are not an eligible approver for this gate."))
		}
	case runpkg.ActionConflict:
		if e.Action == "discard" {
			r.ui.Output("The run can no longer be discarded; it has moved on.")
		} else {
			r.ui.Output("The run can no longer be %s; it has moved on.", e.Action)
		}
	case runpkg.NoChanges:
		r.ui.Output("\nStopping since plan had no changes.")
	case runpkg.AutoApproving:
		r.ui.Output("\nAuto-approving.\n")
	case runpkg.Interrupted:
		r.ui.Output("\nInterrupt received.")
	case runpkg.InterruptFailed:
		r.ui.Errorf("Failed to %s run %s: %v", e.Action, r.runID, e.Err)
	}
}

// waitingMessage describes what the run is waiting on, for the waiting line.
func waitingMessage(w runpkg.Waiting) string {
	switch w.Reason {
	case runpkg.WaitRunStarting:
		return "Waiting on run to start"
	case runpkg.WaitWorkspace:
		return "Waiting for the workspace to be free"
	case runpkg.WaitRunner:
		return "Waiting for a runner"
	case runpkg.WaitTasksRunning:
		return "Running stage tasks"
	case runpkg.WaitApplyStart:
		return "Waiting for the apply to start..."
	case runpkg.WaitOtherApprovers:
		return "Waiting for other approvers (" + w.Detail + ")"
	case runpkg.WaitDecisionElsewhere:
		if w.Stage.Kind == runpkg.StagePlanApproval {
			return "Waiting for the plan to be applied or the run discarded in the Tharsis UI"
		}
		msg := "Waiting for the gate to be approved, overridden or the run discarded in the Tharsis UI"
		if w.Detail != "" {
			msg += " (" + w.Detail + ")"
		}
		return msg
	case runpkg.WaitUnsupportedTask:
		return "Waiting for the run to continue; it is blocked on a task this version of the CLI can't show"
	default:
		return ""
	}
}

// beginCheck starts a check's output: the first check of an evaluation opens the stage's policy
// checks section, and each later one is set off from the one before it.
func (r *renderer) beginCheck(first bool) {
	if first {
		r.ui.Output("\n%s", colorBold("Policy checks"))
	} else {
		r.ui.Output("")
	}
}

// policyCheck prints a policy check's verdict and each policy's result with the violation messages
// it reported, indented to sit under its stage's policy checks heading, for example:
//
//	! OPA policies failed; approval required
//	  ✓ acme/required-tags (hard-mandatory)
//	  ✗ acme/encrypted-buckets (soft-mandatory)
//	      - bucket must be encrypted
func (r *renderer) policyCheck(results *pb.PolicyCheckResults) {
	check := results.GetPolicyCheck()
	name := runpkg.PolicyCheckTypeName(check.GetCheckType())

	switch check.GetStatus() {
	case pb.PolicyCheckStatus_POLICY_CHECK_STATUS_PASSED:
		// Advisory policies can fail without failing the check; say so rather than claim a clean pass.
		if n := runpkg.FailedPolicyCount(results); n > 0 {
			r.ui.Output("  %s %s passed with %d advisory %s", colorWarned("!"), name, n, runpkg.Plural(n, "failure", "failures"))
		} else {
			r.ui.Output("  %s %s passed", colorPassed("✓"), name)
		}
	case pb.PolicyCheckStatus_POLICY_CHECK_STATUS_SOFT_FAILED:
		r.ui.Output("  %s %s failed; approval required", colorWarned("!"), name)
	case pb.PolicyCheckStatus_POLICY_CHECK_STATUS_OVERRIDDEN:
		r.ui.Output("  %s %s failed; overridden", colorWarned("!"), name)
	case pb.PolicyCheckStatus_POLICY_CHECK_STATUS_ERRORED:
		r.ui.Output("  %s %s failed", colorFailed("✗"), name)
	default:
		r.ui.Output("  %s %s finished", colorFaint("○"), name)
	}

	for _, pr := range results.GetPolicyResults() {
		p := pr.GetPolicy()
		level := colorFaint("(" + runpkg.EnforcementLevelName(p.GetEnforcementLevel()) + ")")
		switch p.GetStatus() {
		case pb.PolicyCheckPolicyStatus_POLICY_CHECK_POLICY_STATUS_PASSED:
			r.ui.Output("    %s %s %s", colorPassed("✓"), runpkg.PolicyName(p), level)
		case pb.PolicyCheckPolicyStatus_POLICY_CHECK_POLICY_STATUS_FAILED:
			r.ui.Output("    %s %s %s", failedPolicySymbol(p.GetEnforcementLevel()), runpkg.PolicyName(p), level)
		default:
			r.ui.Output("    %s %s %s %s", colorFaint("○"), runpkg.PolicyName(p), level, colorFaint("not evaluated"))
		}
		for _, msg := range pr.GetMessages() {
			r.ui.Output("        - %s", msg)
		}
	}
}

func failedPolicySymbol(level pb.PolicyEnforcementLevel) string {
	if level == pb.PolicyEnforcementLevel_POLICY_ENFORCEMENT_LEVEL_HARD_MANDATORY {
		return colorFailed("✗")
	}
	return colorWarned("✗")
}

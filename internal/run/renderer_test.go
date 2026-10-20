package run

import (
	"errors"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/runtest"
	runpkg "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/pkg/run"
)

// errorUI also records what is written as an error.
type errorUI struct {
	*fakeUI
	errors []string
}

func (u *errorUI) Errorf(msg string, raw ...interface{}) {
	u.errors = append(u.errors, fmt.Sprintf(msg, raw...))
}

// render sends the events to a renderer and returns what it printed, with color off.
func render(t *testing.T, events ...runpkg.Event) string {
	t.Helper()
	noColor := color.NoColor
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = noColor })

	ui := newFakeUI(true)
	r := &renderer{ui: ui, runID: "run-1"}
	for _, e := range events {
		r.Event(e)
	}
	return ui.output()
}

func TestRenderer_Logs(t *testing.T) {
	renderLogs := func(noColor bool, chunk string) string {
		prev := color.NoColor
		color.NoColor = noColor
		defer func() { color.NoColor = prev }()

		ui := newFakeUI(true)
		r := &renderer{ui: ui}
		assert.NoError(t, r.Logs(runpkg.StageRef{}, chunk))
		return ui.output()
	}

	assert.Equal(t, "hello", renderLogs(true, "\n  hello\n\n"), "surrounding whitespace is trimmed")
	assert.Equal(t, "hello", renderLogs(true, "\x1b[32mhello\x1b[0m"), "ANSI codes are stripped without color")
	assert.Equal(t, "\x1b[32mhello\x1b[0m", renderLogs(false, "\x1b[32mhello\x1b[0m"), "ANSI codes are kept with color")
	assert.Empty(t, renderLogs(true, " \n "), "an empty chunk prints nothing")
}

func TestRenderer_Events(t *testing.T) {
	plan := runpkg.StageRef{Name: "Plan", Kind: runpkg.StagePlan}
	approval := runpkg.StageRef{Name: "Plan approval", Kind: runpkg.StagePlanApproval}
	preApply := runpkg.StageRef{Name: "Pre-apply stage", Kind: runpkg.StageTask, TaskStage: runtest.PreApply}

	tests := []struct {
		name  string
		event runpkg.Event
		want  string
	}{
		{"stage started", runpkg.StageStarted{Stage: plan}, "Plan"},
		{"stage retried", runpkg.StageStarted{Stage: plan, Retry: true}, "Plan (retry)"},
		{"apply start", runpkg.Waiting{Reason: runpkg.WaitApplyStart}, "Waiting for the apply to start..."},
		{"decision made elsewhere at plan approval", runpkg.Waiting{Stage: approval, Reason: runpkg.WaitDecisionElsewhere},
			"Waiting for the plan to be applied or the run discarded in the Tharsis UI"},
		{"no visible tasks", runpkg.NoVisibleTasks{Stage: preApply},
			"\n○ This stage has no tasks this version of the CLI can show; waiting for it to finish."},
		{"unsupported check", runpkg.UnsupportedPolicyCheck{Stage: preApply, CheckType: pb.PolicyCheckType(99), First: true},
			"\nPolicy checks\n  ○ Unsupported policy check type (99); waiting for it to finish."},
		{"override required", runpkg.GatesPending{Gates: []runpkg.GateInfo{{NeedsOverride: true}}}, "\nOverride required to continue."},
		{"deferred plan approval", runpkg.DecisionDeferred{Stage: approval, Reason: "input is disabled"}, "\nNot asking for approval since input is disabled."},
		{"deferred gate decision", runpkg.DecisionDeferred{Stage: preApply, Reason: "input is disabled"}, "\nNot asking for a decision since input is disabled."},
		{"deferred without a reason", runpkg.DecisionDeferred{Stage: approval}, ""},
		{"abandoned", runpkg.DecisionAbandoned{}, "\nThe run moved on before an answer was given."},
		{"approved, more needed", runpkg.GateApproved{}, "\nApproval recorded."},
		{"approved, all in", runpkg.GateApproved{AllApprovalsIn: true}, "\nApproval recorded; all required approvals are in."},
		{"overridden", runpkg.GateOverridden{}, "\nGate overridden."},
		{"override denied", runpkg.GateDenied{Action: runpkg.GateOverride}, "  You don't have permission to override this gate."},
		{"approval denied", runpkg.GateDenied{Action: runpkg.GateApprove}, "  You are not an eligible approver for this gate."},
		{"discard conflict", runpkg.ActionConflict{Action: "discard"}, "The run can no longer be discarded; it has moved on."},
		{"other conflict", runpkg.ActionConflict{Action: "applied"}, "The run can no longer be applied; it has moved on."},
		{"no changes", runpkg.NoChanges{}, "\nStopping since plan had no changes."},
		{"auto-approving", runpkg.AutoApproving{}, "\nAuto-approving.\n"},
		{"interrupted", runpkg.Interrupted{}, "\nInterrupt received."},
		{"a waiting reason with nothing to say", runpkg.Waiting{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, render(t, tt.event))
		})
	}

	t.Run("policy checks are separated, the first opening the section", func(t *testing.T) {
		passed := runtest.SoftFailedResults("c1", 1, 0)
		passed.PolicyCheck.Status = pb.PolicyCheckStatus_POLICY_CHECK_STATUS_PASSED
		passed.PolicyResults = nil

		out := render(t,
			runpkg.PolicyCheckFinished{Stage: preApply, Results: passed, First: true},
			runpkg.PolicyCheckFinished{Stage: preApply, Results: passed})

		assert.Equal(t, "\nPolicy checks\n  ✓ OPA policies passed\n\n  ✓ OPA policies passed", out)
	})

	t.Run("a failed interrupt action is an error", func(t *testing.T) {
		ui := &errorUI{fakeUI: newFakeUI(true)}
		r := &renderer{ui: ui, runID: "run-1"}

		r.Event(runpkg.InterruptFailed{Action: "cancel", Err: errors.New("boom")})

		assert.Equal(t, []string{"Failed to cancel run run-1: boom"}, ui.errors)
	})

	t.Run("an event it doesn't show prints nothing", func(t *testing.T) {
		assert.Empty(t, render(t, runpkg.GatesPending{}))
	})
}

func TestRenderer_GatesPendingApproval(t *testing.T) {
	results := runtest.SoftFailedResults("c1", 2, 1)

	out := render(t, runpkg.GatesPending{Gates: []runpkg.GateInfo{{Gate: results.RunGate, Results: results}}})

	assert.Equal(t, "\nApproval required to continue:\n  group/encrypted-buckets  1 of 2 approvals", out)
}

func TestWaitingMessage(t *testing.T) {
	task := runpkg.StageRef{Kind: runpkg.StageTask}

	tests := []struct {
		name string
		w    runpkg.Waiting
		want string
	}{
		{"run starting", runpkg.Waiting{Reason: runpkg.WaitRunStarting}, "Waiting on run to start"},
		{"workspace", runpkg.Waiting{Reason: runpkg.WaitWorkspace}, "Waiting for the workspace to be free"},
		{"runner", runpkg.Waiting{Reason: runpkg.WaitRunner}, "Waiting for a runner"},
		{"tasks running", runpkg.Waiting{Reason: runpkg.WaitTasksRunning}, "Running stage tasks"},
		{"other approvers", runpkg.Waiting{Reason: runpkg.WaitOtherApprovers, Detail: "1 of 2 approvals"}, "Waiting for other approvers (1 of 2 approvals)"},
		{"gate decided elsewhere", runpkg.Waiting{Stage: task, Reason: runpkg.WaitDecisionElsewhere},
			"Waiting for the gate to be approved, overridden or the run discarded in the Tharsis UI"},
		{"gate decided elsewhere with detail", runpkg.Waiting{Stage: task, Reason: runpkg.WaitDecisionElsewhere, Detail: "1 of 2 approvals"},
			"Waiting for the gate to be approved, overridden or the run discarded in the Tharsis UI (1 of 2 approvals)"},
		{"unsupported task", runpkg.Waiting{Reason: runpkg.WaitUnsupportedTask},
			"Waiting for the run to continue; it is blocked on a task this version of the CLI can't show"},
		{"no reason", runpkg.Waiting{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, waitingMessage(tt.w))
		})
	}
}

func TestRenderer_PolicyCheck(t *testing.T) {
	const (
		hard     = pb.PolicyEnforcementLevel_POLICY_ENFORCEMENT_LEVEL_HARD_MANDATORY
		soft     = pb.PolicyEnforcementLevel_POLICY_ENFORCEMENT_LEVEL_SOFT_MANDATORY
		advisory = pb.PolicyEnforcementLevel_POLICY_ENFORCEMENT_LEVEL_ADVISORY
	)
	policy := func(id string, level pb.PolicyEnforcementLevel, s pb.PolicyCheckPolicyStatus, messages ...string) *pb.PolicyCheckPolicyResult {
		return &pb.PolicyCheckPolicyResult{
			Policy:   &pb.PolicyCheckPolicy{Id: id, EnforcementLevel: level, Status: s},
			Messages: messages,
		}
	}
	const (
		policyPassed = pb.PolicyCheckPolicyStatus_POLICY_CHECK_POLICY_STATUS_PASSED
		policyFailed = pb.PolicyCheckPolicyStatus_POLICY_CHECK_POLICY_STATUS_FAILED
	)
	check := func(s pb.PolicyCheckStatus, results ...*pb.PolicyCheckPolicyResult) *pb.PolicyCheckResults {
		return &pb.PolicyCheckResults{
			PolicyCheck:   &pb.PolicyCheck{Status: s, CheckType: pb.PolicyCheckType_POLICY_CHECK_TYPE_OPA},
			PolicyResults: results,
		}
	}

	tests := []struct {
		name    string
		results *pb.PolicyCheckResults
		want    string
	}{
		{"passed", check(pb.PolicyCheckStatus_POLICY_CHECK_STATUS_PASSED, policy("p1", hard, policyPassed)),
			"  ✓ OPA policies passed\n    ✓ p1 (hard-mandatory)"},
		{"passed with advisory failures", check(pb.PolicyCheckStatus_POLICY_CHECK_STATUS_PASSED, policy("p1", advisory, policyFailed, "tag missing")),
			"  ! OPA policies passed with 1 advisory failure\n    ✗ p1 (advisory)\n        - tag missing"},
		{"passed with several advisory failures", check(pb.PolicyCheckStatus_POLICY_CHECK_STATUS_PASSED,
			policy("p1", advisory, policyFailed), policy("p2", advisory, policyFailed)),
			"  ! OPA policies passed with 2 advisory failures\n    ✗ p1 (advisory)\n    ✗ p2 (advisory)"},
		{"soft failed", check(pb.PolicyCheckStatus_POLICY_CHECK_STATUS_SOFT_FAILED, policy("p1", soft, policyFailed, "bucket must be encrypted")),
			"  ! OPA policies failed; approval required\n    ✗ p1 (soft-mandatory)\n        - bucket must be encrypted"},
		{"overridden", check(pb.PolicyCheckStatus_POLICY_CHECK_STATUS_OVERRIDDEN, policy("p1", soft, policyFailed)),
			"  ! OPA policies failed; overridden\n    ✗ p1 (soft-mandatory)"},
		{"errored", check(pb.PolicyCheckStatus_POLICY_CHECK_STATUS_ERRORED, policy("p1", hard, policyFailed)),
			"  ✗ OPA policies failed\n    ✗ p1 (hard-mandatory)"},
		{"another status", check(pb.PolicyCheckStatus_POLICY_CHECK_STATUS_CANCELED), "  ○ OPA policies finished"},
		{"a policy that wasn't evaluated", check(pb.PolicyCheckStatus_POLICY_CHECK_STATUS_ERRORED,
			policy("p1", hard, pb.PolicyCheckPolicyStatus_POLICY_CHECK_POLICY_STATUS_UNSPECIFIED)),
			"  ✗ OPA policies failed\n    ○ p1 (hard-mandatory) not evaluated"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Not the first check of its evaluation, so a blank line sets it off from the one before.
			assert.Equal(t, "\n"+tt.want, render(t, runpkg.PolicyCheckFinished{Results: tt.results}))
		})
	}
}

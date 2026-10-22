package command

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/run"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	runpkg "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/pkg/run"
)

// TestInterruptContexts verifies the first signal cancels only followCtx and the second cleanupCtx.
func TestInterruptContexts(t *testing.T) {
	followCtx, cleanupCtx, release := interruptContexts(context.Background())
	defer release()

	// The handler is registered, so signaling ourselves is caught rather than killing the test.
	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGINT))
	waitDone(followCtx, t)
	assert.NoError(t, cleanupCtx.Err(), "the first interrupt leaves cleanupCtx alone")

	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGINT))
	waitDone(cleanupCtx, t)
	assert.Equal(t, os.Interrupt, run.InterruptSignal(followCtx), "the signal is recorded as the cause")
	assert.Equal(t, interruptedExitCode, signalExitCode(followCtx))
}

// TestInterruptContexts_SIGTERM verifies SIGTERM is recorded as the cause and exits 143.
func TestInterruptContexts_SIGTERM(t *testing.T) {
	followCtx, _, release := interruptContexts(context.Background())
	defer release()

	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGTERM))
	waitDone(followCtx, t)

	assert.Equal(t, syscall.SIGTERM, run.InterruptSignal(followCtx))
	assert.Equal(t, terminatedExitCode, signalExitCode(followCtx))
}

// TestWaitTimeoutFlag verifies -wait-timeout defaults to DefaultWaitTimeout and rejects non-positive values.
func TestWaitTimeoutFlag(t *testing.T) {
	parse := func(args ...string) (*time.Duration, error) {
		var input *bool
		var waitTimeout *time.Duration
		f := flag.NewSet("test")
		addFollowFlags(f, &input, &waitTimeout)
		err := f.Parse(args)
		return waitTimeout, err
	}

	d, err := parse()
	require.NoError(t, err)
	require.NotNil(t, d, "there is always a timeout")
	assert.Equal(t, runpkg.DefaultWaitTimeout, *d)

	d, err = parse("-wait-timeout", "2h")
	require.NoError(t, err)
	assert.Equal(t, 2*time.Hour, *d)

	for _, bad := range []string{"0s", "-5m", "forever"} {
		_, err = parse("-wait-timeout", bad)
		assert.Error(t, err, bad)
	}
}

// TestInterruptContexts_Release verifies release cancels both contexts.
func TestInterruptContexts_Release(t *testing.T) {
	followCtx, cleanupCtx, release := interruptContexts(context.Background())
	release()

	assert.Error(t, followCtx.Err())
	assert.Error(t, cleanupCtx.Err())
}

func waitDone(ctx context.Context, t *testing.T) {
	t.Helper()
	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("context was not canceled by the signal")
	}
}

// recordingUI captures Output, Warnf and Errorf output.
type recordingUI struct {
	terminal.UI
	outputs  []string
	warnings []string
	errors   []string
}

func (u *recordingUI) Output(format string, a ...any) {
	u.outputs = append(u.outputs, fmt.Sprintf(format, a...))
}

func (u *recordingUI) Warnf(format string, a ...any) {
	u.warnings = append(u.warnings, fmt.Sprintf(format, a...))
}

func (u *recordingUI) Errorf(format string, a ...any) {
	u.errors = append(u.errors, fmt.Sprintf(format, a...))
}

func TestReportOutcome(t *testing.T) {
	runWith := func(status pb.RunStatus) *pb.Run {
		return &pb.Run{Metadata: &pb.ResourceMetadata{Id: "run-1"}, Status: status}
	}
	planApproval := runpkg.StageRef{Name: "Plan approval", Kind: runpkg.StagePlanApproval}
	preApply := runpkg.StageRef{Name: "Pre-apply stage", Kind: runpkg.StageTask, TaskStage: pb.RunTaskStageName_RUN_TASK_STAGE_NAME_PRE_APPLY}

	tests := []struct {
		name        string
		outcome     runpkg.Outcome
		waitTimeout time.Duration
		wantCode    int
		wantOutput  string
		wantWarning string
		wantError   string
	}{
		{
			name:     "completed",
			outcome:  runpkg.Outcome{Kind: runpkg.OutcomeCompleted, Run: runWith(pb.RunStatus_APPLIED)},
			wantCode: 0,
		},
		{
			name:        "left waiting at plan approval",
			outcome:     runpkg.Outcome{Kind: runpkg.OutcomeDetached, Run: runWith(pb.RunStatus_PLANNED), Stage: planApproval},
			wantCode:    0,
			wantWarning: "\nRun run-1 is waiting for its plan to be approved. Apply or discard it in the Tharsis UI.",
		},
		{
			name:        "left waiting at a gate",
			outcome:     runpkg.Outcome{Kind: runpkg.OutcomeDetached, Run: runWith(pb.RunStatus_PRE_APPLY_AWAITING_DECISION), Stage: preApply},
			wantCode:    0,
			wantWarning: "\nRun run-1 is waiting for a decision on its pre-apply stage. Approve, override or discard it in the Tharsis UI.",
		},
		{
			name:        "detached after the wait timed out",
			outcome:     runpkg.Outcome{Kind: runpkg.OutcomeDetached, Run: runWith(pb.RunStatus_PLANNED), Stage: planApproval, TimedOut: true},
			waitTimeout: 5 * time.Minute,
			wantCode:    1,
			wantWarning: "\nStopped waiting after 5m0s. Run run-1 is waiting for its plan to be approved. Apply or discard it in the Tharsis UI.",
		},
		{
			name:        "a timeout without a wait timeout reports the default",
			outcome:     runpkg.Outcome{Kind: runpkg.OutcomeDetached, Run: runWith(pb.RunStatus_PLANNED), Stage: planApproval, TimedOut: true},
			wantCode:    1,
			wantWarning: "\nStopped waiting after " + runpkg.DefaultWaitTimeout.String() + ".",
		},
		{
			name:       "discarded on the decider's decision",
			outcome:    runpkg.Outcome{Kind: runpkg.OutcomeDiscarded, Run: runWith(pb.RunStatus_DISCARDED)},
			wantCode:   0,
			wantOutput: "Run discarded.",
		},
		{
			name:       "interrupted and canceled",
			outcome:    runpkg.Outcome{Kind: runpkg.OutcomeInterrupted, Run: runWith(pb.RunStatus_CANCELED), RunCanceled: true},
			wantCode:   terminatedExitCode,
			wantOutput: "\nRun run-1 was canceled.",
		},
		{
			name:       "interrupted and discarded",
			outcome:    runpkg.Outcome{Kind: runpkg.OutcomeInterrupted, Run: runWith(pb.RunStatus_DISCARDED), RunDiscarded: true},
			wantCode:   terminatedExitCode,
			wantOutput: "\nRun run-1 was discarded.",
		},
		{
			name:       "interrupted after the run ended",
			outcome:    runpkg.Outcome{Kind: runpkg.OutcomeInterrupted, Run: runWith(pb.RunStatus_APPLIED)},
			wantCode:   terminatedExitCode,
			wantOutput: "\nStopped following run run-1. The run already ended (applied).",
		},
		{
			name:       "interrupted while the run continues",
			outcome:    runpkg.Outcome{Kind: runpkg.OutcomeInterrupted, Run: runWith(pb.RunStatus_APPLYING)},
			wantCode:   terminatedExitCode,
			wantOutput: "\nStopped following run run-1 (applying). The run continues in Tharsis.",
		},
		{
			name:      "failed because it was discarded elsewhere",
			outcome:   runpkg.Outcome{Kind: runpkg.OutcomeFailed, Run: runWith(pb.RunStatus_DISCARDED), Stage: preApply},
			wantCode:  1,
			wantError: "Run run-1 was discarded",
		},
		{
			name:      "failed because it was canceled",
			outcome:   runpkg.Outcome{Kind: runpkg.OutcomeFailed, Run: runWith(pb.RunStatus_CANCELED)},
			wantCode:  1,
			wantError: "Run run-1 was canceled",
		},
		{
			name:      "failed during a stage",
			outcome:   runpkg.Outcome{Kind: runpkg.OutcomeFailed, Run: runWith(pb.RunStatus_ERRORED), Stage: preApply},
			wantCode:  1,
			wantError: "Run run-1 failed during pre-apply stage",
		},
		{
			name:      "failed with no stage",
			outcome:   runpkg.Outcome{Kind: runpkg.OutcomeFailed, Run: runWith(pb.RunStatus_ERRORED)},
			wantCode:  1,
			wantError: "Run run-1 ended with status errored",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ui := &recordingUI{UI: terminal.NewNoopUI()}

			code := reportOutcome(ui, &test.outcome, &run.FollowOptions{WaitTimeout: test.waitTimeout}, terminatedExitCode)

			assert.Equal(t, test.wantCode, code)
			assertOnlyLine(t, test.wantOutput, ui.outputs, "output")
			assertOnlyLine(t, test.wantWarning, ui.warnings, "warning")
			assertOnlyLine(t, test.wantError, ui.errors, "error")
		})
	}
}

// assertOnlyLine asserts lines is empty when want is, and otherwise holds one line starting with want.
func assertOnlyLine(t *testing.T, want string, lines []string, kind string) {
	t.Helper()
	if want == "" {
		assert.Empty(t, lines, "no %s is expected", kind)
		return
	}
	if assert.Len(t, lines, 1, "one %s is expected", kind) {
		assert.True(t, strings.HasPrefix(lines[0], want), "%s %q doesn't start with %q", kind, lines[0], want)
	}
}

func TestWarnAdvisories(t *testing.T) {
	run1 := &pb.Run{Metadata: &pb.ResourceMetadata{Id: "run-1"}}

	t.Run("lists each failed advisory policy", func(t *testing.T) {
		ui := &recordingUI{UI: terminal.NewNoopUI()}
		warnAdvisories(ui, &runpkg.Outcome{Run: run1, Advisories: []string{"group/a", "group/b"}})
		assert.Equal(t, []string{"2 advisory policies failed on run run-1:", "  group/a", "  group/b"}, ui.warnings)
	})

	t.Run("says nothing without advisories", func(t *testing.T) {
		ui := &recordingUI{UI: terminal.NewNoopUI()}
		warnAdvisories(ui, &runpkg.Outcome{Run: run1})
		assert.Empty(t, ui.warnings)
	})
}

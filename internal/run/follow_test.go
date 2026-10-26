package run

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/runtest"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	runpkg "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/pkg/run"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// These tests follow runs end to end through the CLI's hooks (rendering, prompting and Ctrl-C
// handling) and the generic follower in pkg/run, against fake clients.

// Shorthands for the shared fakes and builders.
type fakeRuns = runtest.Runs

var (
	newTestRun        = runtest.NewRun
	withAutoApply     = runtest.WithAutoApply
	withDestroy       = runtest.WithDestroy
	withPlan          = runtest.WithPlan
	withApply         = runtest.WithApply
	withStage         = runtest.WithStage
	newCheck          = runtest.NewCheck
	softFailedResults = runtest.SoftFailedResults
	appliedRun        = runtest.AppliedRun
)

const (
	prePlan  = runtest.PrePlan
	postPlan = runtest.PostPlan
	preApply = runtest.PreApply
)

const (
	OutcomeCompleted   = runpkg.OutcomeCompleted
	OutcomeDetached    = runpkg.OutcomeDetached
	OutcomeDiscarded   = runpkg.OutcomeDiscarded
	OutcomeFailed      = runpkg.OutcomeFailed
	OutcomeInterrupted = runpkg.OutcomeInterrupted
)

// fakeUI records output and answers prompts from answers, blocking until one is available.
type fakeUI struct {
	terminal.UI
	interactive bool
	answers     chan string
	// inputErr, when set, is returned by Input, as a UI that can't read at all does.
	inputErr error

	mu     sync.Mutex
	lines  []string
	inputs int
}

func newFakeUI(interactive bool, answers ...string) *fakeUI {
	ch := make(chan string, len(answers))
	for _, a := range answers {
		ch <- a
	}
	return &fakeUI{UI: terminal.NewNoopUI(), interactive: interactive, answers: ch}
}

func (u *fakeUI) Output(msg string, raw ...interface{}) {
	msg, _, _ = terminal.Interpret(msg, raw...)
	u.mu.Lock()
	defer u.mu.Unlock()
	u.lines = append(u.lines, msg)
}

func (u *fakeUI) Interactive() bool { return u.interactive }

func (u *fakeUI) Input(input *terminal.Input) (string, error) {
	if u.inputErr != nil {
		return "", u.inputErr
	}
	u.Output(input.Prompt)
	u.mu.Lock()
	u.inputs++
	u.mu.Unlock()
	answer, ok := <-u.answers
	if !ok {
		return "", io.EOF
	}
	return answer, nil
}

func (u *fakeUI) output() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	return strings.Join(u.lines, "\n")
}

func (u *fakeUI) inputCount() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.inputs
}

// fakeRuns serves run snapshots from a queue: each GetRunByID returns the head, popping it while
// more remain, so a test scripts the run's progress. Actions can replace the queue. It also serves

type harness struct {
	runs *runtest.Runs
	jobs *runtest.Jobs
	ui   *fakeUI
	mgr  *Manager
}

func newHarness(t *testing.T, ui *fakeUI, runs ...*pb.Run) *harness {
	prevPoll, prevRetry := pollInterval, readRetryDelay
	pollInterval = 5 * time.Millisecond
	readRetryDelay = time.Millisecond
	t.Cleanup(func() { pollInterval, readRetryDelay = prevPoll, prevRetry })

	h := &harness{runs: runtest.NewRuns(runs...), jobs: &runtest.Jobs{}, ui: ui}
	h.mgr = &Manager{grpcClient: runtest.Client(h.runs, h.jobs), logger: hclog.NewNullLogger(), ui: ui}
	return h
}

func (h *harness) follow(t *testing.T, created *pb.Run, opts *FollowOptions) *runpkg.Outcome {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	outcome, err := h.mgr.FollowRun(ctx, ctx, created, opts)
	require.NoError(t, err)
	require.NotEqual(t, OutcomeInterrupted, outcome.Kind, "timed out following the run; output:\n%s", h.ui.output())
	return outcome
}

func TestFollowRun_TerraformStages(t *testing.T) {
	created := newTestRun(pb.RunStatus_PENDING)

	t.Run("a run that jumps straight to applied still shows each stage in order", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true), appliedRun())

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, []string{"plan-job", "apply-job"}, h.jobs.Streamed())
		assert.Equal(t, 0, h.ui.inputCount(), "the plan was already approved, so there is nothing to ask")
	})

	t.Run("approving the plan applies the run", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "yes"),
			newTestRun(pb.RunStatus_PLANNED, withPlan(pb.PlanStatus_FINISHED, "plan-job")))
		h.runs.OnApply = func(f *runtest.Runs) { f.Set(appliedRun()) }

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, 1, h.runs.Applied)
		assert.Equal(t, []string{"plan-job", "apply-job"}, h.jobs.Streamed())
		assert.Contains(t, h.ui.output(), "Do you approve to apply the above plan?")
	})

	t.Run("declining the plan discards the run", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "no"),
			newTestRun(pb.RunStatus_PLANNED, withDestroy(), withPlan(pb.PlanStatus_FINISHED, "plan-job")))
		h.runs.OnDiscard = func(f *runtest.Runs) {
			f.Set(newTestRun(pb.RunStatus_DISCARDED, withPlan(pb.PlanStatus_FINISHED, "plan-job"), withApply(pb.ApplyStatus_SKIPPED, "")))
		}

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeDiscarded, outcome.Kind)
		assert.Equal(t, 1, h.runs.Discarded)
		assert.Equal(t, 0, h.runs.Applied)
		assert.Contains(t, h.ui.output(), "Do you approve to destroy the above resources?")
	})

	t.Run("with -wait and without input the plan approval is waited on from the UI", func(t *testing.T) {
		planned := newTestRun(pb.RunStatus_PLANNED, withPlan(pb.PlanStatus_FINISHED, "plan-job"))
		// The plan is approved in the UI after a while.
		h := newHarness(t, newFakeUI(true), planned, planned, planned, appliedRun())
		h.runs.HoldUntil = func() bool { return strings.Contains(h.ui.output(), "Not asking for approval") }

		outcome := h.follow(t, created, &FollowOptions{Input: false, Wait: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, 0, h.ui.inputCount())
		assert.Equal(t, 0, h.runs.Applied, "the CLI didn't apply it")
		assert.Equal(t, []string{"plan-job", "apply-job"}, h.jobs.Streamed(), "the apply is followed once approved elsewhere")
		out := h.ui.output()
		assert.Equal(t, 1, strings.Count(out, "Not asking for approval since -input was false."))
		assert.Contains(t, out, "Waiting for the plan to be applied or the run discarded in the Tharsis UI")
	})

	t.Run("without input or -wait the run is left waiting at plan approval", func(t *testing.T) {
		planned := newTestRun(pb.RunStatus_PLANNED, withPlan(pb.PlanStatus_FINISHED, "plan-job"))
		h := newHarness(t, newFakeUI(true), planned)

		outcome := h.follow(t, created, &FollowOptions{Input: false})

		assert.Equal(t, OutcomeDetached, outcome.Kind)
		assert.False(t, outcome.TimedOut, "it stops without waiting")
		assert.Equal(t, "Plan approval", outcome.Stage.Name)
		assert.Equal(t, "-input was false", outcome.LeaveReason)
		assert.Equal(t, 0, h.ui.inputCount())
		assert.Equal(t, 0, h.runs.Applied)
		assert.Equal(t, 0, h.runs.Discarded)
		assert.NotContains(t, h.ui.output(), "Waiting for the plan", "it doesn't wait in the UI")
	})

	t.Run("without -wait a UI that can't read leaves the run waiting", func(t *testing.T) {
		for name, inputErr := range map[string]error{
			"the terminal is not interactive":    terminal.ErrNonInteractive,
			"no answer could be read from stdin": io.EOF,
		} {
			t.Run(name, func(t *testing.T) {
				ui := newFakeUI(false)
				ui.inputErr = inputErr
				h := newHarness(t, ui, newTestRun(pb.RunStatus_PLANNED, withPlan(pb.PlanStatus_FINISHED, "plan-job")))

				outcome := h.follow(t, created, &FollowOptions{Input: true})

				assert.Equal(t, OutcomeDetached, outcome.Kind)
				assert.Equal(t, name, outcome.LeaveReason)
			})
		}
	})

	t.Run("a piped approval applies the run", func(t *testing.T) {
		h := newHarness(t, newFakeUI(false, "yes"),
			newTestRun(pb.RunStatus_PLANNED, withPlan(pb.PlanStatus_FINISHED, "plan-job")))
		h.runs.OnApply = func(f *runtest.Runs) { f.Set(appliedRun()) }

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, 1, h.runs.Applied)
	})

	t.Run("with -wait a UI that can't read waits for the plan to be decided elsewhere", func(t *testing.T) {
		ui := newFakeUI(false)
		ui.inputErr = terminal.ErrNonInteractive
		planned := newTestRun(pb.RunStatus_PLANNED, withPlan(pb.PlanStatus_FINISHED, "plan-job"))
		// The run is discarded in the UI after a while.
		h := newHarness(t, ui, planned, planned, planned,
			newTestRun(pb.RunStatus_DISCARDED, withPlan(pb.PlanStatus_FINISHED, "plan-job")))
		h.runs.HoldUntil = func() bool { return strings.Contains(ui.output(), "Not asking for approval") }

		outcome := h.follow(t, created, &FollowOptions{Input: true, Wait: true})

		assert.Equal(t, OutcomeFailed, outcome.Kind)
		assert.Equal(t, pb.RunStatus_DISCARDED, outcome.Run.Status)
		assert.Equal(t, 0, h.runs.Applied)
		assert.Equal(t, 0, h.runs.Discarded, "the CLI didn't discard it")
		assert.Equal(t, 1, strings.Count(ui.output(), "Not asking for approval since the terminal is not interactive."))
	})

	t.Run("with -wait empty stdin waits for the plan to be decided elsewhere", func(t *testing.T) {
		ui := newFakeUI(false)
		ui.inputErr = io.EOF
		planned := newTestRun(pb.RunStatus_PLANNED, withPlan(pb.PlanStatus_FINISHED, "plan-job"))
		h := newHarness(t, ui, planned, planned, planned, appliedRun())
		h.runs.HoldUntil = func() bool { return strings.Contains(ui.output(), "Not asking for approval") }

		outcome := h.follow(t, created, &FollowOptions{Input: true, Wait: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, 0, h.runs.Applied)
		assert.Equal(t, 1, strings.Count(ui.output(), "Not asking for approval since no answer could be read from stdin."))
	})

	t.Run("pressing Enter at the plan leaves the run waiting", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, ""),
			newTestRun(pb.RunStatus_PLANNED, withPlan(pb.PlanStatus_FINISHED, "plan-job")))

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeDetached, outcome.Kind)
		assert.Equal(t, "Plan approval", outcome.Stage.Name)
		assert.Empty(t, outcome.LeaveReason, "the user chose to leave it")
		assert.Equal(t, 0, h.runs.Applied)
		assert.Equal(t, 0, h.runs.Discarded)
	})

	t.Run("an unrecognized answer at the plan asks again", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "maybe", "yes"),
			newTestRun(pb.RunStatus_PLANNED, withPlan(pb.PlanStatus_FINISHED, "plan-job")))
		h.runs.OnApply = func(f *runtest.Runs) { f.Set(appliedRun()) }

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, 1, h.runs.Applied)
		assert.Equal(t, 2, h.ui.inputCount())
		assert.Contains(t, h.ui.output(), `"maybe" isn't one of the choices.`)
	})

	t.Run("an apply whose snapshot still shows planned isn't asked about again", func(t *testing.T) {
		planned := newTestRun(pb.RunStatus_PLANNED, withPlan(pb.PlanStatus_FINISHED, "plan-job"))
		h := newHarness(t, newFakeUI(true, "yes", "yes"), planned)
		// The apply conflicts (approved elsewhere at the same moment), and the next reads lag behind.
		h.runs.OnApply = func(f *runtest.Runs) { f.Set(planned, planned, planned, appliedRun()) }

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, 1, h.runs.Applied)
		assert.Equal(t, 1, h.ui.inputCount())
	})

	t.Run("a run that is over isn't waited on, even if a stage looks unstarted", func(t *testing.T) {
		// Applied, but without the apply in the snapshot.
		r := newTestRun(pb.RunStatus_APPLIED, withPlan(pb.PlanStatus_FINISHED, "plan-job"))
		r.Apply = nil
		h := newHarness(t, newFakeUI(true), r)

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
	})

	t.Run("a finished run with a stage still created completes", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true),
			newTestRun(pb.RunStatus_PLANNED_AND_FINISHED, withPlan(pb.PlanStatus_FINISHED, "plan-job"),
				withStage(postPlan, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_CREATED)))

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
	})

	t.Run("the plan approval wait times out", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true), newTestRun(pb.RunStatus_PLANNED, withPlan(pb.PlanStatus_FINISHED, "plan-job")))

		outcome := h.follow(t, created, &FollowOptions{Input: false, Wait: true, WaitTimeout: 30 * time.Millisecond})

		assert.Equal(t, OutcomeDetached, outcome.Kind)
		assert.True(t, outcome.TimedOut)
		assert.Equal(t, "Plan approval", outcome.Stage.Name)
	})

	t.Run("an unanswered plan prompt times out and leaves the run waiting", func(t *testing.T) {
		ui := newFakeUI(true) // never answers
		t.Cleanup(func() { close(ui.answers) })
		h := newHarness(t, ui, newTestRun(pb.RunStatus_PLANNED, withPlan(pb.PlanStatus_FINISHED, "plan-job")))

		outcome := h.follow(t, created, &FollowOptions{Input: true, WaitTimeout: 50 * time.Millisecond})

		assert.Equal(t, OutcomeDetached, outcome.Kind)
		assert.True(t, outcome.TimedOut)
		assert.Equal(t, 1, ui.inputCount())
		assert.Equal(t, 0, h.runs.Applied)
		assert.Equal(t, 0, h.runs.Discarded)
	})

	t.Run("transient read errors are retried", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true), appliedRun())
		h.runs.GetErrs = []error{
			status.Error(codes.Unavailable, "draining"),
			status.Error(codes.Aborted, "optimistic lock"),
		}

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
	})

	t.Run("a read error that isn't transient ends following", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true), appliedRun())
		h.runs.GetErrs = []error{status.Error(codes.PermissionDenied, "no")}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := h.mgr.FollowRun(ctx, ctx, created, &FollowOptions{Input: true})

		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("a plan with no changes stops after the plan", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true),
			newTestRun(pb.RunStatus_PLANNED_AND_FINISHED, withPlan(pb.PlanStatus_FINISHED, "plan-job"), withApply(pb.ApplyStatus_SKIPPED, "")))

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, []string{"plan-job"}, h.jobs.Streamed())
		assert.Contains(t, h.ui.output(), "Stopping since plan had no changes.")
	})

	t.Run("auto-approve goes straight into the apply", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true),
			newTestRun(pb.RunStatus_APPLYING, withAutoApply(), withPlan(pb.PlanStatus_FINISHED, "plan-job"), withApply(pb.ApplyStatus_RUNNING, "apply-job")),
			appliedRun(withAutoApply()))

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, 1, strings.Count(h.ui.output(), "Auto-approving."))
		assert.Equal(t, 0, h.ui.inputCount())
		assert.Equal(t, 0, h.runs.Applied)
	})

	t.Run("a failed plan fails the run", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true),
			newTestRun(pb.RunStatus_ERRORED, withPlan(pb.PlanStatus_ERRORED, "plan-job"), withApply(pb.ApplyStatus_SKIPPED, "")))

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeFailed, outcome.Kind)
		assert.Equal(t, []string{"plan-job"}, h.jobs.Streamed(), "the failed plan's logs are still shown")
	})

	t.Run("a retried plan streams its new job", func(t *testing.T) {
		// The plan errored while its logs streamed and was retried by the time the run was re-read.
		h := newHarness(t, newFakeUI(true, "yes"),
			newTestRun(pb.RunStatus_PLANNING, withPlan(pb.PlanStatus_RUNNING, "plan-job")),
			newTestRun(pb.RunStatus_PLANNED, withPlan(pb.PlanStatus_FINISHED, "plan-job-2")))
		h.runs.OnApply = func(f *runtest.Runs) { f.Set(appliedRun(withPlan(pb.PlanStatus_FINISHED, "plan-job-2"))) }

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, []string{"plan-job", "plan-job-2", "apply-job"}, h.jobs.Streamed())
		assert.Contains(t, h.ui.output(), "Plan (retry)")
	})

	t.Run("an auto-apply run parked at planned waits for the server to start the apply", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true),
			newTestRun(pb.RunStatus_PLANNED, withAutoApply(), withPlan(pb.PlanStatus_FINISHED, "plan-job")),
			appliedRun(withAutoApply()))

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, 0, h.runs.Applied, "the server applies an auto-apply run, not the CLI")
		assert.Equal(t, 0, h.ui.inputCount())
	})
}

func TestFollowRun_Interrupt(t *testing.T) {
	created := newTestRun(pb.RunStatus_PENDING)
	planning := func() *pb.Run {
		return newTestRun(pb.RunStatus_PLAN_QUEUED, withPlan(pb.PlanStatus_QUEUED, ""))
	}

	// interruptWith follows the run with followCtx canceled by sig after a moment, as interruptContexts
	// does on the first signal.
	interruptWith := func(h *harness, sig os.Signal, cleanupCtx context.Context, opts *FollowOptions) *runpkg.Outcome {
		followCtx, stop := context.WithCancelCause(context.Background())
		timer := time.AfterFunc(50*time.Millisecond, func() { stop(&SignalError{Signal: sig}) })
		defer timer.Stop()
		defer stop(nil)
		outcome, err := h.mgr.FollowRun(followCtx, cleanupCtx, created, opts)
		require.NoError(t, err)
		require.Equal(t, OutcomeInterrupted, outcome.Kind)
		return outcome
	}
	interrupt := func(h *harness, cleanupCtx context.Context, opts *FollowOptions) *runpkg.Outcome {
		return interruptWith(h, os.Interrupt, cleanupCtx, opts)
	}
	// assertLeftGoing checks the run was neither canceled nor discarded, and nothing was asked.
	assertLeftGoing := func(t *testing.T, h *harness, outcome *runpkg.Outcome) {
		t.Helper()
		assert.False(t, outcome.RunCanceled)
		assert.Equal(t, 0, h.runs.Count(&h.runs.Canceled))
		assert.Equal(t, 0, h.runs.Count(&h.runs.Discarded))
		assert.NotContains(t, h.ui.output(), "Do you want to cancel")
	}

	t.Run("Ctrl-C at a terminal cancels the run when the user answers yes", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "yes"), planning())

		outcome := interrupt(h, context.Background(), &FollowOptions{Input: true})

		assert.True(t, outcome.RunCanceled)
		assert.Equal(t, 1, h.runs.Count(&h.runs.Canceled))
		assert.Contains(t, h.ui.output(), "Do you want to cancel run run-1?")
	})

	t.Run("Ctrl-C at a terminal leaves the run going for any other answer", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "no"), planning())

		outcome := interrupt(h, context.Background(), &FollowOptions{Input: true})

		assert.False(t, outcome.RunCanceled)
		assert.Equal(t, 0, h.runs.Count(&h.runs.Canceled))
		assert.Contains(t, h.ui.output(), "The run was not canceled.")
	})

	t.Run("a second interrupt exits without canceling", func(t *testing.T) {
		// No answer is ever typed; the second interrupt arrives while the question is showing.
		h := newHarness(t, newFakeUI(true), planning())
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		outcome := interrupt(h, cleanupCtx, &FollowOptions{Input: true})

		assert.False(t, outcome.RunCanceled)
		assert.Equal(t, 0, h.runs.Count(&h.runs.Canceled))
		assert.Contains(t, h.ui.output(), "Two interrupts received. Exiting immediately.")
	})

	t.Run("SIGTERM leaves the run going without asking, even at an interactive terminal", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "yes"), planning())

		outcome := interruptWith(h, syscall.SIGTERM, context.Background(), &FollowOptions{Input: true})

		assertLeftGoing(t, h, outcome)
		assert.Equal(t, 0, h.ui.inputCount())
		assert.Contains(t, h.ui.output(), "Not canceling the run since the CLI was terminated (SIGTERM).")
	})

	t.Run("Ctrl-C without a terminal leaves the run going", func(t *testing.T) {
		for _, input := range []bool{true, false} {
			h := newHarness(t, newFakeUI(false), planning())

			outcome := interrupt(h, context.Background(), &FollowOptions{Input: input})

			assertLeftGoing(t, h, outcome)
			assert.Contains(t, h.ui.output(), "Not canceling the run since the terminal is not interactive.")
		}
	})

	t.Run("Ctrl-C with -input=false at a terminal leaves the run going", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true), planning())

		outcome := interrupt(h, context.Background(), &FollowOptions{Input: false})

		assertLeftGoing(t, h, outcome)
		assert.Contains(t, h.ui.output(), "Not asking whether to cancel the run since -input was false.")
	})

	t.Run("a run waiting on a decision is never canceled or discarded", func(t *testing.T) {
		waiting := map[string]*pb.Run{
			"plan approval": newTestRun(pb.RunStatus_PLANNED, withPlan(pb.PlanStatus_FINISHED, "plan-job")),
			"a policy gate": newTestRun(pb.RunStatus_PRE_APPLY_AWAITING_DECISION, withPlan(pb.PlanStatus_FINISHED, "plan-job"),
				withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_AWAITING_OVERRIDE)),
		}
		cases := []struct {
			desc        string
			interactive bool
			input       bool
			sig         os.Signal
		}{
			{"Ctrl-C at a terminal", true, true, os.Interrupt},
			{"Ctrl-C with -input=false", true, false, os.Interrupt},
			{"Ctrl-C without a terminal", false, true, os.Interrupt},
			{"SIGTERM without a terminal", false, false, syscall.SIGTERM},
		}
		for name, r := range waiting {
			for _, c := range cases {
				t.Run(name+", "+c.desc, func(t *testing.T) {
					ui := newFakeUI(c.interactive)
					if !c.interactive {
						// Plan approval still reads stdin without a terminal; in CI it is often empty.
						ui.inputErr = io.EOF
					}
					h := newHarness(t, ui, r)

					outcome := interruptWith(h, c.sig, context.Background(), &FollowOptions{Input: c.input, Wait: true})

					assertLeftGoing(t, h, outcome)
				})
			}
		}
	})
}

func TestFollowRun_PolicyStages(t *testing.T) {
	created := newTestRun(pb.RunStatus_PENDING, withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_CREATED,
		newCheck("c1", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_CREATED, "")))

	awaiting := func(checkJob string) *pb.Run {
		return newTestRun(pb.RunStatus_PRE_APPLY_AWAITING_DECISION,
			withPlan(pb.PlanStatus_FINISHED, "plan-job"),
			withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_AWAITING_OVERRIDE,
				newCheck("c1", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_SOFT_FAILED, checkJob)))
	}
	cleared := func(checkJob string, checkStatus pb.PolicyCheckStatus) *pb.Run {
		return appliedRun(withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_COMPLETED,
			newCheck("c1", checkStatus, checkJob)))
	}

	t.Run("approving a soft-failed gate lets the run apply", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "approve"), awaiting("check-job"))
		h.runs.Results["c1"] = softFailedResults("c1", 1, 0)
		h.runs.OnApprove = func(f *runtest.Runs) *pb.RunGate {
			f.Set(cleared("check-job", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_OVERRIDDEN))
			g := softFailedResults("c1", 1, 1).RunGate
			g.Status = pb.RunGateStatus_RUN_GATE_STATUS_APPROVED
			return g
		}

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, 1, h.runs.Approved)
		assert.Equal(t, []string{"plan-job", "apply-job"}, h.jobs.Streamed(), "policy check logs aren't streamed")
		out := h.ui.output()
		assert.Contains(t, out, strings.Join([]string{
			"Pre-apply stage",
			"\nPolicy checks",
			"  ! OPA policies failed; approval required",
			"    ✗ group/encrypted-buckets (soft-mandatory)",
			"        - bucket must be encrypted",
		}, "\n"), "the stage's policy checks are a section under its header")
		assert.Contains(t, out, "Approval recorded; all required approvals are in.")
	})

	t.Run("a gate needing more approvals waits for other approvers", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "approve"), awaiting("check-job"))
		h.runs.Results["c1"] = softFailedResults("c1", 2, 0)
		h.runs.OnApprove = func(f *runtest.Runs) *pb.RunGate {
			f.Set(awaiting("check-job"), awaiting("check-job"), cleared("check-job", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_OVERRIDDEN))
			f.Results["c1"] = softFailedResults("c1", 2, 1) // the approval shows in the results
			return f.Results["c1"].RunGate
		}

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, 1, h.ui.inputCount(), "the user isn't asked again after approving")
		out := h.ui.output()
		assert.Contains(t, out, "Approval recorded.")
		assert.Contains(t, out, "Waiting for other approvers (1 of 2 approvals)")
	})

	t.Run("an ineligible approver can leave the run waiting", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "approve", ""), awaiting("check-job"))
		h.runs.Results["c1"] = softFailedResults("c1", 1, 0)
		h.runs.ApproveErr = status.Error(codes.PermissionDenied, "caller is not an eligible approver")

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeDetached, outcome.Kind)
		assert.Equal(t, 2, h.ui.inputCount())
		assert.Contains(t, h.ui.output(), "You are not an eligible approver for this gate.")
	})

	// A gate whose failed policies declared no approvers has no approval rules; only an override clears it.
	noApprovers := func() *pb.PolicyCheckResults {
		r := softFailedResults("c1", 1, 0)
		r.RunGate.ApprovalRules = nil
		return r
	}

	t.Run("a gate with no approvers is overridden", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "override"), awaiting("check-job"))
		h.runs.Results["c1"] = noApprovers()
		h.runs.OnOverride = func(f *runtest.Runs) *pb.RunGate {
			f.Set(cleared("check-job", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_OVERRIDDEN))
			g := noApprovers().RunGate
			g.Status = pb.RunGateStatus_RUN_GATE_STATUS_OVERRIDDEN
			return g
		}

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, 1, h.runs.Overridden)
		assert.Equal(t, 0, h.runs.Approved)
		out := h.ui.output()
		assert.Contains(t, out, "Override required to continue.")
		assert.Contains(t, out, "Enter 'override' to override, 'discard' to discard the run")
		assert.NotContains(t, out, "'approve'")
		assert.Contains(t, out, "Gate overridden.")
	})

	t.Run("a gate with no approvers can't be approved", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "approve", "discard"), awaiting("check-job"))
		h.runs.Results["c1"] = noApprovers()

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeDiscarded, outcome.Kind)
		assert.Equal(t, 0, h.runs.Approved)
		assert.Contains(t, h.ui.output(), "You can't approve this gate.")
	})

	t.Run("a gate with approvers isn't offered an override", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "override", "discard"), awaiting("check-job"))
		h.runs.Results["c1"] = softFailedResults("c1", 1, 0)

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeDiscarded, outcome.Kind)
		assert.Equal(t, 0, h.runs.Overridden)
		assert.Contains(t, h.ui.output(), "You can't override this gate.")
	})

	t.Run("a caller without override permission can leave the run waiting", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "override", ""), awaiting("check-job"))
		h.runs.Results["c1"] = noApprovers()
		h.runs.OverrideErr = status.Error(codes.PermissionDenied, "caller may not override")

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeDetached, outcome.Kind)
		out := h.ui.output()
		assert.Contains(t, out, "You don't have permission to override this gate.")
		assert.Contains(t, out, "Enter 'discard' to discard the run, or press Enter", "override isn't offered again")
	})

	t.Run("an unrecognized answer asks again", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "aprove", "discard"), awaiting("check-job"))
		h.runs.Results["c1"] = softFailedResults("c1", 1, 0)

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeDiscarded, outcome.Kind)
		assert.Equal(t, 2, h.ui.inputCount())
		assert.Contains(t, h.ui.output(), `"aprove" isn't one of the choices.`)
	})

	t.Run("discarding at a gate discards the run", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "discard"), awaiting("check-job"))
		h.runs.Results["c1"] = softFailedResults("c1", 1, 0)

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeDiscarded, outcome.Kind)
		assert.Equal(t, 1, h.runs.Discarded)
	})

	t.Run("with -wait and without input the gate is waited on from the UI", func(t *testing.T) {
		// The gate is approved in the UI after a while.
		h := newHarness(t, newFakeUI(true), awaiting("check-job"), awaiting("check-job"), awaiting("check-job"),
			cleared("check-job", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_OVERRIDDEN))
		h.runs.Results["c1"] = softFailedResults("c1", 2, 1)
		h.runs.HoldUntil = func() bool { return strings.Contains(h.ui.output(), "Not asking for a decision") }

		outcome := h.follow(t, created, &FollowOptions{Input: false, Wait: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, 0, h.ui.inputCount())
		assert.Equal(t, 0, h.runs.Approved, "the CLI didn't approve it")
		out := h.ui.output()
		assert.Contains(t, out, "bucket must be encrypted", "findings are shown even when not asking")
		assert.Equal(t, 1, strings.Count(out, "Approval required to continue:"), "the gate is shown once")
		assert.Contains(t, out, "Not asking for a decision since -input was false.")
		assert.Contains(t, out, "Waiting for the gate to be approved, overridden or the run discarded in the Tharsis UI (1 of 2 approvals)")
	})

	t.Run("without input or -wait the run is left waiting at a gate", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true), awaiting("check-job"))
		h.runs.Results["c1"] = softFailedResults("c1", 2, 1)

		outcome := h.follow(t, created, &FollowOptions{Input: false})

		assert.Equal(t, OutcomeDetached, outcome.Kind)
		assert.False(t, outcome.TimedOut)
		assert.Equal(t, preApply, outcome.Stage.TaskStage)
		assert.Equal(t, "-input was false", outcome.LeaveReason)
		assert.Equal(t, 0, h.ui.inputCount())
		assert.Equal(t, 0, h.runs.Approved)
		assert.Contains(t, h.ui.output(), "bucket must be encrypted", "findings are still shown")
	})

	t.Run("with -wait and without a terminal a gate discarded in the UI fails the run", func(t *testing.T) {
		discarded := newTestRun(pb.RunStatus_DISCARDED, withPlan(pb.PlanStatus_FINISHED, "plan-job"),
			withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_CANCELED,
				newCheck("c1", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_SOFT_FAILED, "check-job")))
		h := newHarness(t, newFakeUI(false), awaiting("check-job"), awaiting("check-job"), discarded)
		h.runs.Results["c1"] = softFailedResults("c1", 1, 0)

		outcome := h.follow(t, created, &FollowOptions{Input: true, Wait: true})

		assert.Equal(t, OutcomeFailed, outcome.Kind)
		assert.Equal(t, pb.RunStatus_DISCARDED, outcome.Run.Status)
		assert.Contains(t, h.ui.output(), "Not asking for a decision since the terminal is not interactive.")
	})

	t.Run("a check retried while the prompt is open is followed again", func(t *testing.T) {
		ui := newFakeUI(true) // never answers
		t.Cleanup(func() { close(ui.answers) })
		retrying := newTestRun(pb.RunStatus_PRE_APPLY_RUNNING,
			withPlan(pb.PlanStatus_FINISHED, "plan-job"),
			withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_RUNNING,
				newCheck("c1", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_RUNNING, "check-job-2")))
		// One snapshot is used streaming the plan job; the second opens the prompt, and the retry is
		// seen while it is open.
		h := newHarness(t, ui, awaiting("check-job"), awaiting("check-job"), retrying,
			cleared("check-job-2", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_PASSED))
		h.runs.Results["c1"] = softFailedResults("c1", 1, 0)

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, []string{"plan-job", "apply-job"}, h.jobs.Streamed())
		out := h.ui.output()
		assert.Contains(t, out, "The run moved on before an answer was given.")
		assert.Contains(t, out, "Pre-apply stage (retry)")
		assert.Equal(t, 2, strings.Count(out, "Pre-apply stage"), "the stage is shown again when it re-runs")
	})

	t.Run("a check that re-ran and settled again between reads is shown again", func(t *testing.T) {
		// The stage never shows as running again: it is awaiting a decision before and after.
		h := newHarness(t, newFakeUI(true), awaiting("check-job"), awaiting("check-job"), awaiting("check-job-2"),
			cleared("check-job-2", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_OVERRIDDEN))
		h.runs.Results["c1"] = softFailedResults("c1", 1, 0)
		h.runs.HoldUntil = func() bool { return strings.Contains(h.ui.output(), "Not asking for a decision") }

		outcome := h.follow(t, created, &FollowOptions{Input: false, Wait: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		out := h.ui.output()
		assert.Contains(t, out, "Pre-apply stage (retry)")
		assert.Equal(t, 2, strings.Count(out, "OPA policies failed; approval required"), "the new findings are printed")
	})

	t.Run("failed advisory policies are reported in the outcome", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true), cleared("check-job", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_PASSED))
		results := softFailedResults("c1", 1, 0)
		results.PolicyCheck.Status = pb.PolicyCheckStatus_POLICY_CHECK_STATUS_PASSED
		results.PolicyResults[0].Policy.EnforcementLevel = pb.PolicyEnforcementLevel_POLICY_ENFORCEMENT_LEVEL_ADVISORY
		results.RunGate = nil
		h.runs.Results["c1"] = results

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, []string{"group/encrypted-buckets"}, outcome.Advisories)
	})

	t.Run("a run without advisory failures reports none", func(t *testing.T) {
		h := newHarness(t, newFakeUI(true, "approve"), awaiting("check-job"))
		h.runs.Results["c1"] = softFailedResults("c1", 1, 0)
		h.runs.OnApprove = func(f *runtest.Runs) *pb.RunGate {
			f.Set(cleared("check-job", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_OVERRIDDEN))
			g := softFailedResults("c1", 1, 1).RunGate
			g.Status = pb.RunGateStatus_RUN_GATE_STATUS_APPROVED
			return g
		}

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Empty(t, outcome.Advisories, "a soft-mandatory failure isn't an advisory")
	})

	t.Run("a hard failure at pre-plan fails the run with its findings", func(t *testing.T) {
		createdPrePlan := newTestRun(pb.RunStatus_PENDING, withStage(prePlan, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_CREATED,
			newCheck("c1", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_CREATED, "")))
		h := newHarness(t, newFakeUI(true), newTestRun(pb.RunStatus_ERRORED,
			withPlan(pb.PlanStatus_UNSPECIFIED, ""),
			withStage(prePlan, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_ERRORED,
				newCheck("c1", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_ERRORED, ""))))
		results := softFailedResults("c1", 1, 0)
		results.PolicyCheck.Status = pb.PolicyCheckStatus_POLICY_CHECK_STATUS_ERRORED
		results.PolicyResults[0].Policy.EnforcementLevel = pb.PolicyEnforcementLevel_POLICY_ENFORCEMENT_LEVEL_HARD_MANDATORY
		results.RunGate = nil
		h.runs.Results["c1"] = results

		outcome := h.follow(t, createdPrePlan, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeFailed, outcome.Kind)
		assert.Empty(t, h.jobs.Streamed(), "the plan never ran")
		out := h.ui.output()
		assert.Contains(t, out, "(hard-mandatory)")
		assert.Contains(t, out, "bucket must be encrypted")
	})
}

func TestFollowRun_UnsupportedTasks(t *testing.T) {
	// A policy type added to the API after this CLI was built.
	const futureType = pb.PolicyCheckType(99)
	futureCheck := func(s pb.PolicyCheckStatus) *pb.PolicyCheck {
		c := newCheck("c1", s, "check-job")
		c.CheckType = futureType
		return c
	}
	created := newTestRun(pb.RunStatus_PENDING, withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_CREATED,
		futureCheck(pb.PolicyCheckStatus_POLICY_CHECK_STATUS_CREATED)))
	awaiting := newTestRun(pb.RunStatus_PRE_APPLY_AWAITING_DECISION, withAutoApply(),
		withPlan(pb.PlanStatus_FINISHED, "plan-job"),
		withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_AWAITING_OVERRIDE,
			futureCheck(pb.PolicyCheckStatus_POLICY_CHECK_STATUS_SOFT_FAILED)))
	cleared := appliedRun(withAutoApply(), withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_COMPLETED,
		futureCheck(pb.PolicyCheckStatus_POLICY_CHECK_STATUS_OVERRIDDEN)))

	t.Run("a check of an unsupported type is noted and its results aren't fetched", func(t *testing.T) {
		// The fake has no results for c1, so fetching them would fail the follow.
		h := newHarness(t, newFakeUI(true), cleared)

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		out := h.ui.output()
		assert.Contains(t, out, "Pre-apply stage")
		assert.Contains(t, out, "Unsupported policy check type (99); waiting for it to finish.")
	})

	for _, c := range []struct {
		interactive, wait bool
	}{{true, false}, {false, true}} {
		t.Run(fmt.Sprintf("a stage blocked on an unsupported check waits instead of prompting (interactive=%v, wait=%v)", c.interactive, c.wait), func(t *testing.T) {
			h := newHarness(t, newFakeUI(c.interactive), awaiting, awaiting, cleared)
			h.runs.HoldUntil = func() bool { return strings.Contains(h.ui.output(), "Waiting for the run to continue") }

			outcome := h.follow(t, created, &FollowOptions{Input: true, Wait: c.wait})

			assert.Equal(t, OutcomeCompleted, outcome.Kind)
			assert.Equal(t, 0, h.ui.inputCount(), "there is nothing this CLI can ask about")
			out := h.ui.output()
			assert.Equal(t, 1, strings.Count(out, "Unsupported policy check type"), "the check is noted once")
			assert.Contains(t, out, "Waiting for the run to continue; it is blocked on a task this version of the CLI can't show")
		})
	}

	t.Run("without a terminal or -wait a stage blocked on an unsupported check is left waiting", func(t *testing.T) {
		h := newHarness(t, newFakeUI(false), awaiting)

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeDetached, outcome.Kind)
		assert.False(t, outcome.TimedOut, "it stops without waiting")
		assert.Equal(t, preApply, outcome.Stage.TaskStage)
		assert.Equal(t, "the terminal is not interactive", outcome.LeaveReason)
		assert.Equal(t, 0, h.ui.inputCount())
	})

	t.Run("with -input=false a stage with no checks this CLI can see is left waiting", func(t *testing.T) {
		blocked := newTestRun(pb.RunStatus_PRE_APPLY_AWAITING_DECISION, withAutoApply(), withPlan(pb.PlanStatus_FINISHED, "plan-job"),
			withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_AWAITING_OVERRIDE))
		h := newHarness(t, newFakeUI(true), blocked)

		outcome := h.follow(t, newTestRun(pb.RunStatus_PENDING, withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_CREATED)),
			&FollowOptions{Input: false})

		assert.Equal(t, OutcomeDetached, outcome.Kind)
		assert.Equal(t, "-input was false", outcome.LeaveReason)
	})

	t.Run("a stage with no checks this CLI can see waits for it to finish", func(t *testing.T) {
		// A stage holding only other kinds of tasks, which this CLI's API version doesn't carry.
		createdNoChecks := newTestRun(pb.RunStatus_PENDING, withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_CREATED))
		running := newTestRun(pb.RunStatus_PRE_APPLY_RUNNING, withAutoApply(), withPlan(pb.PlanStatus_FINISHED, "plan-job"),
			withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_RUNNING))
		blocked := newTestRun(pb.RunStatus_PRE_APPLY_AWAITING_DECISION, withAutoApply(), withPlan(pb.PlanStatus_FINISHED, "plan-job"),
			withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_AWAITING_OVERRIDE))
		done := appliedRun(withAutoApply(), withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_COMPLETED))
		h := newHarness(t, newFakeUI(true), running, running, blocked, blocked, done)

		outcome := h.follow(t, createdNoChecks, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, 0, h.ui.inputCount())
		out := h.ui.output()
		assert.Equal(t, 1, strings.Count(out, "This stage has no tasks this version of the CLI can show"))
		assert.Contains(t, out, "Waiting for the run to continue")
	})

	t.Run("a supported gate is still offered alongside an unsupported check", func(t *testing.T) {
		mixed := newTestRun(pb.RunStatus_PRE_APPLY_AWAITING_DECISION,
			withPlan(pb.PlanStatus_FINISHED, "plan-job"),
			withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_AWAITING_OVERRIDE,
				futureCheck(pb.PolicyCheckStatus_POLICY_CHECK_STATUS_SOFT_FAILED),
				newCheck("c2", pb.PolicyCheckStatus_POLICY_CHECK_STATUS_SOFT_FAILED, "check-job-2")))
		h := newHarness(t, newFakeUI(true, "approve"), mixed)
		h.runs.Results["c2"] = softFailedResults("c2", 1, 0)
		h.runs.OnApprove = func(f *runtest.Runs) *pb.RunGate {
			// The supported gate clears, but the unsupported check still blocks the stage for a while.
			f.Results["c2"].RunGate.Status = pb.RunGateStatus_RUN_GATE_STATUS_APPROVED
			f.Set(mixed, appliedRun(withStage(preApply, pb.RunTaskStageStatus_RUN_TASK_STAGE_STATUS_COMPLETED)))
			g := softFailedResults("c2", 1, 1).RunGate
			g.Status = pb.RunGateStatus_RUN_GATE_STATUS_APPROVED
			return g
		}
		// Stay blocked until the CLI is waiting, however many reads asking about the stage takes.
		h.runs.HoldUntil = func() bool { return strings.Contains(h.ui.output(), "Waiting for the run to continue") }

		outcome := h.follow(t, created, &FollowOptions{Input: true})

		assert.Equal(t, OutcomeCompleted, outcome.Kind)
		assert.Equal(t, 1, h.runs.Approved)
		assert.Equal(t, 1, h.ui.inputCount())
		out := h.ui.output()
		assert.Contains(t, out, "Approval recorded; all required approvals are in.")
		assert.Contains(t, out, "Waiting for the run to continue")
		assert.Equal(t, 1, strings.Count(out, "Policy checks"), "every check shares the stage's one policy checks section")
		assert.Contains(t, out, "Policy checks\n  ○ Unsupported policy check type (99); waiting for it to finish.\n\n  ! OPA policies failed")
	})
}

func TestRenderPolicyCheck(t *testing.T) {
	results := softFailedResults("c1", 1, 0)
	results.PolicyCheck.Status = pb.PolicyCheckStatus_POLICY_CHECK_STATUS_PASSED
	results.PolicyResults[0].Policy.EnforcementLevel = pb.PolicyEnforcementLevel_POLICY_ENFORCEMENT_LEVEL_ADVISORY
	results.PolicyResults = append(results.PolicyResults, &pb.PolicyCheckPolicyResult{Policy: &pb.PolicyCheckPolicy{
		Id:               "p2",
		EnforcementLevel: pb.PolicyEnforcementLevel_POLICY_ENFORCEMENT_LEVEL_HARD_MANDATORY,
		Status:           pb.PolicyCheckPolicyStatus_POLICY_CHECK_POLICY_STATUS_PASSED,
		Provenance:       &pb.PolicyCheckPolicyProvenance{PolicyTrn: "trn:policy:group/required-tags"},
	}})
	ui := newFakeUI(false)

	(&renderer{ui: ui}).policyCheck(results)

	assert.Equal(t, strings.Join([]string{
		"  ! OPA policies passed with 1 advisory failure",
		"    ✗ group/encrypted-buckets (advisory)",
		"        - bucket must be encrypted",
		"    ✓ group/required-tags (hard-mandatory)",
	}, "\n"), ui.output())
}

func TestRenderPolicyCheck_PartialResults(t *testing.T) {
	results := &pb.PolicyCheckResults{
		PolicyCheck: &pb.PolicyCheck{Id: "c1", CheckType: pb.PolicyCheckType_POLICY_CHECK_TYPE_OPA, Status: pb.PolicyCheckStatus_POLICY_CHECK_STATUS_PASSED},
		PolicyResults: []*pb.PolicyCheckPolicyResult{
			{Policy: &pb.PolicyCheckPolicy{Id: "p1", EnforcementLevel: pb.PolicyEnforcementLevel(42), Status: pb.PolicyCheckPolicyStatus(42)}},
			{},
		},
	}
	ui := newFakeUI(false)

	assert.NotPanics(t, func() { (&renderer{ui: ui}).policyCheck(results) })
	assert.Contains(t, ui.output(), "p1 (unknown) not evaluated")
	assert.NotPanics(t, func() { (&renderer{ui: ui}).policyCheck(&pb.PolicyCheckResults{}) })
}

package command

import (
	"context"
	"fmt"
	"os"
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

// TestInterruptContexts verifies the first signal cancels only stopCtx and the second cancelCtx.
func TestInterruptContexts(t *testing.T) {
	stopCtx, cancelCtx, release := interruptContexts(context.Background())
	defer release()

	// The handler is registered, so signaling ourselves is caught rather than killing the test.
	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGINT))
	waitDone(stopCtx, t)
	assert.NoError(t, cancelCtx.Err(), "the first interrupt leaves cancelCtx alone")

	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGINT))
	waitDone(cancelCtx, t)
	assert.Equal(t, os.Interrupt, run.InterruptSignal(stopCtx), "the signal is recorded as the cause")
	assert.Equal(t, interruptedExitCode, signalExitCode(stopCtx))
}

// TestInterruptContexts_SIGTERM verifies SIGTERM is recorded as the cause and exits 143.
func TestInterruptContexts_SIGTERM(t *testing.T) {
	stopCtx, _, release := interruptContexts(context.Background())
	defer release()

	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGTERM))
	waitDone(stopCtx, t)

	assert.Equal(t, syscall.SIGTERM, run.InterruptSignal(stopCtx))
	assert.Equal(t, terminatedExitCode, signalExitCode(stopCtx))
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
	stopCtx, cancelCtx, release := interruptContexts(context.Background())
	release()

	assert.Error(t, stopCtx.Err())
	assert.Error(t, cancelCtx.Err())
}

func waitDone(ctx context.Context, t *testing.T) {
	t.Helper()
	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("context was not canceled by the signal")
	}
}

// recordingUI captures Warnf output.
type recordingUI struct {
	terminal.UI
	warnings []string
}

func (u *recordingUI) Warnf(format string, a ...any) {
	u.warnings = append(u.warnings, fmt.Sprintf(format, a...))
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

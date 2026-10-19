package run

import (
	"context"
	"errors"
	"os"
	"syscall"
)

// SignalError is the cause a context is canceled with when a signal interrupts the CLI, so the code
// handling the interrupt can tell Ctrl-C (SIGINT) from SIGTERM.
type SignalError struct {
	Signal os.Signal
}

func (e *SignalError) Error() string {
	return "received " + e.Signal.String()
}

// InterruptSignal returns the signal that canceled ctx, or nil if ctx wasn't canceled by a signal.
func InterruptSignal(ctx context.Context) os.Signal {
	var se *SignalError
	if errors.As(context.Cause(ctx), &se) {
		return se.Signal
	}
	return nil
}

// isTerminated reports whether ctx was canceled by SIGTERM.
func isTerminated(ctx context.Context) bool {
	return InterruptSignal(ctx) == syscall.SIGTERM
}

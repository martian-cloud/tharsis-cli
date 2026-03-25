package terminal

import (
	"io"
)

type noopUI struct{}

// NewNoopUI returns a UI that discards all output and doesn't support interaction.
// Useful for testing or when UI output should be suppressed.
func NewNoopUI() UI {
	return &noopUI{}
}

// Input returns ErrNonInteractive since noop UI doesn't support interaction.
func (*noopUI) Input(*Input) (string, error) { return "", ErrNonInteractive }

// Interactive returns false since noop UI doesn't support interaction.
func (*noopUI) Interactive() bool { return false }

// Confirm returns ErrNonInteractive since noop UI doesn't support interaction.
func (*noopUI) Confirm(string) (bool, error) { return false, ErrNonInteractive }

// Output discards all output.
func (*noopUI) Output(string, ...any) {}

// NamedValues discards all output.
func (*noopUI) NamedValues([]NamedValue, ...Option) {}

// OutputWriters returns writers that discard all output.
func (*noopUI) OutputWriters() (io.Writer, io.Writer, error) {
	return io.Discard, io.Discard, nil
}

// Status returns a noop status that discards all updates.
func (*noopUI) Status() Status { return &noopStatus{} }

// Table discards all output.
func (*noopUI) Table(*Table, ...Option) {}

// StepGroup returns a noop step group that discards all output.
func (*noopUI) StepGroup() StepGroup { return &noopStepGroup{} }

// Successf discards all output.
func (*noopUI) Successf(string, ...any) {}

// Errorf discards all output.
func (*noopUI) Errorf(string, ...any) {}

// ErrorWithSummary discards all output.
func (*noopUI) ErrorWithSummary(error, string, ...any) {}

// Warnf discards all output.
func (*noopUI) Warnf(string, ...any) {}

// Infof discards all output.
func (*noopUI) Infof(string, ...any) {}

// JSON discards all output.
func (*noopUI) JSON(any) error { return nil }

// Close is a no-op for noop UI.
func (*noopUI) Close() error { return nil }

type noopStatus struct{}

func (*noopStatus) Update(string)       {}
func (*noopStatus) Step(string, string) {}
func (*noopStatus) Close() error        { return nil }

type noopStepGroup struct{}

func (*noopStepGroup) Add(string, ...any) Step { return &noopStep{} }
func (*noopStepGroup) Wait()                   {}

type noopStep struct{}

func (*noopStep) TermOutput() io.Writer { return io.Discard }
func (*noopStep) Update(string, ...any) {}
func (*noopStep) Status(string)         {}
func (*noopStep) Done()                 {}
func (*noopStep) Abort()                {}

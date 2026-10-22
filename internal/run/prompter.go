package run

import (
	"context"
	"fmt"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
)

// Prompter asks the user for decisions while a run is followed.
//
// A read from stdin can't be canceled, so at most one read is ever in flight. When a prompt is
// abandoned (its context is canceled before an answer), its read stays in flight and the next prompt
// picks it up. An answer read by a read that started for an earlier prompt is thrown away, even if
// it arrives after the current prompt is showing: the user may have been typing it in answer to the
// earlier question, such as a gate that has since been re-evaluated.
type Prompter struct {
	ui terminal.UI
	// input is false with -input=false; the Prompter then never reads.
	input bool
	// terminal is true when stdin is a terminal.
	terminal bool
	// prompts counts the prompts shown; each read records the prompt it was started for.
	prompts int
	pending *pendingRead
}

type pendingRead struct {
	ch chan inputResult
	// prompt is the number of the prompt the read was started for.
	prompt int
}

type inputResult struct {
	answer string
	err    error
}

// NewPrompter returns a Prompter. It never prompts when input is false.
func NewPrompter(ui terminal.UI, input bool) *Prompter {
	return &Prompter{
		ui:       ui,
		input:    input,
		terminal: ui.Interactive(),
	}
}

// Interactive reports whether a user at a terminal can be asked: input is enabled and stdin is a
// terminal.
func (p *Prompter) Interactive() bool {
	return p.input && p.terminal
}

// Terminal reports whether stdin is a terminal, whether or not input is enabled.
func (p *Prompter) Terminal() bool {
	return p.terminal
}

// AcceptsInput reports whether input is enabled at all. Unlike Interactive, it holds when stdin is
// piped, so an answer such as `yes | tharsis apply` can be read. Ask still fails with
// terminal.ErrNonInteractive when the UI can't read (stdout isn't a terminal).
func (p *Prompter) AcceptsInput() bool {
	return p.input
}

// cannotAskReason says why the user isn't being asked, for when Interactive is false.
func (p *Prompter) cannotAskReason() string {
	if !p.input {
		return "-input was false"
	}
	return "the terminal is not interactive"
}

// Ask shows prompt and waits for an answer, until ctx is canceled. A failed read is returned
// wrapped, so errors.Is still matches terminal.ErrNonInteractive and io.EOF.
func (p *Prompter) Ask(ctx context.Context, prompt string) (string, error) {
	if !p.input {
		return "", terminal.ErrNonInteractive
	}

	p.prompts++
	current := p.prompts
	p.read(prompt, current)

	for {
		// A finished read takes priority over cancellation, so a read that has already failed or been
		// answered isn't lost to a race with the run moving on.
		var r inputResult
		select {
		case r = <-p.pending.ch:
		default:
			select {
			case r = <-p.pending.ch:
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}

		readFor := p.pending.prompt
		p.pending = nil
		if r.err != nil {
			// Reading failed (no terminal, or stdin is closed); asking again won't help.
			return "", fmt.Errorf("failed to read answer: %w", r.err)
		}
		if readFor != current {
			// Read for a prompt that is no longer showing, so it may not answer this one.
			p.ui.Output("  The question changed while that answer was being typed; please answer again.")
			p.read(prompt, current)
			continue
		}
		return r.answer, nil
	}
}

// read starts a read for the given prompt unless one is already in flight.
func (p *Prompter) read(prompt string, number int) {
	if p.pending != nil {
		// The read from an abandoned prompt is still waiting; show the new prompt text.
		p.ui.Output(prompt)
		return
	}

	ch := make(chan inputResult, 1)
	p.pending = &pendingRead{ch: ch, prompt: number}
	go func() {
		answer, err := p.ui.Input(&terminal.Input{Prompt: prompt})
		ch <- inputResult{answer: answer, err: err}
	}()
}

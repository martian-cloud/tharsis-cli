package run

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
)

func TestPrompter(t *testing.T) {
	// abandon asks and cancels the prompt once its read is in flight, as when the run moves on.
	abandon := func(t *testing.T, p *Prompter, ui *fakeUI, prompt string, readsBefore int) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			require.Eventually(t, func() bool { return ui.inputCount() == readsBefore+1 }, time.Second, time.Millisecond)
			cancel()
		}()
		_, err := p.Ask(ctx, prompt)
		require.ErrorIs(t, err, context.Canceled)
	}

	t.Run("returns the answer", func(t *testing.T) {
		ui := newFakeUI(true, "yes")
		p := NewPrompter(ui, true)

		answer, err := p.Ask(context.Background(), "approve? ")

		require.NoError(t, err)
		assert.Equal(t, "yes", answer)
	})

	t.Run("never reads input with -input=false", func(t *testing.T) {
		ui := newFakeUI(true, "yes")
		p := NewPrompter(ui, false)

		assert.False(t, p.Interactive())
		assert.True(t, p.Terminal())
		assert.False(t, p.AcceptsInput())
		_, err := p.Ask(context.Background(), "approve? ")
		assert.ErrorIs(t, err, terminal.ErrNonInteractive)
		assert.Equal(t, 0, ui.inputCount())
	})

	t.Run("reads piped stdin", func(t *testing.T) {
		ui := newFakeUI(false, "yes")
		p := NewPrompter(ui, true)

		assert.False(t, p.Interactive())
		assert.True(t, p.AcceptsInput())
		answer, err := p.Ask(context.Background(), "approve? ")
		require.NoError(t, err)
		assert.Equal(t, "yes", answer)
	})

	t.Run("a failed read is returned", func(t *testing.T) {
		ui := newFakeUI(false)
		ui.inputErr = io.EOF
		p := NewPrompter(ui, true)

		_, err := p.Ask(context.Background(), "approve? ")

		assert.ErrorIs(t, err, io.EOF)
	})

	t.Run("keeps at most one read in flight and discards answers to an abandoned prompt", func(t *testing.T) {
		ui := newFakeUI(true)
		p := NewPrompter(ui, true)
		abandon(t, p, ui, "first? ", 0)

		// The answer finishes only once the second prompt is showing, but its read began earlier.
		go func() {
			ui.answers <- "approve"
			ui.answers <- "discard"
		}()
		answer, err := p.Ask(context.Background(), "second? ")

		require.NoError(t, err)
		assert.Equal(t, "discard", answer)
		assert.Equal(t, 2, ui.inputCount(), "a new read starts only after the stale one completes")
		assert.Contains(t, ui.output(), "The question changed while that answer was being typed; please answer again.")
	})

	t.Run("stops when the context is canceled", func(t *testing.T) {
		ui := newFakeUI(true)
		t.Cleanup(func() { close(ui.answers) })
		p := NewPrompter(ui, true)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := p.Ask(ctx, "approve? ")

		assert.ErrorIs(t, err, context.Canceled)
	})
}

package terminal

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/containerd/console"
	"github.com/fatih/color"
	"github.com/lab47/vterm/parser"
	"github.com/lab47/vterm/screen"
	"github.com/lab47/vterm/state"
	"github.com/morikuni/aec"
)

var spinnerSet = spinner.CharSets[11]

// DisplayEntry represents a single entry in the display.
type DisplayEntry struct {
	d       *Display
	line    uint
	spinner bool
	text    string
	status  string

	body []string
}

// Display manages terminal output with live updating capabilities.
type Display struct {
	mu      sync.Mutex
	Entries []*DisplayEntry

	w       io.Writer
	newEnt  chan *DisplayEntry
	updates chan *DisplayEntry
	resize  chan struct{} // sent to when an entry has resized itself.
	line    uint
	width   int

	wg       sync.WaitGroup
	spinning int
}

// NewDisplay creates a new Display for the given writer.
func NewDisplay(ctx context.Context, w io.Writer) *Display {
	d := &Display{
		w:       w,
		width:   80,
		updates: make(chan *DisplayEntry),
		resize:  make(chan struct{}),
		newEnt:  make(chan *DisplayEntry),
	}

	if f, ok := w.(*os.File); ok {
		if c, err := console.ConsoleFromFile(f); err == nil {
			if sz, err := c.Size(); err == nil {
				if sz.Width >= 10 {
					d.width = int(sz.Width) - 1
				}
			}
		}
	}

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.Display(ctx)
	}()

	return d
}

// Close waits for the display to finish.
func (d *Display) Close() error {
	d.wg.Wait()
	return nil
}

func (d *Display) flushAll() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for range d.Entries {
		fmt.Fprintln(d.w, "")
	}

	d.line = uint(len(d.Entries))
}

func (d *Display) renderEntry(ent *DisplayEntry, spin int) {
	b := aec.EmptyBuilder

	diff := d.line - ent.line

	text := strings.TrimRight(ent.text, " \t\n")

	if len(text) >= d.width {
		text = text[:d.width-1]
	}

	prefix := ""
	if ent.spinner {
		prefix = spinnerSet[spin] + " "
	}

	var statusColor *aec.Builder
	if ent.status != "" {
		icon, ok := statusIcons[ent.status]
		if !ok {
			icon = ent.status
		}

		if len(prefix) > 0 {
			prefix = prefix + " " + icon + " "
		} else {
			prefix = icon + " "
		}

		if !color.NoColor {
			if codes, ok := colorStatus[ent.status]; ok {
				statusColor = b.With(codes...)
			}
		}
	}

	line := fmt.Sprintf("%s%s%s",
		b.
			Up(diff).
			Column(0).
			EraseLine(aec.EraseModes.All).
			ANSI,
		prefix,
		text,
	)

	if statusColor != nil {
		line = statusColor.ANSI.Apply(line)
	}

	fmt.Fprint(d.w, line)

	for _, body := range ent.body {
		fmt.Fprintf(d.w, "%s%s",
			b.
				Down(1).
				Column(0).
				ANSI,
			body,
		)
		diff--
	}

	fmt.Fprintf(d.w, "%s",
		b.
			Down(diff).
			Column(0).
			ANSI,
	)
}

// Display runs the display loop.
func (d *Display) Display(ctx context.Context) {
	// d.flushAll()

	ticker := time.NewTicker(time.Second / 6)

	var spin int

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			spin++
			if spin >= len(spinnerSet) {
				spin = 0
			}

			d.mu.Lock()
			update := d.spinning > 0

			if !update {
				d.mu.Unlock()
				continue
			}

			for _, ent := range d.Entries {
				if !ent.spinner {
					continue
				}

				d.renderEntry(ent, spin)
			}

			d.mu.Unlock()
		case ent := <-d.newEnt:
			d.mu.Lock()
			ent.line = d.line
			d.Entries = append(d.Entries, ent)
			d.line++
			d.line += uint(len(ent.body))
			fmt.Fprintln(d.w, "")
			for i := 0; i < len(ent.body); i++ {
				fmt.Fprintln(d.w, "")
			}

			d.mu.Unlock()

		case ent := <-d.updates:
			d.mu.Lock()
			d.renderEntry(ent, spin)
			d.mu.Unlock()
		case <-d.resize:
			d.mu.Lock()

			var newLine uint

			for _, ent := range d.Entries {
				newLine++
				newLine += uint(len(ent.body))
			}

			diff := newLine - d.line

			// TODO should we support shrinking?
			if diff > 0 {
				// Pad down
				for i := uint(0); i < diff; i++ {
					fmt.Fprintln(d.w, "")
				}

				d.line = newLine

				var cnt uint

				for _, ent := range d.Entries {
					ent.line = cnt
					cnt++
					cnt += uint(len(ent.body))

					d.renderEntry(ent, spin)
				}
			}

			d.mu.Unlock()
		}
	}
}

// NewStatus creates a new status entry.
func (d *Display) NewStatus() *DisplayEntry {
	de := &DisplayEntry{
		d: d,
	}

	d.newEnt <- de

	return de
}

// NewStatusWithBody creates a new status entry with body lines.
func (d *Display) NewStatusWithBody(lines int) *DisplayEntry {
	de := &DisplayEntry{
		d:    d,
		body: make([]string, lines),
	}

	d.newEnt <- de

	return de
}

// StartSpinner starts the spinner animation.
func (e *DisplayEntry) StartSpinner() {
	e.d.mu.Lock()

	e.spinner = true
	e.d.spinning++

	e.d.mu.Unlock()

	e.d.updates <- e
}

// StopSpinner stops the spinner animation.
func (e *DisplayEntry) StopSpinner() {
	e.d.mu.Lock()

	e.spinner = false
	e.d.spinning--

	e.d.mu.Unlock()

	e.d.updates <- e
}

// SetStatus sets the status of the entry.
func (e *DisplayEntry) SetStatus(status string) {
	e.d.mu.Lock()
	defer e.d.mu.Unlock()

	e.status = status
}

// Update updates the entry text.
func (e *DisplayEntry) Update(str string, args ...interface{}) {
	e.d.mu.Lock()
	e.text = fmt.Sprintf(str, args...)
	e.d.mu.Unlock()

	e.d.updates <- e
}

// SetBody sets a body line.
func (e *DisplayEntry) SetBody(line int, data string) {
	e.d.mu.Lock()

	var resize bool

	if line >= len(e.body) {
		nb := make([]string, line+1)
		copy(nb, e.body)

		e.body = nb
		resize = true
	}

	e.body[line] = data
	e.d.mu.Unlock()

	if resize {
		e.d.resize <- struct{}{}
	}

	e.d.updates <- e
}

// Term provides terminal emulation for step output.
type Term struct {
	ent    *DisplayEntry
	scr    *screen.Screen
	w      io.Writer
	ctx    context.Context
	cancel func()

	output [][]rune

	wg       sync.WaitGroup
	parseErr error
}

// DamageDone handles screen damage events.
func (t *Term) DamageDone(r state.Rect, cr screen.CellReader) error {
	for row := r.Start.Row; row <= r.End.Row; row++ {
		for col := r.Start.Col; col <= r.End.Col; col++ {
			cell := cr.GetCell(row, col)

			if cell == nil {
				t.output[row][col] = ' '
			} else {
				val, _ := cell.Value()

				if val == 0 {
					t.output[row][col] = ' '
				} else {
					t.output[row][col] = val
				}
			}
		}
	}

	for row := r.Start.Row; row <= r.End.Row; row++ {
		b := aec.EmptyBuilder
		if color.NoColor {
			t.ent.SetBody(row, fmt.Sprintf(" │ %s", string(t.output[row])))
		} else {
			blue := b.LightBlueF()
			t.ent.SetBody(row, fmt.Sprintf(" │ %s%s%s", blue.ANSI, string(t.output[row]), aec.Reset))
		}
	}

	return nil
}

// MoveCursor handles cursor movement (ignored).
func (t *Term) MoveCursor(_ state.Pos) error {
	return nil
}

// SetTermProp handles terminal property changes (ignored).
func (t *Term) SetTermProp(_ state.TermAttr, _ interface{}) error {
	return nil
}

// Output handles output events (ignored).
func (t *Term) Output(_ []byte) error {
	return nil
}

// StringEvent handles string events (ignored).
func (t *Term) StringEvent(_ string, _ []byte) error {
	return nil
}

// NewTerm creates a new terminal emulator.
func NewTerm(ctx context.Context, d *DisplayEntry, height, width int) (*Term, error) {
	term := &Term{
		ent:    d,
		output: make([][]rune, height),
	}

	for i := range term.output {
		term.output[i] = make([]rune, width)
	}

	scr, err := screen.NewScreen(height, width, term)
	if err != nil {
		return nil, err
	}

	term.scr = scr

	st, err := state.NewState(height, width, scr)
	if err != nil {
		return nil, err
	}

	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	term.w = w

	prs, err := parser.NewParser(r, st)
	if err != nil {
		return nil, err
	}

	term.ctx, term.cancel = context.WithCancel(ctx)

	term.wg.Add(1)
	go func() {
		defer term.wg.Done()

		err := prs.Drive(term.ctx)
		if err != nil && err != context.Canceled {
			term.parseErr = err
		}
	}()

	return term, nil
}

// Write writes to the terminal.
func (t *Term) Write(b []byte) (int, error) {
	return t.w.Write(b)
}

// Close closes the terminal.
func (t *Term) Close() error {
	t.cancel()
	t.wg.Wait()
	return t.parseErr
}

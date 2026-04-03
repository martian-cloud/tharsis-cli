package terminal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/bgentry/speakeasy"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
)

type basicUI struct {
	ctx    context.Context
	status *spinnerStatus
}

// ConsoleUI returns a UI which will write to the current processes stdout/stderr.
func ConsoleUI(ctx context.Context) UI {
	// We do both of these checks because some sneaky environments fool
	// one or the other and we really only want the interactive UI in
	// truly interactive environments.
	if isatty.IsTerminal(os.Stdout.Fd()) && term.IsTerminal(int(os.Stdout.Fd())) {
		return &basicUI{ctx: ctx}
	}
	return NonInteractiveUI(ctx)
}

// Input implements UI
func (ui *basicUI) Input(input *Input) (string, error) {
	var buf bytes.Buffer

	// Write the prompt, add a space.
	ui.Output(input.Prompt, WithStyle(input.Style), WithWriter(&buf))
	fmt.Fprint(color.Output, strings.TrimRight(buf.String(), "\r\n"))
	fmt.Fprint(color.Output, " ")

	// Ask for input in a go-routine so that we can ignore it.
	errCh := make(chan error, 1)
	lineCh := make(chan string, 1)
	go func() {
		var line string
		var err error
		if input.Secret && isatty.IsTerminal(os.Stdin.Fd()) {
			line, err = speakeasy.Ask("")
		} else {
			r := bufio.NewReader(os.Stdin)
			line, err = r.ReadString('\n')
		}
		if err != nil {
			errCh <- err
			return
		}

		lineCh <- strings.TrimRight(line, "\r\n")
	}()

	select {
	case err := <-errCh:
		return "", err
	case line := <-lineCh:
		return line, nil
	case <-ui.ctx.Done():
		fmt.Fprintln(color.Output)
		return "", ui.ctx.Err()
	}
}

func (ui *basicUI) Interactive() bool {
	return isatty.IsTerminal(os.Stdin.Fd())
}

func (ui *basicUI) Output(msg string, raw ...interface{}) {
	msg, style, w := Interpret(msg, raw...)

	switch style {
	case HeaderStyle:
		msg = colorHeader.Sprintf("\n==> %s", msg)
	case ErrorStyle:
		msg = colorError.Sprint(msg)
	case ErrorBoldStyle:
		msg = colorErrorBold.Sprint(msg)
	case WarningStyle:
		msg = colorWarning.Sprint(msg)
	case WarningBoldStyle:
		msg = colorWarningBold.Sprint(msg)
	case SuccessStyle:
		msg = colorSuccess.Sprint(msg)
	case SuccessBoldStyle:
		msg = colorSuccessBold.Sprint(msg)
	case InfoStyle:
		lines := strings.Split(msg, "\n")
		for i, line := range lines {
			lines[i] = colorInfo.Sprintf("    %s", line)
		}
		msg = strings.Join(lines, "\n")
	}

	st := ui.status
	if st != nil {
		if st.Pause() {
			defer st.Start()
		}
	}

	fmt.Fprintln(w, msg)
}

func (ui *basicUI) NamedValues(rows []NamedValue, opts ...Option) {
	cfg := &config{Writer: color.Output}
	for _, opt := range opts {
		opt(cfg)
	}

	var buf bytes.Buffer
	tr := tabwriter.NewWriter(&buf, 1, 8, 0, ' ', tabwriter.AlignRight)
	for _, row := range rows {
		switch v := row.Value.(type) {
		case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
			fmt.Fprintf(tr, "  %s: \t%d\n", row.Name, row.Value)
		case float32, float64:
			fmt.Fprintf(tr, "  %s: \t%f\n", row.Name, row.Value)
		case bool:
			fmt.Fprintf(tr, "  %s: \t%v\n", row.Name, row.Value)
		case string:
			if v == "" {
				continue
			}
			fmt.Fprintf(tr, "  %s: \t%s\n", row.Name, row.Value)
		default:
			fmt.Fprintf(tr, "  %s: \t%s\n", row.Name, row.Value)
		}
	}

	tr.Flush()
	colorInfo.Fprint(cfg.Writer, buf.String())
}

func (ui *basicUI) OutputWriters() (io.Writer, io.Writer, error) {
	return os.Stdout, os.Stderr, nil
}

func (ui *basicUI) Status() Status {
	if ui.status == nil {
		ui.status = newSpinnerStatus(ui.ctx)
	}
	return ui.status
}

func (ui *basicUI) StepGroup() StepGroup {
	ctx, cancel := context.WithCancel(ui.ctx)
	display := NewDisplay(ctx, color.Output)

	return &fancyStepGroup{
		ctx:     ctx,
		cancel:  cancel,
		display: display,
		done:    make(chan struct{}),
	}
}

func (ui *basicUI) Table(tbl *Table, opts ...Option) {
	cfg := &config{Writer: color.Output}
	for _, opt := range opts {
		opt(cfg)
	}

	table := tablewriter.NewWriter(cfg.Writer)
	table.SetHeader(tbl.Headers)
	table.SetBorder(false)
	table.SetAutoWrapText(true)

	for _, row := range tbl.Rows {
		colors := make([]tablewriter.Colors, len(row))
		entries := make([]string, len(row))

		for i, ent := range row {
			entries[i] = ent.Value
			if !color.NoColor {
				if c, ok := colorMapping[ent.Color]; ok {
					colors[i] = tablewriter.Colors{c}
				}
			}
		}

		table.Rich(entries, colors)
	}

	table.Render()
}

func (ui *basicUI) Confirm(prompt string) (bool, error) {
	result, err := ui.Input(&Input{Prompt: prompt + " [y/N]:"})
	if err != nil {
		return false, err
	}
	result = strings.ToLower(strings.TrimSpace(result))
	return result == "y" || result == "yes", nil
}

func (ui *basicUI) Close() error {
	return nil
}

func (ui *basicUI) Successf(format string, a ...any) {
	fmt.Fprintln(color.Output, colorSuccess.Sprintf(format, a...))
}

func (ui *basicUI) Errorf(format string, a ...any) {
	fmt.Fprint(color.Output, formatError(nil, format, a...))
}

func (ui *basicUI) ErrorWithSummary(err error, summary string, a ...any) {
	fmt.Fprint(color.Output, formatError(err, summary, a...))
}

func (ui *basicUI) Warnf(format string, a ...any) {
	fmt.Fprintln(color.Output, colorWarning.Sprintf(format, a...))
}

func (ui *basicUI) Infof(format string, a ...any) {
	fmt.Fprintln(color.Output, color.New(color.FgCyan).Sprintf(format, a...))
}

func (ui *basicUI) JSON(v any) error {
	buf, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}

	if !color.NoColor {
		var highlighted bytes.Buffer
		if err := quick.Highlight(&highlighted, string(buf), "json", "terminal16m", "monokai"); err == nil {
			fmt.Fprintln(color.Output, highlighted.String())
			return nil
		}
	}

	fmt.Fprintln(color.Output, string(buf))
	return nil
}

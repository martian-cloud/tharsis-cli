package terminal

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/morikuni/aec"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal/spinner"
)

// Status constants for step completion states.
const (
	StatusOK      = "ok"
	StatusError   = "error"
	StatusWarn    = "warn"
	StatusTimeout = "timeout"
	StatusAbort   = "abort"
)

var emojiStatus = map[string]string{
	StatusOK:      "✓",
	StatusError:   "✗",
	StatusWarn:    "⚠",
	StatusTimeout: "⏱",
	StatusAbort:   "⊘",
}

var textStatus = map[string]string{
	StatusOK:      " +",
	StatusError:   " X",
	StatusWarn:    " !",
	StatusTimeout: "<>",
	StatusAbort:   " -",
}

var colorStatus = map[string][]aec.ANSI{
	StatusOK:    {aec.GreenF},
	StatusError: {aec.RedF},
	StatusWarn:  {aec.YellowF},
	StatusAbort: {aec.RedF},
}

// RegisterStatus registers a custom status with emoji, text fallback, and color.
// This allows packages to define their own status types.
func RegisterStatus(name, emoji, text string, color aec.ANSI) {
	emojiStatus[name] = emoji
	textStatus[name] = text
	colorStatus[name] = []aec.ANSI{color}
}

// Status is used to provide an updating status to the user. The status
// usually has some animated element along with it such as a spinner.
type Status interface {
	// Update writes a new status. This should be a single line.
	Update(msg string)

	// Indicate that a step has finished, confering an ok, error, or warn upon
	// it's finishing state. If the status is not StatusOK, StatusError, or StatusWarn
	// then the status text is written directly to the output, allowing for custom
	// statuses.
	Step(status, msg string)

	// Close should be called when the live updating is complete. The
	// status will be cleared from the line.
	Close() error
}

// spinnerStatus implements Status and uses a spinner to show updates.
type spinnerStatus struct {
	mu      sync.Mutex
	spinner *spinner.Spinner
	running bool
}

var statusIcons map[string]string

func init() {
	if strings.Contains(os.Getenv("LANG"), "UTF-8") {
		statusIcons = emojiStatus
	} else {
		statusIcons = textStatus
	}
}

func newSpinnerStatus(ctx context.Context) *spinnerStatus {
	return &spinnerStatus{
		spinner: spinner.New(
			ctx,
			spinner.CharSets[11],
			time.Second/6,
			spinner.WithColor("bold"),
		),
	}
}

func (s *spinnerStatus) Update(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.spinner.Suffix = " " + msg

	if !s.running {
		s.spinner.Start()
		s.running = true
	}
}

func (s *spinnerStatus) Step(status, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.spinner.Stop()
	s.running = false

	pad := ""

	statusIcon := emojiStatus[status]
	if statusIcon == "" {
		statusIcon = status
	} else if status == StatusWarn {
		pad = " "
	}

	fmt.Fprintf(color.Output, "%s%s %s\n", statusIcon, pad, msg)
}

func (s *spinnerStatus) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		s.running = false
		s.spinner.Suffix = ""
	}

	s.spinner.Stop()

	return nil
}

func (s *spinnerStatus) Pause() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	wasRunning := s.running

	if s.running {
		s.running = false
		s.spinner.Stop()
	}

	return wasRunning
}

func (s *spinnerStatus) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		s.running = true
		s.spinner.Start()
	}
}

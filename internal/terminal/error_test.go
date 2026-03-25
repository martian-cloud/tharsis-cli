package terminal

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFormatError(t *testing.T) {
	t.Run("formats message with error bar", func(t *testing.T) {
		result := formatError(nil, "something went wrong")
		if !strings.Contains(result, "│") {
			t.Error("expected error bar in output")
		}
		if !strings.Contains(result, "Error:") {
			t.Error("expected 'Error:' label")
		}
		if !strings.Contains(result, "something went wrong") {
			t.Error("expected message in output")
		}
	})

	t.Run("includes wrapped error message", func(t *testing.T) {
		err := errors.New("connection refused")
		result := formatError(err, "failed to connect")
		if !strings.Contains(result, "failed to connect") {
			t.Error("expected format message")
		}
		if !strings.Contains(result, "connection refused") {
			t.Error("expected wrapped error message")
		}
	})

	t.Run("handles gRPC status errors", func(t *testing.T) {
		err := status.Error(codes.NotFound, "resource not found")
		result := formatError(err, "lookup failed")
		if !strings.Contains(result, "Code: NotFound") {
			t.Error("expected gRPC status code")
		}
		if !strings.Contains(result, "resource not found") {
			t.Error("expected gRPC error message")
		}
	})

	t.Run("handles wrapped gRPC status errors", func(t *testing.T) {
		grpcErr := status.Error(codes.InvalidArgument, "no matching version found")
		err := fmt.Errorf("failed to create run: %w", grpcErr)
		result := formatError(err, "operation failed")
		if !strings.Contains(result, "Code: InvalidArgument") {
			t.Error("expected gRPC status code")
		}
		if !strings.Contains(result, "failed to create run") {
			t.Error("expected outer error message")
		}
		if !strings.Contains(result, "no matching version found") {
			t.Error("expected inner gRPC error message")
		}
		if strings.Contains(result, "rpc error") {
			t.Error("should not contain raw rpc error prefix")
		}
	})

	t.Run("formats with arguments", func(t *testing.T) {
		result := formatError(nil, "failed after %d attempts on %s", 3, "server1")
		if !strings.Contains(result, "failed after 3 attempts on server1") {
			t.Error("expected formatted message with args")
		}
	})

	t.Run("wraps long error messages", func(t *testing.T) {
		longMsg := strings.Repeat("error ", 50)
		err := errors.New(longMsg)
		result := formatError(err, "operation failed")
		lines := strings.Split(result, "\n")
		// Should have multiple lines due to wrapping
		if len(lines) < 5 {
			t.Error("expected long message to be wrapped into multiple lines")
		}
	})

	t.Run("newline in format message gets bar on every line", func(t *testing.T) {
		result := formatError(nil, "line one\nline two\nline three")
		assertAllContentLinesHaveBar(t, result)
		if !strings.Contains(result, "line one") || !strings.Contains(result, "line two") || !strings.Contains(result, "line three") {
			t.Error("expected all format parts in output")
		}
		// Newlines in the header should be collapsed to spaces, producing a single Error: line.
		if strings.Count(result, "Error:") != 1 {
			t.Error("expected single Error: label for header")
		}
	})

	t.Run("newline in error message gets bar on every line", func(t *testing.T) {
		err := errors.New("first\nsecond\nthird")
		result := formatError(err, "multi-line error")
		assertAllContentLinesHaveBar(t, result)
		if !strings.Contains(result, "first") || !strings.Contains(result, "second") || !strings.Contains(result, "third") {
			t.Error("expected all error lines in output")
		}
	})

	t.Run("newlines in both format and error message", func(t *testing.T) {
		err := errors.New("err line 1\nerr line 2")
		result := formatError(err, "fmt line 1\nfmt line 2")
		assertAllContentLinesHaveBar(t, result)
	})

	t.Run("trailing newline in error message", func(t *testing.T) {
		err := errors.New("message with trailing newline\n") //revive:disable-line
		result := formatError(err, "test")
		assertAllContentLinesHaveBar(t, result)
	})

	t.Run("only newlines in error message", func(t *testing.T) {
		err := errors.New("\n\n\n") //revive:disable-line
		result := formatError(err, "test")
		assertAllContentLinesHaveBar(t, result)
	})

	t.Run("empty error message", func(t *testing.T) {
		err := errors.New("")
		result := formatError(err, "test")
		assertAllContentLinesHaveBar(t, result)
	})

	t.Run("empty format message", func(t *testing.T) {
		result := formatError(nil, "")
		assertAllContentLinesHaveBar(t, result)
	})

	t.Run("gRPC error with newlines in message", func(t *testing.T) {
		err := status.Error(codes.InvalidArgument, "field A is invalid\nfield B is required")
		result := formatError(err, "validation failed")
		assertAllContentLinesHaveBar(t, result)
		if !strings.Contains(result, "Code: InvalidArgument") {
			t.Error("expected gRPC status code")
		}
	})

	t.Run("consecutive newlines preserve paragraph gap", func(t *testing.T) {
		err := errors.New("paragraph one\n\nparagraph two")
		result := formatError(err, "test")
		assertAllContentLinesHaveBar(t, result)
		if !strings.Contains(result, "paragraph one") || !strings.Contains(result, "paragraph two") {
			t.Error("expected both paragraphs in output")
		}
	})

	t.Run("leading newline in error message", func(t *testing.T) {
		err := errors.New("\nstarts after newline")
		result := formatError(err, "test")
		assertAllContentLinesHaveBar(t, result)
		if !strings.Contains(result, "starts after newline") {
			t.Error("expected message content in output")
		}
	})

	t.Run("leading newline in format message", func(t *testing.T) {
		result := formatError(nil, "\nstarts after newline")
		assertAllContentLinesHaveBar(t, result)
		if !strings.Contains(result, "starts after newline") {
			t.Error("expected message content in output")
		}
		if strings.Count(result, "Error:") != 1 {
			t.Error("expected single Error: label for header")
		}
	})

	t.Run("carriage return in error message", func(t *testing.T) {
		err := errors.New("line one\r\nline two")
		result := formatError(err, "test")
		assertAllContentLinesHaveBar(t, result)
	})

	t.Run("long message with embedded newlines wraps each paragraph", func(t *testing.T) {
		long := strings.Repeat("word ", 30)
		err := errors.New(long + "\n" + long)
		result := formatError(err, "test")
		assertAllContentLinesHaveBar(t, result)
		lines := strings.Split(strings.TrimSpace(result), "\n")
		// Two long paragraphs should produce more lines than two.
		if len(lines) < 6 {
			t.Error("expected both paragraphs to be wrapped into multiple lines")
		}
	})

	t.Run("format message with percent literal", func(t *testing.T) {
		result := formatError(nil, "100%% complete")
		assertAllContentLinesHaveBar(t, result)
		if !strings.Contains(result, "100% complete") {
			t.Error("expected percent literal in output")
		}
	})

	t.Run("wrapped error chain with newlines at each level", func(t *testing.T) {
		inner := errors.New("root\ncause")
		outer := fmt.Errorf("outer\nwrapper: %w", inner)
		result := formatError(outer, "failed")
		assertAllContentLinesHaveBar(t, result)
	})

	t.Run("nil error with multiline format", func(t *testing.T) {
		result := formatError(nil, "line 1\nline 2\nline 3")
		assertAllContentLinesHaveBar(t, result)
		// Newlines in the header are collapsed, so only one Error: label.
		if strings.Count(result, "Error:") != 1 {
			t.Error("expected single Error: label for header")
		}
	})
}

// assertAllContentLinesHaveBar verifies that every non-empty line between
// the leading and trailing blank lines has the error bar prefix.
func assertAllContentLinesHaveBar(t *testing.T, output string) {
	t.Helper()
	for i, line := range strings.Split(strings.TrimPrefix(strings.TrimSuffix(output, "\n"), "\n"), "\n") {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "│") {
			t.Errorf("line %d missing bar prefix: %q", i, line)
		}
	}
}

func TestBuildErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "plain error",
			err:      errors.New("something failed"),
			expected: "something failed",
		},
		{
			name:     "gRPC status error",
			err:      status.Error(codes.NotFound, "resource not found"),
			expected: "resource not found",
		},
		{
			name:     "wrapped gRPC error",
			err:      fmt.Errorf("failed to create run: %w", status.Error(codes.InvalidArgument, "no matching version found")),
			expected: "failed to create run: no matching version found",
		},
		{
			name:     "deeply wrapped gRPC error",
			err:      fmt.Errorf("outer: %w", fmt.Errorf("middle: %w", status.Error(codes.Internal, "db connection lost"))),
			expected: "outer: middle: db connection lost",
		},
		{
			name:     "wrapped plain errors",
			err:      fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", errors.New("root cause"))),
			expected: "outer: inner: root cause",
		},
		{
			name:     "single wrapped error",
			err:      fmt.Errorf("wrapper: %w", errors.New("cause")),
			expected: "wrapper: cause",
		},
		{
			name:     "gRPC error with empty message",
			err:      status.Error(codes.Unknown, ""),
			expected: "",
		},
		{
			name:     "nil inner from unwrap",
			err:      errors.New("no wrapping here"),
			expected: "no wrapping here",
		},
		{
			// %v doesn't create an unwrap chain, so the raw gRPC prefix leaks through.
			name:     "gRPC error wrapped with percent v is not cleaned",
			err:      fmt.Errorf("failed to create run: %v", status.Error(codes.InvalidArgument, "bad input")),
			expected: "failed to create run: rpc error: code = InvalidArgument desc = bad input",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "error message containing colon but no wrapping",
			err:      errors.New("invalid value: must be positive"),
			expected: "invalid value: must be positive",
		},
		{
			name:     "error with newlines in message",
			err:      errors.New("line one\nline two"),
			expected: "line one\nline two",
		},
		{
			name:     "wrapped error where outer contains colon in its own text",
			err:      fmt.Errorf("step 1: validate: %w", errors.New("field missing")),
			expected: "step 1: field missing",
		},
		{
			name:     "empty outer message wrapping an error",
			err:      fmt.Errorf(": %w", errors.New("inner")),
			expected: ": inner",
		},
		{
			name:     "gRPC error with newlines in message",
			err:      status.Error(codes.InvalidArgument, "field A\nfield B"),
			expected: "field A\nfield B",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := buildErrorMessage(test.err)
			if result != test.expected {
				t.Errorf("expected %q, got %q", test.expected, result)
			}
		})
	}
}

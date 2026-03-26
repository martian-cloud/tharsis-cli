package flag

import (
	"fmt"
	"strings"
)

// Marker represents a flag annotation symbol.
type Marker string

// Marker symbols for flag annotations.
const (
	MarkerRequired   Marker = "*"
	MarkerDeprecated Marker = "!"
	MarkerRepeatable Marker = "..."
)

// Color returns the display color name for the marker.
func (m Marker) Color() string {
	switch m {
	case MarkerRequired:
		return "red"
	case MarkerDeprecated:
		return "orange"
	case MarkerRepeatable:
		return "green"
	default:
		return ""
	}
}

// String returns the marker symbol.
func (m Marker) String() string {
	return string(m)
}

// Flag describes a single flag in a Set.
type Flag struct {
	Name  string
	Usage string

	required      bool
	repeatable    bool
	informational bool
	aliases       []string
	predictors    []string
	deprecated    string
	envVar        string
	defaultVal    any
	validValues   []string
	validate      func(string) error
	transform     func(string) string
}

// Markers returns all applicable marker symbols for the flag.
// "*" = required, "!" = deprecated, "..." = repeatable.
func (f *Flag) Markers() []Marker {
	var m []Marker
	if f.required {
		m = append(m, MarkerRequired)
	}

	if f.deprecated != "" {
		m = append(m, MarkerDeprecated)
	}

	if f.repeatable {
		m = append(m, MarkerRepeatable)
	}

	return m
}

// IsRequired reports whether the flag is required.
func (f *Flag) IsRequired() bool {
	return f.required
}

// IsRepeatable reports whether the flag can be specified multiple times.
func (f *Flag) IsRepeatable() bool {
	return f.repeatable
}

// DefValue returns the default value string representation.
func (f *Flag) DefValue() string {
	if f.defaultVal == nil {
		return ""
	}

	return fmt.Sprintf("%v", f.defaultVal)
}

// IsDeprecated reports whether the flag is deprecated.
func (f *Flag) IsDeprecated() bool {
	return f.deprecated != ""
}

// DeprecationMessage returns the deprecation message, or "".
func (f *Flag) DeprecationMessage() string {
	return f.deprecated
}

// FormattedName returns the flag name and aliases joined with ", ".
// e.g. "verbose, v".
func (f *Flag) FormattedName() string {
	parts := append([]string{f.Name}, f.aliases...)
	return strings.Join(parts, ", ")
}

// Aliases returns the flag's alternate names.
func (f *Flag) Aliases() []string {
	return f.aliases
}

// EnvVar returns the environment variable name, or "".
func (f *Flag) EnvVar() string {
	return f.envVar
}

// ValidValues returns the list of valid values, or nil.
func (f *Flag) ValidValues() []string {
	return f.validValues
}

// CompletionValues returns the shell completion candidates.
func (f *Flag) CompletionValues() []string {
	return f.predictors
}

// Validate runs the flag's validator, if any.
// wasSet reports whether this flag or any of its aliases appear in seen.
func (f *Flag) wasSet(seen map[string]bool) bool {
	if seen[f.Name] {
		return true
	}

	for _, a := range f.aliases {
		if seen[a] {
			return true
		}
	}

	return false
}

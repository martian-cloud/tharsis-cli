package flag

import (
	"fmt"
	"sort"
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

	required           bool
	repeatable         bool
	isBool             bool
	aliases            []string
	predictors         []string
	deprecationMessage string
	envVar             string
	defaultVal         any
	validValues        []string
	validate           func(string) error
	transform          func(string) string
	// value holds the current string representation of the flag's parsed value,
	// set by the flag's setter during parsing. This allows callers to read the
	// parsed value via Lookup().Value() without needing access to the flag set.
	value string
}

// Markers returns all applicable marker symbols for the flag.
// "*" = required, "!" = deprecated, "..." = repeatable.
func (f *Flag) Markers() []Marker {
	var m []Marker
	if f.required {
		m = append(m, MarkerRequired)
	}

	if f.deprecationMessage != "" {
		m = append(m, MarkerDeprecated)
	}

	if f.repeatable {
		m = append(m, MarkerRepeatable)
	}

	return m
}

// Predictors returns the completion predictor values for the flag.
func (f *Flag) Predictors() []string {
	return f.predictors
}

// IsBool reports whether the flag is a boolean flag.
func (f *Flag) IsBool() bool {
	return f.isBool
}

// Value returns the current string value of the flag after parsing.
func (f *Flag) Value() string {
	return f.value
}

// DefaultValue returns the default value string representation.
func (f *Flag) DefaultValue() string {
	if f.defaultVal == nil {
		return ""
	}

	return fmt.Sprintf("%v", f.defaultVal)
}

// DeprecationMessage returns the deprecation message, or "".
func (f *Flag) DeprecationMessage() string {
	return f.deprecationMessage
}

// Names returns the flag's primary name and all aliases.
func (f *Flag) Names() []string {
	return append([]string{f.Name}, f.aliases...)
}

// EnvVar returns the environment variable name, or "".
func (f *Flag) EnvVar() string {
	return f.envVar
}

// ValidValues returns the list of valid values sorted alphabetically, or nil.
func (f *Flag) ValidValues() []string {
	if len(f.validValues) == 0 {
		return nil
	}

	sorted := make([]string, len(f.validValues))
	copy(sorted, f.validValues)
	sort.Strings(sorted)

	return sorted
}

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

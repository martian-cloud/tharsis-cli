package flag

import (
	stdflag "flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

// Option configures a flag at registration time.
type Option func(*Flag)

// Required marks a flag as required. [Set.Parse] returns an error if it is
// not set. Panics if combined with [Default].
func Required() Option {
	return func(f *Flag) { f.required = true }
}

// Default sets a fallback value for a flag. Applied at registration time and
// overwritten if the flag is explicitly set. Panics if combined with [Required]
// or if the value type does not match the flag type.
func Default(value any) Option {
	return func(f *Flag) { f.defaultVal = value }
}

// Deprecated marks a flag as deprecated with a message shown when it is used.
// A deprecated flag cannot be required since it has a replacement; panics if
// combined with [Required].
func Deprecated(message string) Option {
	return func(f *Flag) { f.deprecationMessage = message }
}

// TransformString applies a function to the raw string value before storing it.
// Only applies to string and string slice flags.
func TransformString(fn func(string) string) Option {
	return func(f *Flag) { f.transform = fn }
}

// PredictValues provides shell completion candidates for a flag.
func PredictValues(values ...string) Option {
	return func(f *Flag) { f.predictors = values }
}

// Aliases registers short or alternate names for a flag (e.g. Aliases("n")
// lets -n work as an alias for -name).
func Aliases(names ...string) Option {
	return func(f *Flag) { f.aliases = names }
}

// EnvVar sets an environment variable that provides a fallback value for the
// flag. The env var is read at registration time; an explicit flag value
// always wins.
func EnvVar(key string) Option {
	return func(f *Flag) { f.envVar = key }
}

// ValidValues restricts a flag to one of the given values.
func ValidValues(values ...string) Option {
	return func(f *Flag) {
		f.validValues = values
		f.validate = func(s string) error {
			if slices.Contains(values, s) {
				return nil
			}

			return fmt.Errorf("invalid value %q for flag %s, must be one of: %s",
				s, f.Name, strings.Join(values, ", "))
		}
	}
}

// ValidRange restricts a numeric flag to the given inclusive range.
func ValidRange(minVal, maxVal int) Option {
	return func(f *Flag) {
		f.validate = func(s string) error {
			v, err := strconv.ParseInt(s, 0, 64)
			if err != nil {
				return err
			}

			if int(v) < minVal || int(v) > maxVal {
				return fmt.Errorf("value %d for flag %s must be between %d and %d",
					v, f.Name, minVal, maxVal)
			}

			return nil
		}
	}
}

// Validate sets a custom validation function for the flag value.
func Validate(fn func(string) error) Option {
	return func(f *Flag) { f.validate = fn }
}

// ---------------------------------------------------------------------------
// Set
// ---------------------------------------------------------------------------

// Set wraps [stdflag.Set] with required flags, defaults, validation,
// and shell completion support.
type Set struct {
	stdfs *stdflag.FlagSet
	flags map[string]*Flag
	// deprecations holds warnings for deprecated flags used during parsing.
	// Populated by Parse; read via Deprecations.
	deprecations []string
	// mutuallyExclusive holds groups of flag names where at most one may be set.
	mutuallyExclusive [][]string
}

// NewSet creates a new Set. Error handling is set to
// [stdflag.ContinueOnError] so callers can inspect parse errors.
func NewSet(name string) *Set {
	return &Set{
		stdfs: stdflag.NewFlagSet(name, stdflag.ContinueOnError),
		flags: make(map[string]*Flag),
	}
}

// Name returns the name of the flag set.
func (fs *Set) Name() string {
	return fs.stdfs.Name()
}

// MutuallyExclusive declares a group of flags where at most one may be set.
// Parse returns an error if more than one flag in the group is provided.
func (fs *Set) MutuallyExclusive(names ...string) {
	fs.mutuallyExclusive = append(fs.mutuallyExclusive, names)

	for _, name := range names {
		if f, ok := fs.flags[name]; ok {
			for _, other := range names {
				if other != name {
					f.exclusiveWith = append(f.exclusiveWith, other)
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Delegated methods
// ---------------------------------------------------------------------------

// SetOutput sets the output writer for error messages and usage.
func (fs *Set) SetOutput(output io.Writer) {
	fs.stdfs.SetOutput(output)
}

// Args returns the non-flag arguments remaining after parsing.
func (fs *Set) Args() []string {
	return fs.stdfs.Args()
}

// Deprecations returns warning messages for any deprecated flags that were
// used during the last call to [Set.Parse]. Returns nil if none were used.
// Callers should display these in non-JSON output modes.
func (fs *Set) Deprecations() []string {
	return fs.deprecations
}

// ---------------------------------------------------------------------------
// Query methods
// ---------------------------------------------------------------------------

// Lookup returns the Flag for the named flag, or nil if not found.
func (fs *Set) Lookup(name string) *Flag {
	return fs.flags[name]
}

// VisitAll calls fn for each flag in the set (sorted alphabetically),
// including informational and unset flags. Aliases are excluded.
func (fs *Set) VisitAll(fn func(*Flag)) {
	names := make([]string, 0, len(fs.flags))
	for name, f := range fs.flags {
		if name == f.Name {
			names = append(names, name)
		}
	}

	slices.Sort(names)

	for _, name := range names {
		fn(fs.flags[name])
	}
}

// Informational registers a flag for help and documentation display only.
// It is not parsed from the command line.
func (fs *Set) Informational(name string, usage string, opts ...any) {
	fs.register(name, usage, opts)
}

// ---------------------------------------------------------------------------
// Scalar flag registration (**T pointers, nullable)
// ---------------------------------------------------------------------------

// StringVar defines a string flag. Optional by default (nil if not set).
func (fs *Set) StringVar(p **string, name string, usage string, args ...any) {
	f := fs.register(name, usage, args)

	setter := func(s string) error {
		s = f.transform(s)

		if err := f.validate(s); err != nil {
			return err
		}

		*p = &s
		f.value = s

		return nil
	}

	fs.stdfs.Func(name, f.Usage, setter)
	fs.registerAliases(f, setter)

	setDefault(fs, f, p)
	setEnvDefault(f, p)
}

// IntVar defines an int flag. Optional by default (nil if not set).
func (fs *Set) IntVar(p **int, name string, usage string, args ...any) {
	f := fs.register(name, usage, args)

	setter := func(s string) error {
		if err := f.validate(s); err != nil {
			return err
		}

		v, err := strconv.ParseInt(s, 0, 0)
		if err != nil {
			return err
		}

		intVal := int(v)
		*p = &intVal
		f.value = s

		return nil
	}

	fs.stdfs.Func(name, f.Usage, setter)
	fs.registerAliases(f, setter)

	setDefault(fs, f, p)
	setEnvDefault(f, p)
}

// Int32Var defines an int32 flag. Optional by default (nil if not set).
func (fs *Set) Int32Var(p **int32, name string, usage string, args ...any) {
	f := fs.register(name, usage, args)

	setter := func(s string) error {
		if err := f.validate(s); err != nil {
			return err
		}

		v, err := strconv.ParseInt(s, 0, 32)
		if err != nil {
			return err
		}

		v32 := int32(v)
		*p = &v32
		f.value = s

		return nil
	}

	fs.stdfs.Func(name, f.Usage, setter)
	fs.registerAliases(f, setter)

	setDefault(fs, f, p)
	setEnvDefault(f, p)
}

// Int64Var defines an int64 flag. Optional by default (nil if not set).
func (fs *Set) Int64Var(p **int64, name string, usage string, args ...any) {
	f := fs.register(name, usage, args)

	setter := func(s string) error {
		if err := f.validate(s); err != nil {
			return err
		}

		v, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			return err
		}

		*p = &v
		f.value = s

		return nil
	}

	fs.stdfs.Func(name, f.Usage, setter)
	fs.registerAliases(f, setter)

	setDefault(fs, f, p)
	setEnvDefault(f, p)
}

// BoolVar defines a bool flag. Optional by default (nil if not set).
func (fs *Set) BoolVar(p **bool, name string, usage string, args ...any) {
	f := fs.register(name, usage, args)
	f.isBool = true

	setter := func(s string) error {
		v, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}

		*p = &v
		f.value = s

		return nil
	}

	fs.stdfs.BoolFunc(name, f.Usage, setter)
	for _, alias := range f.aliases {
		fs.stdfs.BoolFunc(alias, f.Usage, setter)
		fs.flags[alias] = f
	}

	setDefault(fs, f, p)
	setEnvDefault(f, p)
}

// ---------------------------------------------------------------------------
// Repeatable flag registration
// ---------------------------------------------------------------------------

// StringSliceVar defines a repeatable string flag that appends to a slice.
func (fs *Set) StringSliceVar(p *[]string, name string, usage string, args ...any) {
	f := fs.register(name, usage, args)
	f.repeatable = true

	setter := func(s string) error {
		s = f.transform(s)

		if err := f.validate(s); err != nil {
			return err
		}

		*p = append(*p, s)

		return nil
	}

	fs.stdfs.Func(name, f.Usage, setter)
	fs.registerAliases(f, setter)
}

// MapVar defines a repeatable flag that parses key=value pairs into a map.
// Use key=- to remove a key.
func (fs *Set) MapVar(p *map[string]string, name string, usage string, args ...any) {
	f := fs.register(name, usage, args)
	f.repeatable = true

	if *p == nil {
		*p = make(map[string]string)
	}

	setter := func(s string) error {
		if err := f.validate(s); err != nil {
			return err
		}

		parts := strings.SplitN(s, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format: %s (expected key=value)", s)
		}

		if parts[0] == "" {
			return fmt.Errorf("key cannot be empty")
		}

		// key=- removes the key from the map.
		if parts[1] == "-" {
			delete(*p, parts[0])
		} else {
			(*p)[parts[0]] = parts[1]
		}

		return nil
	}

	fs.stdfs.Func(name, f.Usage, setter)
	fs.registerAliases(f, setter)
}

// ---------------------------------------------------------------------------
// Parsing
// ---------------------------------------------------------------------------

// Parse parses the command-line flags and checks that all required flags were set.
// Both -flag and --flag styles are supported.
// After a successful parse, call [Set.Deprecations] to check for deprecated flag usage.
func (fs *Set) Parse(arguments []string) error {
	arguments = fs.rewriteBoolArgs(arguments)

	if err := fs.stdfs.Parse(arguments); err != nil {
		return err
	}

	seen := make(map[string]bool)
	fs.stdfs.Visit(func(sf *stdflag.Flag) { seen[sf.Name] = true })

	// Collect deprecation warnings for used deprecated flags.
	for name, f := range fs.flags {
		if name != f.Name {
			continue
		}

		// Surface deferred parse errors (e.g. invalid env var) for flags
		// that were not explicitly set on the command line.
		if f.parseErr != nil && !f.wasSet(seen) {
			return f.parseErr
		}

		if f.deprecationMessage != "" && f.wasSet(seen) {
			fs.deprecations = append(fs.deprecations, fmt.Sprintf("flag -%s is deprecated: %s", f.Name, f.deprecationMessage))
		}
	}

	slices.Sort(fs.deprecations)

	// Check mutually exclusive groups.
	for _, group := range fs.mutuallyExclusive {
		var set []string
		for _, name := range group {
			if f, ok := fs.flags[name]; ok && f.wasSet(seen) {
				set = append(set, "-"+name)
			}
		}

		if len(set) > 1 {
			return fmt.Errorf("only one of %s may be set", strings.Join(set, ", "))
		}
	}

	var missing []string
	for name, f := range fs.flags {
		if name == f.Name && f.required && !f.wasSet(seen) {
			missing = append(missing, name)
		}
	}

	slices.Sort(missing)

	if len(missing) == 1 {
		return fmt.Errorf("flag -%s is required", missing[0])
	}

	if len(missing) > 1 {
		return fmt.Errorf("flags -%s are required", strings.Join(missing, ", -"))
	}

	return nil
}

// rewriteBoolArgs rewrites "-flag true/false" to "-flag=true/false" for bool
// flags so that the value is consumed as part of the flag rather than left as
// a positional argument. This preserves backwards compatibility with the old
// optparser which treated bool flags as string-valued.
func (fs *Set) rewriteBoolArgs(args []string) []string {
	for i := 0; i < len(args)-1; i++ {
		arg := args[i]
		name := strings.TrimLeft(arg, "-")
		if name == "" || strings.Contains(arg, "=") {
			continue
		}

		f, ok := fs.flags[name]
		if !ok || !f.isBool {
			continue
		}

		next := args[i+1]
		if _, err := strconv.ParseBool(next); err != nil {
			continue
		}

		args = append(args[:i+1], args[i+2:]...)
		args[i] = arg + "=" + next
		fs.deprecations = append(fs.deprecations, fmt.Sprintf("use %s=%s instead of %s %s", arg, next, arg, next))
	}

	return args
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// register creates a Flag, applies options, stores it, and returns it.
func (fs *Set) register(name string, usage string, args []any) *Flag {
	opts, formatArgs := extractOptions(args)

	if len(formatArgs) > 0 {
		usage = fmt.Sprintf(usage, formatArgs...)
	}

	f := &Flag{
		Name:      name,
		Usage:     usage,
		transform: func(s string) string { return s },
		validate:  func(string) error { return nil },
	}
	for _, opt := range opts {
		opt(f)
	}

	// Specifying a default means the flag isn't required or vice-versa.
	if f.required && f.defaultVal != nil {
		panic(fmt.Sprintf("flag %q: cannot be both required and have a default", name))
	}

	// Deprecated flags should always be optional since they're being replaced by another.
	if f.required && f.deprecationMessage != "" {
		panic(fmt.Sprintf("flag %q: cannot be both required and deprecated", name))
	}

	fs.flags[name] = f

	return f
}

// registerAliases registers alternate names for a flag that delegate to the
// same setter function.
func (fs *Set) registerAliases(f *Flag, setter func(string) error) {
	for _, alias := range f.aliases {
		fs.stdfs.Func(alias, f.Usage, setter)
		fs.flags[alias] = f
	}
}

// extractOptions separates FlagOptions from format args in a variadic slice.
// Format args must come before FlagOptions. Panics if a non-FlagOption value
// appears after a FlagOption.
func extractOptions(args []any) ([]Option, []any) {
	var opts []Option
	var formatArgs []any

	foundOption := false

	for _, arg := range args {
		if opt, ok := arg.(Option); ok {
			opts = append(opts, opt)
			foundOption = true
		} else if foundOption {
			panic(fmt.Sprintf("flags: format argument %v must come before FlagOptions", arg))
		} else {
			formatArgs = append(formatArgs, arg)
		}
	}

	return opts, formatArgs
}

// setDefault applies the default value for the named flag to the pointer.
// Called by each XxxVar after registration. Panics on type mismatch.
func setDefault[T any](fs *Set, f *Flag, p **T) {
	if f.defaultVal == nil {
		return
	}

	v, ok := f.defaultVal.(T)
	if !ok {
		panic(fmt.Sprintf("flag %q: default value has type %T, expected %T", f.Name, f.defaultVal, *new(T)))
	}

	*p = &v

	if sf := fs.stdfs.Lookup(f.Name); sf != nil {
		sf.DefValue = fmt.Sprintf("%v", v)
	}
}

// setEnvDefault reads the flag's env var (if configured) and applies it as a
// fallback. It runs after setDefault so that env vars take precedence over
// coded defaults but explicit flags still win (handled at parse time).
// Stores an error on the flag if the env var value fails validation or parsing.
func setEnvDefault[T any](f *Flag, p **T) {
	if f.envVar == "" {
		return
	}

	v, ok := os.LookupEnv(f.envVar)
	if !ok || v == "" {
		return
	}

	if err := f.validate(v); err != nil {
		f.parseErr = fmt.Errorf("environment variable %s: %w", f.envVar, err)
		return
	}

	// Parse the env value into the target type.
	var result any
	var err error

	switch any(*new(T)).(type) {
	case string:
		v = f.transform(v)
		result = v
	case int:
		var n int64
		n, err = strconv.ParseInt(v, 0, 0)
		result = int(n)
	case int32:
		var n int64
		n, err = strconv.ParseInt(v, 0, 32)
		result = int32(n)
	case int64:
		result, err = strconv.ParseInt(v, 0, 64)
	case bool:
		result, err = strconv.ParseBool(v)
	}

	if err != nil {
		f.parseErr = fmt.Errorf("environment variable %s: %w", f.envVar, err)
		return
	}

	typed, ok := result.(T)
	if !ok {
		return
	}

	*p = &typed
}

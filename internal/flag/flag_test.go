package flag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlagIsDeprecated(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		expectBool bool
	}{
		{name: "not deprecated", message: "", expectBool: false},
		{name: "deprecated", message: "use --new instead", expectBool: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Flag{deprecated: tc.message}
			assert.Equal(t, tc.expectBool, f.IsDeprecated())
			assert.Equal(t, tc.message, f.DeprecationMessage())
		})
	}
}

func TestFlagAliases(t *testing.T) {
	tests := []struct {
		name    string
		aliases []string
	}{
		{name: "nil", aliases: nil},
		{name: "single", aliases: []string{"n"}},
		{name: "multiple", aliases: []string{"n", "nm"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Flag{aliases: tc.aliases}
			assert.Equal(t, tc.aliases, f.Aliases())
		})
	}
}

func TestFlagEnvVar(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
	}{
		{name: "empty", envVar: ""},
		{name: "set", envVar: "MY_TOKEN"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &Flag{envVar: tc.envVar}
			assert.Equal(t, tc.envVar, f.EnvVar())
		})
	}
}

func TestFlagValidate(t *testing.T) {
	tests := []struct {
		name      string
		opts      []any
		input     string
		expectErr bool
	}{
		{name: "no validator", input: "anything"},
		{name: "passes", opts: []any{ValidValues("ok")}, input: "ok"},
		{name: "fails", opts: []any{ValidValues("ok")}, input: "bad", expectErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var val *string
			fs := NewSet("test")
			fs.StringVar(&val, "flag", "usage", tc.opts...)
			err := fs.Parse([]string{"-flag", tc.input})
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFlagWasSet(t *testing.T) {
	tests := []struct {
		name   string
		flag   *Flag
		seen   map[string]bool
		expect bool
	}{
		{
			name:   "primary name seen",
			flag:   &Flag{Name: "name"},
			seen:   map[string]bool{"name": true},
			expect: true,
		},
		{
			name:   "alias seen",
			flag:   &Flag{Name: "name", aliases: []string{"n"}},
			seen:   map[string]bool{"n": true},
			expect: true,
		},
		{
			name:   "neither seen",
			flag:   &Flag{Name: "name", aliases: []string{"n"}},
			seen:   map[string]bool{"other": true},
			expect: false,
		},
		{
			name:   "empty seen",
			flag:   &Flag{Name: "name"},
			seen:   map[string]bool{},
			expect: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, tc.flag.wasSet(tc.seen))
		})
	}
}

func TestLookup(t *testing.T) {
	fs := NewSet("test")
	var s *string
	fs.StringVar(&s, "format", "output format",
		Default("json"),
		Deprecated("use --output instead"),
		PredictValues("json", "table"),
	)

	t.Run("found with metadata", func(t *testing.T) {
		f := fs.Lookup("format")
		require.NotNil(t, f)
		assert.Equal(t, "format", f.Name)
		assert.Equal(t, "output format", f.Usage)
		assert.Equal(t, "json", f.DefValue())
		assert.True(t, f.IsDeprecated())
		assert.Equal(t, "use --output instead", f.DeprecationMessage())
	})

	t.Run("required flag", func(t *testing.T) {
		var name *string
		fs.StringVar(&name, "name", "a name", Required())

		f := fs.Lookup("name")
		require.NotNil(t, f)
		assert.Contains(t, f.Markers(), Marker("*"))
	})

	t.Run("flag with aliases", func(t *testing.T) {
		var verbose *bool
		fs.BoolVar(&verbose, "verbose", "verbose output", Aliases("v"))

		f := fs.Lookup("verbose")
		require.NotNil(t, f)
		assert.Equal(t, []string{"v"}, f.Aliases())
	})

	t.Run("flag with env var", func(t *testing.T) {
		var token *string
		fs.StringVar(&token, "token", "auth token", EnvVar("MY_TOKEN"))

		f := fs.Lookup("token")
		require.NotNil(t, f)
		assert.Equal(t, "MY_TOKEN", f.EnvVar())
	})

	t.Run("not found", func(t *testing.T) {
		assert.Nil(t, fs.Lookup("missing"))
	})
}

func TestVisitAll(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*Set)
		expectedNames []string
	}{
		{
			name:          "empty set",
			setup:         func(_ *Set) {},
			expectedNames: nil,
		},
		{
			name: "sorted alphabetically",
			setup: func(fs *Set) {
				var a, b *string
				fs.StringVar(&a, "zebra", "last")
				fs.StringVar(&b, "alpha", "first")
			},
			expectedNames: []string{"alpha", "zebra"},
		},
		{
			name: "aliases excluded",
			setup: func(fs *Set) {
				var v *string
				fs.StringVar(&v, "verbose", "verbose output", Aliases("v"))
			},
			expectedNames: []string{"verbose"},
		},
		{
			name: "informational flags included",
			setup: func(fs *Set) {
				var v *bool
				fs.BoolVar(&v, "debug", "enable debug", Default(false))
				fs.Informational("help", "show help")
			},
			expectedNames: []string{"debug", "help"},
		},
		{
			name: "mixed parsed and informational sorted together",
			setup: func(fs *Set) {
				var v *string
				fs.StringVar(&v, "name", "the name")
				fs.Informational("version", "show version")
				fs.Informational("help", "show help")
			},
			expectedNames: []string{"help", "name", "version"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewSet("test")
			tc.setup(fs)

			var names []string
			fs.VisitAll(func(f *Flag) {
				names = append(names, f.Name)
			})

			assert.Equal(t, tc.expectedNames, names)
		})
	}
}

func TestPredictors(t *testing.T) {
	t.Run("with predict values", func(t *testing.T) {
		fs := NewSet("test")
		var s *string
		fs.StringVar(&s, "format", "output format",
			PredictValues("json", "table"), Aliases("f"),
		)

		var found *Flag
		fs.VisitAll(func(f *Flag) {
			if f.Name == "format" {
				found = f
			}
		})

		require.NotNil(t, found)
		assert.Equal(t, []string{"json", "table"}, found.Predictors())
	})

	t.Run("no predictors", func(t *testing.T) {
		fs := NewSet("test")
		var s *string
		fs.StringVar(&s, "plain", "no predictions")

		var found *Flag
		fs.VisitAll(func(f *Flag) {
			if f.Name == "plain" {
				found = f
			}
		})

		require.NotNil(t, found)
		assert.Nil(t, found.Predictors())
	})
}

func TestName(t *testing.T) {
	fs := NewSet("Global options")
	assert.Equal(t, "Global options", fs.Name())
}

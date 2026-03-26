package flag

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringVar(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		opts      []any
		expectNil bool
		expectVal string
		expectErr string
	}{
		{
			name:      "required and set",
			args:      []string{"-name", "value"},
			opts:      []any{Required()},
			expectVal: "value",
		},
		{
			name:      "required and missing",
			args:      []string{},
			opts:      []any{Required()},
			expectErr: "flag name is required",
		},
		{
			name:      "optional not set",
			args:      []string{},
			expectNil: true,
		},
		{
			name:      "optional set",
			args:      []string{"-name", "hello"},
			expectVal: "hello",
		},
		{
			name:      "default not set",
			args:      []string{},
			opts:      []any{Default("json")},
			expectVal: "json",
		},
		{
			name:      "default overridden",
			args:      []string{"-name", "yaml"},
			opts:      []any{Default("json")},
			expectVal: "yaml",
		},
		{
			name:      "duplicate last wins",
			args:      []string{"-name", "foo", "-name", "bar"},
			expectVal: "bar",
		},
		{
			name:      "alias",
			args:      []string{"-n", "short"},
			opts:      []any{Aliases("n")},
			expectVal: "short",
		},
		{
			name:      "required satisfied by alias",
			args:      []string{"-n", "ok"},
			opts:      []any{Required(), Aliases("n")},
			expectVal: "ok",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewSet("test")
			fs.SetOutput(io.Discard)
			var val *string
			fs.StringVar(&val, "name", "name flag", tc.opts...)

			err := fs.Parse(tc.args)

			if tc.expectErr != "" {
				assert.ErrorContains(t, err, tc.expectErr)
				return
			}

			require.NoError(t, err)

			if tc.expectNil {
				assert.Nil(t, val)
				return
			}

			require.NotNil(t, val)
			assert.Equal(t, tc.expectVal, *val)
		})
	}
}

func TestIntVar(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		opts      []any
		expectNil bool
		expectVal int
		expectErr string
	}{
		{
			name:      "optional not set",
			args:      []string{},
			expectNil: true,
		},
		{
			name:      "set",
			args:      []string{"-n", "42"},
			expectVal: 42,
		},
		{
			name:      "default not set",
			args:      []string{},
			opts:      []any{Default(100)},
			expectVal: 100,
		},
		{
			name:      "hex",
			args:      []string{"-n", "0xFF"},
			expectVal: 255,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewSet("test")
			fs.SetOutput(io.Discard)
			var val *int
			fs.IntVar(&val, "n", "number", tc.opts...)

			err := fs.Parse(tc.args)

			if tc.expectErr != "" {
				assert.ErrorContains(t, err, tc.expectErr)
				return
			}

			require.NoError(t, err)

			if tc.expectNil {
				assert.Nil(t, val)
				return
			}

			require.NotNil(t, val)
			assert.Equal(t, tc.expectVal, *val)
		})
	}
}

func TestInt32Var(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		opts      []any
		expectNil bool
		expectVal int32
	}{
		{
			name:      "optional not set",
			args:      []string{},
			expectNil: true,
		},
		{
			name:      "set",
			args:      []string{"-n", "3000"},
			expectVal: 3000,
		},
		{
			name:      "default not set",
			args:      []string{},
			opts:      []any{Default(int32(8080))},
			expectVal: 8080,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewSet("test")
			var val *int32
			fs.Int32Var(&val, "n", "number", tc.opts...)

			err := fs.Parse(tc.args)
			require.NoError(t, err)

			if tc.expectNil {
				assert.Nil(t, val)
				return
			}

			require.NotNil(t, val)
			assert.Equal(t, tc.expectVal, *val)
		})
	}
}

func TestInt64Var(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectNil bool
		expectVal int64
	}{
		{
			name:      "optional not set",
			args:      []string{},
			expectNil: true,
		},
		{
			name:      "set",
			args:      []string{"-n", "9223372036854775807"},
			expectVal: 9223372036854775807,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewSet("test")
			var val *int64
			fs.Int64Var(&val, "n", "number")

			err := fs.Parse(tc.args)
			require.NoError(t, err)

			if tc.expectNil {
				assert.Nil(t, val)
				return
			}

			require.NotNil(t, val)
			assert.Equal(t, tc.expectVal, *val)
		})
	}
}

func TestBoolVar(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		opts      []any
		expectNil bool
		expectVal bool
	}{
		{
			name:      "optional not set",
			args:      []string{},
			expectNil: true,
		},
		{
			name:      "set true",
			args:      []string{"-b", "true"},
			expectVal: true,
		},
		{
			name:      "bare flag implies true",
			args:      []string{"-b"},
			expectVal: true,
		},
		{
			name:      "default false not set",
			args:      []string{},
			opts:      []any{Default(false)},
			expectVal: false,
		},
		{
			name:      "default false overridden",
			args:      []string{"-b", "true"},
			opts:      []any{Default(false)},
			expectVal: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewSet("test")
			var val *bool
			fs.BoolVar(&val, "b", "bool flag", tc.opts...)

			err := fs.Parse(tc.args)
			require.NoError(t, err)

			if tc.expectNil {
				assert.Nil(t, val)
				return
			}

			require.NotNil(t, val)
			assert.Equal(t, tc.expectVal, *val)
		})
	}
}

func TestEnvVarFallback(t *testing.T) {
	t.Run("env var used when flag not set", func(t *testing.T) {
		t.Setenv("TEST_NAME", "from-env")

		fs := NewSet("test")
		var val *string
		fs.StringVar(&val, "name", "name flag", EnvVar("TEST_NAME"))

		err := fs.Parse([]string{})
		require.NoError(t, err)
		require.NotNil(t, val)
		assert.Equal(t, "from-env", *val)
	})

	t.Run("explicit flag overrides env var", func(t *testing.T) {
		t.Setenv("TEST_NAME", "from-env")

		fs := NewSet("test")
		var val *string
		fs.StringVar(&val, "name", "name flag", EnvVar("TEST_NAME"))

		err := fs.Parse([]string{"-name", "from-flag"})
		require.NoError(t, err)
		require.NotNil(t, val)
		assert.Equal(t, "from-flag", *val)
	})

	t.Run("env var not set leaves nil", func(t *testing.T) {
		fs := NewSet("test")
		var val *string
		fs.StringVar(&val, "name", "name flag", EnvVar("TEST_UNSET_VAR"))

		err := fs.Parse([]string{})
		require.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("default takes precedence over missing env var", func(t *testing.T) {
		fs := NewSet("test")
		var val *string
		fs.StringVar(&val, "name", "name flag", Default("fallback"), EnvVar("TEST_UNSET_VAR"))

		err := fs.Parse([]string{})
		require.NoError(t, err)
		require.NotNil(t, val)
		assert.Equal(t, "fallback", *val)
	})

	t.Run("env var overrides default", func(t *testing.T) {
		t.Setenv("TEST_NAME", "from-env")

		fs := NewSet("test")
		var val *string
		fs.StringVar(&val, "name", "name flag", Default("fallback"), EnvVar("TEST_NAME"))

		err := fs.Parse([]string{})
		require.NoError(t, err)
		require.NotNil(t, val)
		assert.Equal(t, "from-env", *val)
	})
}

func TestDeprecated(t *testing.T) {
	fs := NewSet("test")
	var old *string
	fs.StringVar(&old, "old", "old flag", Deprecated("use --new instead"))

	var found *Flag
	fs.VisitAll(func(f *Flag) {
		if f.Name == "old" {
			found = f
		}
	})

	require.NotNil(t, found)
	assert.Equal(t, "use --new instead", found.DeprecationMessage())
}

func TestMultipleRequiredMissing(t *testing.T) {
	fs := NewSet("test")
	var name, id *string
	fs.StringVar(&name, "name", "name flag", Required())
	fs.StringVar(&id, "id", "id flag", Required())

	err := fs.Parse([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "flags")
	assert.Contains(t, err.Error(), "are required")
}

func TestStringSliceVar(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectVal []string
	}{
		{
			name:      "multiple values",
			args:      []string{"-tag", "foo", "-tag", "bar", "-tag", "baz"},
			expectVal: []string{"foo", "bar", "baz"},
		},
		{
			name:      "empty",
			args:      []string{},
			expectVal: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewSet("test")
			var tags []string
			fs.StringSliceVar(&tags, "tag", "tag (repeatable)")

			err := fs.Parse(tc.args)
			require.NoError(t, err)
			assert.Equal(t, tc.expectVal, tags)
		})
	}
}

func TestMapVar(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectVal map[string]string
		expectErr string
	}{
		{
			name:      "multiple pairs",
			args:      []string{"-label", "env=prod", "-label", "tier=frontend"},
			expectVal: map[string]string{"env": "prod", "tier": "frontend"},
		},
		{
			name:      "invalid format",
			args:      []string{"-label", "invalid"},
			expectErr: "invalid format",
		},
		{
			name:      "empty key",
			args:      []string{"-label", "=value"},
			expectErr: "key cannot be empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewSet("test")
			fs.SetOutput(io.Discard)
			var labels map[string]string
			fs.MapVar(&labels, "label", "label key=value pair")

			err := fs.Parse(tc.args)

			if tc.expectErr != "" {
				assert.ErrorContains(t, err, tc.expectErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectVal, labels)
		})
	}
}

func TestFormatArgs(t *testing.T) {
	tests := []struct {
		name        string
		usage       string
		args        []any
		expectUsage string
	}{
		{
			name:        "format args only",
			usage:       "name for %s in %s",
			args:        []any{"resource", "group", Required()},
			expectUsage: "name for resource in group",
		},
		{
			name:        "format args with options",
			usage:       "output format for %s",
			args:        []any{"results", PredictValues("json", "yaml")},
			expectUsage: "output format for results",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewSet("test")
			var val *string
			fs.StringVar(&val, "name", tc.usage, tc.args...)

			var usage string
			fs.VisitAll(func(f *Flag) {
				if f.Name == "name" {
					usage = f.Usage
				}
			})

			assert.Equal(t, tc.expectUsage, usage)
		})
	}
}

func TestValidValues(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectVal string
		expectErr string
	}{
		{
			name:      "valid value",
			args:      []string{"-format", "json"},
			expectVal: "json",
		},
		{
			name:      "invalid value",
			args:      []string{"-format", "invalid"},
			expectErr: "invalid value \"invalid\" for flag format",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewSet("test")
			fs.SetOutput(io.Discard)
			var format *string
			fs.StringVar(&format, "format", "output format", ValidValues("json", "yaml", "table"), Required())

			err := fs.Parse(tc.args)

			if tc.expectErr != "" {
				assert.ErrorContains(t, err, tc.expectErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, format)
			assert.Equal(t, tc.expectVal, *format)
		})
	}
}

func TestValidRange(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectVal int
		expectErr string
	}{
		{
			name:      "in range",
			args:      []string{"-count", "5"},
			expectVal: 5,
		},
		{
			name:      "below range",
			args:      []string{"-count", "0"},
			expectErr: "value 0 for flag count must be between 1 and 10",
		},
		{
			name:      "above range",
			args:      []string{"-count", "11"},
			expectErr: "value 11 for flag count must be between 1 and 10",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewSet("test")
			fs.SetOutput(io.Discard)
			var count *int
			fs.IntVar(&count, "count", "count value", ValidRange(1, 10), Required())

			err := fs.Parse(tc.args)

			if tc.expectErr != "" {
				assert.ErrorContains(t, err, tc.expectErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, count)
			assert.Equal(t, tc.expectVal, *count)
		})
	}
}

func TestValidValuesOnStringSlice(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectVal []string
		expectErr string
	}{
		{
			name:      "valid values",
			args:      []string{"-env", "dev", "-env", "prod"},
			expectVal: []string{"dev", "prod"},
		},
		{
			name:      "invalid value",
			args:      []string{"-env", "invalid"},
			expectErr: "invalid value \"invalid\" for flag env",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewSet("test")
			fs.SetOutput(io.Discard)
			var envs []string
			fs.StringSliceVar(&envs, "env", "environment", ValidValues("dev", "staging", "prod"))

			err := fs.Parse(tc.args)

			if tc.expectErr != "" {
				assert.ErrorContains(t, err, tc.expectErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectVal, envs)
		})
	}
}

func TestPanics(t *testing.T) {
	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "misplaced format args",
			fn: func() {
				extractOptions([]any{"format", PredictValues("json"), "oops"})
			},
		},
		{
			name: "default wrong type",
			fn: func() {
				fs := NewSet("test")
				var name *string
				fs.StringVar(&name, "name", "name flag", Default(123))
			},
		},
		{
			name: "required with default",
			fn: func() {
				fs := NewSet("test")
				var name *string
				fs.StringVar(&name, "name", "name flag", Required(), Default("x"))
			},
		},
		{
			name: "required with deprecated",
			fn: func() {
				fs := NewSet("test")
				var name *string
				fs.StringVar(&name, "name", "name flag", Required(), Deprecated("use --new"))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Panics(t, tc.fn)
		})
	}
}

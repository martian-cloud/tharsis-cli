package command

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestBuildTerraformBackendConfig(t *testing.T) {
	tests := []struct {
		name          string
		tharsisURL    string
		workspacePath string
		wantHostname  string
		wantOrg       string
		wantWorkspace string
		wantErr       bool
	}{
		{
			name:          "standard four-segment path",
			tharsisURL:    "https://tharsis.cts.infor.com",
			workspacePath: "cts/tharsis/sandbox/demo-project",
			wantHostname:  "tharsis.cts.infor.com",
			wantOrg:       "cts.tharsis.sandbox",
			wantWorkspace: "demo-project",
		},
		{
			name:          "two-segment path (single group)",
			tharsisURL:    "https://tharsis.example.com",
			workspacePath: "mygroup/myworkspace",
			wantHostname:  "tharsis.example.com",
			wantOrg:       "mygroup",
			wantWorkspace: "myworkspace",
		},
		{
			name:          "leading and trailing slashes are trimmed",
			tharsisURL:    "https://tharsis.example.com",
			workspacePath: "/a/b/c/",
			wantHostname:  "tharsis.example.com",
			wantOrg:       "a.b",
			wantWorkspace: "c",
		},
		{
			name:          "hostname with port strips port",
			tharsisURL:    "https://tharsis.example.com:8080",
			workspacePath: "g/ws",
			wantHostname:  "tharsis.example.com",
			wantOrg:       "g",
			wantWorkspace: "ws",
		},
		{
			name:          "invalid URL",
			tharsisURL:    "://bad",
			workspacePath: "g/ws",
			wantErr:       true,
		},
		{
			name:          "empty hostname",
			tharsisURL:    "https://",
			workspacePath: "g/ws",
			wantErr:       true,
		},
		{
			name:          "single segment workspace path",
			tharsisURL:    "https://tharsis.example.com",
			workspacePath: "onlyone",
			wantErr:       true,
		},
		{
			name:          "empty workspace path",
			tharsisURL:    "https://tharsis.example.com",
			workspacePath: "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildTerraformBackendConfig(tt.tharsisURL, tt.workspacePath)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Contains(t, got, `backend "remote"`)
			assert.Contains(t, got, tt.wantHostname)
			assert.Contains(t, got, tt.wantOrg)
			assert.Contains(t, got, tt.wantWorkspace)
		})
	}
}

func TestTerraformFlags(t *testing.T) {
	// Verify that Flags() correctly parses --workspace, --tf-path, and --work-dir,
	// and that any args following the first non-flag token are left in Args() for
	// passthrough to terraform.
	tests := []struct {
		name              string
		args              []string
		wantTFPath        string
		wantWorkspacePath string
		wantWorkDir       string
		wantRemaining     []string
	}{
		{
			name:              "--workspace equals form",
			args:              []string{"--workspace=my/group/ws"},
			wantWorkspacePath: "my/group/ws",
			wantRemaining:     []string{},
		},
		{
			name:              "--workspace space form",
			args:              []string{"--workspace", "my/group/ws"},
			wantWorkspacePath: "my/group/ws",
			wantRemaining:     []string{},
		},
		{
			name:          "--tf-path equals form",
			args:          []string{"--tf-path=/usr/bin/terraform"},
			wantTFPath:    "/usr/bin/terraform",
			wantRemaining: []string{},
		},
		{
			name:          "--tf-path space form",
			args:          []string{"--tf-path", "/usr/bin/terraform"},
			wantTFPath:    "/usr/bin/terraform",
			wantRemaining: []string{},
		},
		{
			name:          "--work-dir equals form",
			args:          []string{"--work-dir=/my/dir"},
			wantWorkDir:   "/my/dir",
			wantRemaining: []string{},
		},
		{
			name:          "--work-dir space form",
			args:          []string{"--work-dir", "/my/dir"},
			wantWorkDir:   "/my/dir",
			wantRemaining: []string{},
		},
		{
			name:              "all three flags before tf subcommand",
			args:              []string{"--workspace", "my/ws", "--tf-path", "/bin/tf", "--work-dir", "/my/dir", "plan", "-var", "foo=bar"},
			wantWorkspacePath: "my/ws",
			wantTFPath:        "/bin/tf",
			wantWorkDir:       "/my/dir",
			wantRemaining:     []string{"plan", "-var", "foo=bar"},
		},
		{
			name:          "no flags - args unchanged",
			args:          []string{"plan", "-out", "tfplan"},
			wantRemaining: []string{"plan", "-out", "tfplan"},
		},
		{
			name:          "empty args",
			args:          []string{},
			wantRemaining: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &tfExecCommand{BaseCommand: &BaseCommand{}}
			fs := cmd.Flags()
			fs.SetOutput(io.Discard)
			_ = fs.Parse(tt.args)

			assert.Equal(t, tt.wantTFPath, ptr.ToString(cmd.tfPath))
			assert.Equal(t, tt.wantWorkspacePath, ptr.ToString(cmd.workspace))
			assert.Equal(t, tt.wantWorkDir, ptr.ToString(cmd.workDir))
			assert.Equal(t, tt.wantRemaining, fs.Args())
		})
	}
}

func TestBuildTFTokenEnvKey(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr bool
	}{
		{
			name:   "simple hostname",
			rawURL: "https://tharsis.example.com",
			want:   "TF_TOKEN_tharsis_example_com",
		},
		{
			name:   "hostname with port stripped",
			rawURL: "https://tharsis.example.com:8080",
			want:   "TF_TOKEN_tharsis_example_com",
		},
		{
			name:    "invalid URL",
			rawURL:  "://invalid",
			wantErr: true,
		},
		{
			name:    "empty host",
			rawURL:  "https://",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildTFTokenEnvKey(tt.rawURL)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAppendAuthToken(t *testing.T) {
	t.Run("nil token returns env unchanged", func(t *testing.T) {
		base := []string{"EXISTING=val"}
		result := appendAuthToken(base, "https://tharsis.example.com", nil)
		assert.Equal(t, base, result)
		for _, e := range result {
			assert.False(t, strings.HasPrefix(e, "TF_TOKEN_"), "should not add TF_TOKEN_* when token is nil")
		}
	})

	t.Run("empty tharsisURL returns env unchanged", func(t *testing.T) {
		base := []string{"EXISTING=val"}
		token := "mytoken"
		result := appendAuthToken(base, "", &token)
		assert.Equal(t, base, result)
	})

	t.Run("valid token and URL injects TF_TOKEN", func(t *testing.T) {
		base := []string{"EXISTING=val"}
		token := "mytoken"
		result := appendAuthToken(base, "https://tharsis.example.com", &token)
		assert.Contains(t, result, "TF_TOKEN_tharsis_example_com=mytoken")
		assert.Contains(t, result, "EXISTING=val")
	})

	t.Run("real-world Tharsis URL", func(t *testing.T) {
		token := "tok-abc123"
		result := appendAuthToken(nil, "https://tharsis.cts.infor.com", &token)
		assert.Contains(t, result, "TF_TOKEN_tharsis_cts_infor_com=tok-abc123")
	})

	t.Run("invalid URL is silently skipped", func(t *testing.T) {
		base := []string{"EXISTING=val"}
		token := "mytoken"
		result := appendAuthToken(base, "://bad-url", &token)
		assert.Equal(t, base, result)
	})
}

func TestHasTerraformFiles(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string // returns the dir to test; empty string means non-existent dir
		want    bool
		wantErr bool
	}{
		{
			name: "directory with a .tf file",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf"), []byte(""), 0o600))
				return dir
			},
			want: true,
		},
		{
			name: "directory without .tf files",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte(""), 0o600))
				return dir
			},
			want: false,
		},
		{
			name: "directory with .tf only in a subdirectory (non-recursive)",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				sub := filepath.Join(dir, "sub")
				require.NoError(t, os.Mkdir(sub, 0o750))
				require.NoError(t, os.WriteFile(filepath.Join(sub, "main.tf"), []byte(""), 0o600))
				return dir
			},
			want: false,
		},
		{
			name: "empty directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			want: false,
		},
		{
			name: "non-existent directory",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "does-not-exist")
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			got, err := hasTerraformFiles(dir)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsHCLVariableValue(t *testing.T) {
	strVar := &tfconfig.Variable{Type: "string"}
	mapVar := &tfconfig.Variable{Type: "map(string)"}
	untypedVar := &tfconfig.Variable{Type: ""}

	tests := []struct {
		name  string
		value string
		def   *tfconfig.Variable
		want  bool
	}{
		// Declared type is authoritative.
		{name: "declared string with JSON-looking value is a string", value: `{"user":"admin"}`, def: strVar, want: false},
		{name: "declared string with bool-looking value is a string", value: "true", def: strVar, want: false},
		{name: "declared map is HCL", value: `{"a":"b"}`, def: mapVar, want: true},
		// No declaration / unknown type: infer from value, primitives are strings.
		{name: "plain text without declaration is a string", value: "us-east-1", def: nil, want: false},
		{name: "bool literal without declaration is a string", value: "true", def: untypedVar, want: false},
		{name: "number literal without declaration is a string", value: "42", def: nil, want: false},
		{name: "quoted string literal without declaration is a string", value: `"hello"`, def: nil, want: false},
		{name: "object literal without declaration is HCL", value: `{"a":"b"}`, def: nil, want: true},
		{name: "array literal without declaration is HCL", value: `[1,2,3]`, def: nil, want: true},
		{name: "unparseable text without declaration is a string", value: "two words", def: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isHCLVariableValue(tt.value, tt.def))
		})
	}
}

func TestWriteTharsisAutoTfvars(t *testing.T) {
	newCmd := func() *tfExecCommand {
		return &tfExecCommand{BaseCommand: &BaseCommand{Logger: hclog.NewNullLogger()}}
	}

	// parseTfvars reads and parses the generated file into its attributes.
	parseTfvars := func(t *testing.T, dir string) map[string]cty.Value {
		t.Helper()
		filePath := filepath.Join(dir, tharsisAutoTfvarsFile)
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)
		f, diags := hclparse.NewParser().ParseHCL(data, filePath)
		require.False(t, diags.HasErrors(), "generated tfvars is not valid HCL: %s", diags.Error())
		attrs, diags := f.Body.JustAttributes()
		require.False(t, diags.HasErrors())
		out := map[string]cty.Value{}
		for k, a := range attrs {
			v, _ := a.Expr.Value(nil)
			out[k] = v
		}
		return out
	}

	t.Run("no module: classifies by value and skips environment vars", func(t *testing.T) {
		dir := t.TempDir()
		vars := []resolvedVar{
			{Key: "region", Value: "us-east-1", Category: "terraform"},
			{Key: "tags", Value: `{"team":"infra"}`, Category: "terraform"},
			{Key: "counts", Value: `[1,2,3]`, Category: "terraform"},
			{Key: "enabled", Value: "true", Category: "terraform"},
			{Key: "API_KEY", Value: "secret", Category: "environment"},
		}

		wrote, err := newCmd().writeTharsisAutoTfvars(dir, vars)
		require.NoError(t, err)
		assert.True(t, wrote)

		attrs := parseTfvars(t, dir)

		// Plain string stays a string.
		assert.Equal(t, "us-east-1", attrs["region"].AsString())
		// Collections/objects render as raw HCL.
		assert.True(t, attrs["tags"].Type().IsObjectType())
		assert.True(t, attrs["counts"].Type().IsTupleType() || attrs["counts"].Type().IsListType())
		// Without a declared type, a bool literal is treated as a string (matches the job executor).
		assert.Equal(t, cty.String, attrs["enabled"].Type())
		assert.Equal(t, "true", attrs["enabled"].AsString())
		// Environment variables are excluded.
		_, hasAPIKey := attrs["API_KEY"]
		assert.False(t, hasAPIKey)
	})

	t.Run("declared string type keeps a JSON-looking secret as a quoted string", func(t *testing.T) {
		dir := t.TempDir()
		// Declare variable types in the module config.
		require.NoError(t, os.WriteFile(filepath.Join(dir, "variables.tf"), []byte(`
variable "db_secret" { type = string }
variable "tags"      { type = map(string) }
`), 0o600))

		vars := []resolvedVar{
			{Key: "db_secret", Value: `{"user":"admin","pass":"p@ss"}`, Category: "terraform", Sensitive: true},
			{Key: "tags", Value: `{"team":"infra"}`, Category: "terraform"},
		}

		wrote, err := newCmd().writeTharsisAutoTfvars(dir, vars)
		require.NoError(t, err)
		assert.True(t, wrote)

		attrs := parseTfvars(t, dir)

		// The string-typed secret must remain a string equal to its raw JSON text,
		// NOT be parsed into an object.
		require.Equal(t, cty.String, attrs["db_secret"].Type())
		assert.Equal(t, `{"user":"admin","pass":"p@ss"}`, attrs["db_secret"].AsString())
		// The map-typed variable renders as a raw HCL object.
		assert.True(t, attrs["tags"].Type().IsObjectType())
	})

	t.Run("module loaded: variable not declared is skipped", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "variables.tf"), []byte(`
variable "declared" { type = string }
`), 0o600))

		vars := []resolvedVar{
			{Key: "declared", Value: "yes", Category: "terraform"},
			{Key: "undeclared", Value: "nope", Category: "terraform"},
		}

		wrote, err := newCmd().writeTharsisAutoTfvars(dir, vars)
		require.NoError(t, err)
		assert.True(t, wrote)

		attrs := parseTfvars(t, dir)
		assert.Equal(t, "yes", attrs["declared"].AsString())
		_, hasUndeclared := attrs["undeclared"]
		assert.False(t, hasUndeclared, "variables not declared in the module should be skipped")
	})

	t.Run("no terraform vars writes no file and removes stale file", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, tharsisAutoTfvarsFile)
		// Pre-existing stale file from a prior run.
		require.NoError(t, os.WriteFile(filePath, []byte(`old = "value"`), 0o600))

		wrote, err := newCmd().writeTharsisAutoTfvars(dir, []resolvedVar{
			{Key: "API_KEY", Value: "secret", Category: "environment"},
		})
		require.NoError(t, err)
		assert.False(t, wrote)

		_, statErr := os.Stat(filePath)
		assert.True(t, os.IsNotExist(statErr), "stale tfvars file should have been removed")
	})

	t.Run("empty vars writes no file", func(t *testing.T) {
		dir := t.TempDir()
		wrote, err := newCmd().writeTharsisAutoTfvars(dir, nil)
		require.NoError(t, err)
		assert.False(t, wrote)
		_, statErr := os.Stat(filepath.Join(dir, tharsisAutoTfvarsFile))
		assert.True(t, os.IsNotExist(statErr))
	})
}

func TestResolveTerraformBinaryPathFallback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake executable setup is POSIX-specific")
	}

	newCmd := func() *tfExecCommand {
		return &tfExecCommand{BaseCommand: &BaseCommand{Logger: hclog.NewNullLogger()}}
	}

	t.Run("no applied run falls back to terraform on PATH", func(t *testing.T) {
		binDir := t.TempDir()
		fakeTF := filepath.Join(binDir, "terraform")
		require.NoError(t, os.WriteFile(fakeTF, []byte("#!/bin/sh\n"), 0o700))
		t.Setenv("PATH", binDir)

		got, err := newCmd().resolveTerraformBinary("")
		require.NoError(t, err)
		assert.Equal(t, fakeTF, got)
	})

	t.Run("no applied run and no terraform on PATH errors mentioning -tf-path", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir()) // empty dir, no terraform
		_, err := newCmd().resolveTerraformBinary("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "-tf-path")
	})

	t.Run("explicit -tf-path takes precedence over version", func(t *testing.T) {
		binDir := t.TempDir()
		fakeTF := filepath.Join(binDir, "terraform")
		require.NoError(t, os.WriteFile(fakeTF, []byte("#!/bin/sh\n"), 0o700))

		c := newCmd()
		c.tfPath = ptr.String(fakeTF)
		got, err := c.resolveTerraformBinary("")
		require.NoError(t, err)
		assert.Equal(t, fakeTF, got)
	})
}

func TestPersistentWorkdir(t *testing.T) {
	const (
		url1  = "https://tharsis.example.com"
		url2  = "https://other.example.com"
		path1 = "group/workspace"
		path2 = "group/other"
	)

	t.Run("same URL and path always returns the same result", func(t *testing.T) {
		a, err := persistentWorkdir(url1, path1)
		require.NoError(t, err)
		b, err := persistentWorkdir(url1, path1)
		require.NoError(t, err)
		assert.Equal(t, a, b)
	})

	t.Run("different URL produces different result", func(t *testing.T) {
		a, err := persistentWorkdir(url1, path1)
		require.NoError(t, err)
		b, err := persistentWorkdir(url2, path1)
		require.NoError(t, err)
		assert.NotEqual(t, a, b)
	})

	t.Run("different path produces different result", func(t *testing.T) {
		a, err := persistentWorkdir(url1, path1)
		require.NoError(t, err)
		b, err := persistentWorkdir(url1, path2)
		require.NoError(t, err)
		assert.NotEqual(t, a, b)
	})

	t.Run("result contains full 64-char hex string", func(t *testing.T) {
		result, err := persistentWorkdir(url1, path1)
		require.NoError(t, err)
		// The last path component is the hex-encoded SHA-256 digest (32 bytes = 64 hex chars).
		base := filepath.Base(result)
		assert.Len(t, base, 64, "expected full SHA-256 hex digest (64 chars), got %q", base)
	})

	t.Run("empty hostname returns error", func(t *testing.T) {
		_, err := persistentWorkdir("https://", path1)
		assert.Error(t, err)
	})

	t.Run("invalid URL returns error", func(t *testing.T) {
		_, err := persistentWorkdir("://bad", path1)
		assert.Error(t, err)
	})
}

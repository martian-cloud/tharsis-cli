package varparser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTerraformVariables(t *testing.T) {
	t.Run("parses key=value pairs", func(t *testing.T) {
		vp := NewVariableParser(nil, false)
		vars, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{
			TfVariables: []string{"region=us-east-1", "env=prod"},
		})

		require.NoError(t, err)
		assert.Len(t, vars, 2)
		assertContainsVar(t, vars, "region", "us-east-1", TerraformVariableCategory)
		assertContainsVar(t, vars, "env", "prod", TerraformVariableCategory)
	})

	t.Run("last value wins for duplicate keys", func(t *testing.T) {
		vp := NewVariableParser(nil, false)
		vars, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{
			TfVariables: []string{"region=us-east-1", "region=us-west-2"},
		})

		require.NoError(t, err)
		assert.Len(t, vars, 1)
		assertContainsVar(t, vars, "region", "us-west-2", TerraformVariableCategory)
	})

	t.Run("value can contain equals sign", func(t *testing.T) {
		vp := NewVariableParser(nil, false)
		vars, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{
			TfVariables: []string{"config=key=value"},
		})

		require.NoError(t, err)
		assert.Len(t, vars, 1)
		assertContainsVar(t, vars, "config", "key=value", TerraformVariableCategory)
	})

	t.Run("value can be empty", func(t *testing.T) {
		vp := NewVariableParser(nil, false)
		vars, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{
			TfVariables: []string{"empty="},
		})

		require.NoError(t, err)
		assert.Len(t, vars, 1)
		assertContainsVar(t, vars, "empty", "", TerraformVariableCategory)
	})

	t.Run("errors on missing equals", func(t *testing.T) {
		vp := NewVariableParser(nil, false)
		_, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{
			TfVariables: []string{"noequals"},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a key=value pair")
	})

	t.Run("errors on empty key", func(t *testing.T) {
		vp := NewVariableParser(nil, false)
		_, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{
			TfVariables: []string{"=value"},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key")
	})

	t.Run("errors on key with spaces", func(t *testing.T) {
		vp := NewVariableParser(nil, false)
		_, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{
			TfVariables: []string{"bad key=value"},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key")
	})

	t.Run("skips empty strings", func(t *testing.T) {
		vp := NewVariableParser(nil, false)
		vars, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{
			TfVariables: []string{"", "region=us-east-1", "  "},
		})

		require.NoError(t, err)
		assert.Len(t, vars, 1)
	})

	t.Run("reads TF_VAR_ from environment", func(t *testing.T) {
		t.Setenv("TF_VAR_region", "us-east-1")

		vp := NewVariableParser(nil, true)
		vars, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{})

		require.NoError(t, err)
		assertContainsVar(t, vars, "region", "us-east-1", TerraformVariableCategory)
	})

	t.Run("flag variables override environment", func(t *testing.T) {
		t.Setenv("TF_VAR_region", "us-east-1")

		vp := NewVariableParser(nil, true)
		vars, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{
			TfVariables: []string{"region=us-west-2"},
		})

		require.NoError(t, err)
		assertContainsVar(t, vars, "region", "us-west-2", TerraformVariableCategory)
	})

	t.Run("parses tfvars file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.tfvars")
		require.NoError(t, os.WriteFile(path, []byte(`region = "us-east-1"`+"\n"), 0o600))

		vp := NewVariableParser(nil, false)
		vars, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{
			TfVarFilePaths: []string{path},
		})

		require.NoError(t, err)
		assertContainsVar(t, vars, "region", "us-east-1", TerraformVariableCategory)
	})

	t.Run("parses tfvars.json file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.tfvars.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"region": "us-east-1"}`), 0o600))

		vp := NewVariableParser(nil, false)
		vars, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{
			TfVarFilePaths: []string{path},
		})

		require.NoError(t, err)
		assertContainsVar(t, vars, "region", "us-east-1", TerraformVariableCategory)
	})

	t.Run("errors on invalid file extension", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		require.NoError(t, os.WriteFile(path, []byte("region=us-east-1"), 0o600))

		vp := NewVariableParser(nil, false)
		_, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{
			TfVarFilePaths: []string{path},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file extension")
	})

	t.Run("reads terraform.tfvars from module directory", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "terraform.tfvars"), []byte(`region = "us-east-1"`+"\n"), 0o600))

		vp := NewVariableParser(&dir, false)
		vars, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{})

		require.NoError(t, err)
		assertContainsVar(t, vars, "region", "us-east-1", TerraformVariableCategory)
	})

	t.Run("reads auto.tfvars from module directory", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "custom.auto.tfvars"), []byte(`env = "prod"`+"\n"), 0o600))

		vp := NewVariableParser(&dir, false)
		vars, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{})

		require.NoError(t, err)
		assertContainsVar(t, vars, "env", "prod", TerraformVariableCategory)
	})

	t.Run("no module directory skips file-based parsing", func(t *testing.T) {
		vp := NewVariableParser(nil, false)
		vars, err := vp.ParseTerraformVariables(&ParseTerraformVariablesInput{})

		require.NoError(t, err)
		assert.Empty(t, vars)
	})
}

func TestParseEnvironmentVariables(t *testing.T) {
	t.Run("parses key=value pairs", func(t *testing.T) {
		vp := NewVariableParser(nil, false)
		vars, err := vp.ParseEnvironmentVariables(&ParseEnvironmentVariablesInput{
			EnvVariables: []string{"DB_HOST=localhost", "DB_PORT=5432"},
		})

		require.NoError(t, err)
		assert.Len(t, vars, 2)
		assertContainsVar(t, vars, "DB_HOST", "localhost", EnvironmentVariableCategory)
		assertContainsVar(t, vars, "DB_PORT", "5432", EnvironmentVariableCategory)
	})

	t.Run("reads env var file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "vars.env")
		require.NoError(t, os.WriteFile(path, []byte("DB_HOST=localhost\nDB_PORT=5432\n"), 0o600))

		vp := NewVariableParser(nil, false)
		vars, err := vp.ParseEnvironmentVariables(&ParseEnvironmentVariablesInput{
			EnvVarFilePaths: []string{path},
		})

		require.NoError(t, err)
		assert.Len(t, vars, 2)
		assertContainsVar(t, vars, "DB_HOST", "localhost", EnvironmentVariableCategory)
	})

	t.Run("flag variables override file variables", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "vars.env")
		require.NoError(t, os.WriteFile(path, []byte("DB_HOST=filehost\n"), 0o600))

		vp := NewVariableParser(nil, false)
		vars, err := vp.ParseEnvironmentVariables(&ParseEnvironmentVariablesInput{
			EnvVarFilePaths: []string{path},
			EnvVariables:    []string{"DB_HOST=flaghost"},
		})

		require.NoError(t, err)
		assert.Len(t, vars, 1)
		assertContainsVar(t, vars, "DB_HOST", "flaghost", EnvironmentVariableCategory)
	})

	t.Run("errors on nonexistent file", func(t *testing.T) {
		vp := NewVariableParser(nil, false)
		_, err := vp.ParseEnvironmentVariables(&ParseEnvironmentVariablesInput{
			EnvVarFilePaths: []string{"/nonexistent/vars.env"},
		})

		assert.Error(t, err)
	})
}

func assertContainsVar(t *testing.T, vars []Variable, key, value string, category VariableCategory) {
	t.Helper()
	for _, v := range vars {
		if v.Key == key {
			assert.Equal(t, value, v.Value)
			assert.Equal(t, category, v.Category)
			return
		}
	}

	t.Errorf("variable %q not found", key)
}

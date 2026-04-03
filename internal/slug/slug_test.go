package slug

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSlug(t *testing.T) {
	t.Run("creates slug from directory", func(t *testing.T) {
		srcDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "main.tf"), []byte(`resource "null" "test" {}`), 0o600))

		slugPath := filepath.Join(t.TempDir(), "test.tar.gz")
		s, err := NewSlug(srcDir, slugPath)

		require.NoError(t, err)
		assert.Equal(t, slugPath, s.SlugPath)
		assert.NotEmpty(t, s.SHASum)
		assert.Greater(t, s.Size, int64(0))

		info, err := os.Stat(slugPath)
		require.NoError(t, err)
		assert.Greater(t, info.Size(), int64(0))
	})

	t.Run("errors on nonexistent directory", func(t *testing.T) {
		_, err := NewSlug("/nonexistent/path", filepath.Join(t.TempDir(), "test.tar.gz"))

		assert.Error(t, err)
	})

	t.Run("errors on file instead of directory", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "notadir.txt")
		require.NoError(t, os.WriteFile(file, []byte("hello"), 0o600))

		_, err := NewSlug(file, filepath.Join(t.TempDir(), "test.tar.gz"))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})

	t.Run("produces deterministic checksum", func(t *testing.T) {
		srcDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "main.tf"), []byte(`resource "null" "test" {}`), 0o600))

		s1, err := NewSlug(srcDir, filepath.Join(t.TempDir(), "a.tar.gz"))
		require.NoError(t, err)

		s2, err := NewSlug(srcDir, filepath.Join(t.TempDir(), "b.tar.gz"))
		require.NoError(t, err)

		assert.Equal(t, s1.SHASum, s2.SHASum)
	})
}

func TestSlugOpen(t *testing.T) {
	srcDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "main.tf"), []byte("# test"), 0o600))

	slugPath := filepath.Join(t.TempDir(), "test.tar.gz")
	s, err := NewSlug(srcDir, slugPath)
	require.NoError(t, err)

	reader, err := s.Open()
	require.NoError(t, err)
	defer reader.Close()

	buf := make([]byte, 16)
	n, err := reader.Read(buf)

	assert.NoError(t, err)
	assert.Greater(t, n, 0)
}

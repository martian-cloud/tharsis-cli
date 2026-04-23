// Package slug is used to create Terraform slugs
package slug

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-slug"
)

// Slug is a tar.gz file containing a Terraform configuration
type Slug struct {
	SlugPath string
	SHASum   []byte
	Size     int64
}

// NewSlug creates a slug from the srcDir, writing it to a temporary file.
// Files are copied to a temp directory with normalized timestamps to ensure
// deterministic slug digests regardless of when files were checked out.
// The caller is responsible for removing SlugPath when done.
func NewSlug(srcDir string) (*Slug, error) {
	// Check the directory path.
	stat, err := os.Stat(srcDir)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", srcDir)
	}

	// Copy files to a temp directory and normalize timestamps.
	tmpDir, err := copyAndNormalizeTimestamps(srcDir)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare deterministic source: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a temporary file for the slug.
	slugFile, err := os.CreateTemp("", "terraform-slug-*.tar.gz")
	if err != nil {
		return nil, fmt.Errorf("failed to create slug file: %w", err)
	}

	checksum := sha256.New()

	meta, err := slug.Pack(tmpDir, io.MultiWriter(slugFile, checksum), true)
	if err != nil {
		slugFile.Close()
		os.Remove(slugFile.Name())
		return nil, err
	}

	if err := slugFile.Close(); err != nil {
		os.Remove(slugFile.Name())
		return nil, fmt.Errorf("failed to write slug file: %w", err)
	}

	return &Slug{SlugPath: slugFile.Name(), Size: meta.Size, SHASum: checksum.Sum(nil)}, nil
}

// copyAndNormalizeTimestamps copies srcDir to a temp directory and sets all
// file modification times to the Unix epoch for deterministic packing.
func copyAndNormalizeTimestamps(srcDir string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "slug-*")
	if err != nil {
		return "", err
	}

	epoch := time.Unix(0, 0)

	err = filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		dst := filepath.Join(tmpDir, rel)

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			return os.MkdirAll(dst, info.Mode())
		}

		if d.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(target, dst)
		}

		return copyFile(path, dst)
	})
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	// Set all timestamps to epoch after all files are written.
	// Use WalkDir and skip symlinks since os.Chtimes follows
	// symlinks and would modify targets outside the temp directory.
	err = filepath.WalkDir(tmpDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		return os.Chtimes(path, epoch, epoch)
	})
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	return tmpDir, nil
}

// Open opens the slug file for reading
func (s *Slug) Open() (io.ReadCloser, error) {
	return os.Open(s.SlugPath)
}

// copyFile copies a single file from src to dst, preserving permissions.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		return err
	}

	return out.Close()
}

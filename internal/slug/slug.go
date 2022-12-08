// Package slug is used to create Terraform slugs
package slug

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/go-slug"
)

const (
	// options for creating a temporary tarfile
	tarFlagWrite = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	tarMode      = 0600
)

// Slug is a tar.gz file containing a Terraform configuration
type Slug struct {
	SlugPath string
	SHASum   []byte
	Size     int64
}

// NewSlug creates a slug from the srcDir and outputs it to the slugPath
func NewSlug(srcDir string, slugPath string) (*Slug, error) {
	// Check the directory path.
	stat, err := os.Stat(srcDir)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", srcDir)
	}

	// Open a writer to the temporary tar.gz file.
	fileWriter, err := os.OpenFile(slugPath, tarFlagWrite, tarMode)
	if err != nil {
		return nil, err
	}
	defer fileWriter.Close()

	checksum := sha256.New()

	meta, err := slug.Pack(srcDir, io.MultiWriter(fileWriter, checksum), true)
	if err != nil {
		return nil, err
	}

	return &Slug{SlugPath: slugPath, Size: meta.Size, SHASum: checksum.Sum(nil)}, nil
}

// Open opens the slug file for reading
func (s *Slug) Open() (io.ReadCloser, error) {
	return os.Open(s.SlugPath)
}

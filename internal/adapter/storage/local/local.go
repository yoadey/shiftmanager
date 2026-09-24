// Package local implements port.MediaStorage by writing to a local directory
// on disk, served back out via the /uploads/* static route
// (internal/adapter/http/router.go). This is the historical, default
// behavior; T-013 adds S3-compatible object storage as an alternative.
package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/yoadey/shiftmanager/internal/port"
)

// Storage implements port.MediaStorage against a local directory.
type Storage struct {
	dir        string
	publicBase string
}

var _ port.MediaStorage = (*Storage)(nil)

// New creates a Storage rooted at dir, serving files back out under
// publicBase+"/uploads/". dir defaults to "./uploads" if empty.
func New(dir, publicBase string) *Storage {
	if dir == "" {
		dir = "./uploads"
	}
	return &Storage{dir: dir, publicBase: publicBase}
}

// Put writes r to a file named name under the storage directory.
func (s *Storage) Put(ctx context.Context, name string, r io.Reader, size int64, contentType string) (string, error) {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return "", fmt.Errorf("create upload dir: %w", err)
	}
	dst := filepath.Join(s.dir, name)
	out, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	written, copyErr := io.Copy(out, r)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(dst)
		return "", fmt.Errorf("write file: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(dst)
		return "", fmt.Errorf("close file: %w", closeErr)
	}
	if written != size {
		_ = os.Remove(dst)
		return "", fmt.Errorf("short write: expected %d bytes, wrote %d", size, written)
	}
	return strings.TrimRight(s.publicBase, "/") + "/uploads/" + name, nil
}

// Delete removes the file a previous Put returned as url. Only ever touches
// files under our own upload dir, never an arbitrary path from the URL
// (filepath.Base strips any directory components).
func (s *Storage) Delete(ctx context.Context, url string) error {
	name := filepath.Base(url)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return nil
	}
	if err := os.Remove(filepath.Join(s.dir, name)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

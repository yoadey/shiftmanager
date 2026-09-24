package local

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutWritesFileAndReturnsURL(t *testing.T) {
	dir := t.TempDir()
	s := New(dir, "http://localhost:8080")

	content := "hello world"
	url, err := s.Put(context.Background(), "logo-abc.png", strings.NewReader(content), int64(len(content)), "image/png")
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/uploads/logo-abc.png", url)

	got, err := os.ReadFile(filepath.Join(dir, "logo-abc.png"))
	require.NoError(t, err)
	assert.Equal(t, content, string(got))
}

func TestPutTrimsTrailingSlashOnPublicBase(t *testing.T) {
	s := New(t.TempDir(), "http://localhost:8080/")
	url, err := s.Put(context.Background(), "x.png", strings.NewReader("a"), 1, "image/png")
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/uploads/x.png", url)
}

func TestPutCreatesUploadDirIfMissing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "uploads")
	s := New(dir, "http://localhost")
	_, err := s.Put(context.Background(), "x.png", strings.NewReader("a"), 1, "image/png")
	require.NoError(t, err)
	_, err = os.Stat(dir)
	require.NoError(t, err)
}

func TestPutRejectsShortWrite(t *testing.T) {
	dir := t.TempDir()
	s := New(dir, "http://localhost")

	// Declared size (10) doesn't match what the reader actually provides (1 byte).
	_, err := s.Put(context.Background(), "x.png", strings.NewReader("a"), 10, "image/png")
	assert.Error(t, err)

	// The partially-written file must not be left behind.
	_, statErr := os.Stat(filepath.Join(dir, "x.png"))
	assert.True(t, os.IsNotExist(statErr))
}

func TestDeleteRemovesFile(t *testing.T) {
	dir := t.TempDir()
	s := New(dir, "http://localhost")
	url, err := s.Put(context.Background(), "x.png", strings.NewReader("a"), 1, "image/png")
	require.NoError(t, err)

	require.NoError(t, s.Delete(context.Background(), url))
	_, statErr := os.Stat(filepath.Join(dir, "x.png"))
	assert.True(t, os.IsNotExist(statErr))
}

func TestDeleteOfMissingFileIsNoop(t *testing.T) {
	s := New(t.TempDir(), "http://localhost")
	err := s.Delete(context.Background(), "http://localhost/uploads/does-not-exist.png")
	assert.NoError(t, err)
}

// Delete must only ever remove a file directly inside the storage
// directory, never honor path components smuggled in via the URL.
func TestDeleteIgnoresPathTraversal(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.txt")
	require.NoError(t, os.WriteFile(outside, []byte("do not delete me"), 0o644))

	s := New(dir, "http://localhost")
	err := s.Delete(context.Background(), "http://localhost/uploads/../../../"+outside)
	require.NoError(t, err)

	_, statErr := os.Stat(outside)
	assert.NoError(t, statErr, "file outside the upload dir must survive Delete")
}

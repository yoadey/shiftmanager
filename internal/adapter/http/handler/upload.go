package handler

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// uploadOverhead is added on top of the caller's file-size limit before it's
// applied to the whole request body via http.MaxBytesReader. Multipart
// requests carry boundary markers and per-part headers in addition to the
// file's own bytes, so capping the raw body at exactly maxSize would reject
// a file whose content is within the documented limit but whose total
// request is a few hundred bytes larger. This is only an outer safety net
// against pathological bodies; the actual maxSize limit is still enforced
// precisely below via the parsed part's own size.
const uploadOverhead = 64 * 1024

// receiveUpload reads a single "file" multipart field, rejects it if its own
// content exceeds maxSize, validates it via isAllowed, and writes it to
// uploadDir under a name of namePrefix+<uuid>+ext. Shared by the logo
// (B-004) and event attachment (V-008) upload endpoints.
func receiveUpload(w http.ResponseWriter, r *http.Request, uploadDir string, maxSize int64, namePrefix string, isAllowed func(ext, contentType string) bool, rejectMsg string) (storedName, originalName, contentType string, size int64, ok bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize+uploadOverhead)
	if err := r.ParseMultipartForm(maxSize + uploadOverhead); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "file exceeds size limit")
			return "", "", "", 0, false
		}
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return "", "", "", 0, false
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return "", "", "", 0, false
	}
	defer file.Close()

	if hdr.Size > maxSize {
		writeError(w, http.StatusRequestEntityTooLarge, "file exceeds size limit")
		return "", "", "", 0, false
	}

	ext := strings.ToLower(filepath.Ext(hdr.Filename))
	contentType = hdr.Header.Get("Content-Type")
	if !isAllowed(ext, contentType) {
		writeError(w, http.StatusUnsupportedMediaType, rejectMsg)
		return "", "", "", 0, false
	}

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "cannot create upload dir")
		return "", "", "", 0, false
	}
	storedName = namePrefix + uuid.New().String() + ext
	dst := filepath.Join(uploadDir, storedName)
	out, err := os.Create(dst)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot store file")
		return "", "", "", 0, false
	}
	written, err := io.Copy(out, file)
	_ = out.Close()
	if err != nil {
		_ = os.Remove(dst)
		writeError(w, http.StatusInternalServerError, "cannot write file")
		return "", "", "", 0, false
	}
	return storedName, hdr.Filename, contentType, written, true
}

// Package localstore saves Rider KYC document files on local disk — same
// approach as the driver service's KYC document storage (no cloud upload),
// duplicated here as its own small package rather than imported across the
// driver/user module boundary (mirrors this codebase's existing precedent
// of each service owning its own small infrastructure helpers).
package localstore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	domainerrors "github.com/fairride/shared/errors"
)

// DocumentStore writes/reads Rider KYC document files under a base
// directory, laid out as {baseDir}/{userID}/{docType}_{random}{ext}.
type DocumentStore struct {
	baseDir string
}

// NewDocumentStore constructs a DocumentStore rooted at baseDir. baseDir is
// created on first Save if it doesn't already exist.
func NewDocumentStore(baseDir string) *DocumentStore {
	return &DocumentStore{baseDir: baseDir}
}

// Save writes data to disk under userID/docType and returns the path
// relative to baseDir — the value to persist in RiderVerification's
// CCCDFrontPath/CCCDBackPath. Never echo this path back in an API response.
func (s *DocumentStore) Save(_ context.Context, userID string, docType string, ext string, data io.Reader) (string, error) {
	userDir := sanitize(userID)
	dir := filepath.Join(s.baseDir, userDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", domainerrors.Internal("rider kyc: create storage directory failed").WithMeta("error", err.Error())
	}
	token, err := randomToken()
	if err != nil {
		return "", domainerrors.Internal("rider kyc: generate filename failed").WithMeta("error", err.Error())
	}
	filename := fmt.Sprintf("%s_%s%s", sanitize(docType), token, sanitizeExt(ext))
	relPath := filepath.Join(userDir, filename)
	fullPath := filepath.Join(s.baseDir, relPath)

	f, err := os.Create(fullPath)
	if err != nil {
		return "", domainerrors.Internal("rider kyc: create file failed").WithMeta("error", err.Error())
	}
	defer f.Close()
	if _, err := io.Copy(f, data); err != nil {
		return "", domainerrors.Internal("rider kyc: write file failed").WithMeta("error", err.Error())
	}
	return filepath.ToSlash(relPath), nil
}

// Open opens a previously-saved file for reading.
func (s *DocumentStore) Open(relPath string) (*os.File, error) {
	fullPath := filepath.Join(s.baseDir, filepath.FromSlash(relPath))
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, domainerrors.NotFound("rider kyc: document file not found")
	}
	return f, nil
}

func sanitize(s string) string {
	s = strings.ReplaceAll(s, "..", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "\\", "")
	return s
}

func sanitizeExt(ext string) string {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	switch ext {
	case "jpg", "jpeg", "png", "pdf", "webp":
		return "." + ext
	default:
		return ".bin"
	}
}

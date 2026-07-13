// Package localstore saves KYC document files on local disk — no cloud
// upload (Phần 4: "Không upload cloud. Lưu local giống avatar hiện tại.").
// Kept as its own small package (not tied to Postgres) so a future
// swap to S3/GCS/Azure Blob for production deployment only means writing a
// new implementation of the same two methods, never touching the app layer
// that calls it.
package localstore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/google/uuid"
)

// DocumentStore writes/reads KYC document files under a base directory,
// laid out as {baseDir}/{driverID}/{docType}_{uuid}{ext}.
type DocumentStore struct {
	baseDir string
}

// NewDocumentStore constructs a DocumentStore rooted at baseDir. baseDir is
// created on first Save if it doesn't already exist.
func NewDocumentStore(baseDir string) *DocumentStore {
	return &DocumentStore{baseDir: baseDir}
}

// Save writes data to disk under driverID/docType and returns the path
// relative to baseDir — the value to persist in KYCDocument.StoragePath.
// The relative path is an internal implementation detail: callers (the
// gateway HTTP layer) must never echo it back in an API response.
func (s *DocumentStore) Save(_ context.Context, driverID string, docType string, ext string, data io.Reader) (string, error) {
	driverDir := sanitize(driverID)
	dir := filepath.Join(s.baseDir, driverDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", domainerrors.Internal("kyc: create storage directory failed").WithMeta("error", err.Error())
	}
	filename := fmt.Sprintf("%s_%s%s", sanitize(docType), uuid.NewString(), sanitizeExt(ext))
	relPath := filepath.Join(driverDir, filename)
	fullPath := filepath.Join(s.baseDir, relPath)

	f, err := os.Create(fullPath)
	if err != nil {
		return "", domainerrors.Internal("kyc: create file failed").WithMeta("error", err.Error())
	}
	defer f.Close()
	if _, err := io.Copy(f, data); err != nil {
		return "", domainerrors.Internal("kyc: write file failed").WithMeta("error", err.Error())
	}
	return filepath.ToSlash(relPath), nil
}

// Open opens a previously-saved file for reading — used only by the
// admin-only document-review endpoint. relPath must be a value that was
// previously returned by Save (read from a KYCDocument row looked up
// server-side); never accept a raw path from client input.
func (s *DocumentStore) Open(relPath string) (*os.File, error) {
	fullPath := filepath.Join(s.baseDir, filepath.FromSlash(relPath))
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, domainerrors.NotFound("kyc: document file not found")
	}
	return f, nil
}

// sanitize strips path-traversal characters from a value used as a path
// segment (driverID/docType are both server-controlled enums/IDs already,
// but this is cheap defense in depth).
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

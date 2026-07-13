package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

const kycDocumentFields = `id, driver_id, document_type, storage_path, content_type, version, expires_at, uploaded_by, uploaded_at`

// KYCDocumentRepository is the PostgreSQL implementation of repository.KYCDocumentRepository.
type KYCDocumentRepository struct {
	pool *pgxpool.Pool
}

var _ repository.KYCDocumentRepository = (*KYCDocumentRepository)(nil)

func NewKYCDocumentRepository(pool *pgxpool.Pool) *KYCDocumentRepository {
	return &KYCDocumentRepository{pool: pool}
}

// Save always INSERTs a new row (Phần 4 — versioning: never overwrites a
// previous upload). Callers are responsible for computing the next Version
// number (see app.UploadKYCDocumentUseCase, which reads the current max via
// ListVersionsByDriverAndType before saving).
func (r *KYCDocumentRepository) Save(ctx context.Context, d *entity.KYCDocument) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO kyc_documents (`+kycDocumentFields+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		d.ID, d.DriverID, string(d.DocumentType), d.StoragePath, d.ContentType, d.Version, d.ExpiresAt, d.UploadedBy, d.UploadedAt.UTC(),
	)
	if err != nil {
		return domainerrors.Internal("kyc_document: save failed").WithMeta("error", err.Error())
	}
	return nil
}

// FindByDriverAndType returns the latest version.
func (r *KYCDocumentRepository) FindByDriverAndType(ctx context.Context, driverID string, docType entity.DocumentType) (*entity.KYCDocument, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+kycDocumentFields+` FROM kyc_documents
		WHERE driver_id = $1 AND document_type = $2
		ORDER BY version DESC LIMIT 1`, driverID, string(docType))
	return scanKYCDocument(row)
}

func (r *KYCDocumentRepository) FindByID(ctx context.Context, id string) (*entity.KYCDocument, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+kycDocumentFields+` FROM kyc_documents WHERE id = $1`, id)
	return scanKYCDocument(row)
}

// ListByDriverID returns the latest version of every document type
// driverID has uploaded (one row per type — DISTINCT ON + explicit
// ordering, not full history; see ListVersionsByDriverAndType for that).
func (r *KYCDocumentRepository) ListByDriverID(ctx context.Context, driverID string) ([]*entity.KYCDocument, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ON (document_type) `+kycDocumentFields+`
		FROM kyc_documents
		WHERE driver_id = $1
		ORDER BY document_type, version DESC`, driverID)
	if err != nil {
		return nil, domainerrors.Internal("kyc_document: list failed").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var out []*entity.KYCDocument
	for rows.Next() {
		d, err := scanKYCDocument(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("kyc_document: rows error").WithMeta("error", err.Error())
	}
	return out, nil
}

// ListVersionsByDriverAndType returns every version ever uploaded, newest first.
func (r *KYCDocumentRepository) ListVersionsByDriverAndType(ctx context.Context, driverID string, docType entity.DocumentType) ([]*entity.KYCDocument, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+kycDocumentFields+` FROM kyc_documents
		WHERE driver_id = $1 AND document_type = $2
		ORDER BY version DESC`, driverID, string(docType))
	if err != nil {
		return nil, domainerrors.Internal("kyc_document: list versions failed").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var out []*entity.KYCDocument
	for rows.Next() {
		d, err := scanKYCDocument(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("kyc_document: rows error").WithMeta("error", err.Error())
	}
	return out, nil
}

func scanKYCDocument(row rowScanner) (*entity.KYCDocument, error) {
	var (
		id, driverID, docType, storagePath, contentType string
		version                                         int
		expiresAt                                       *time.Time
		uploadedBy                                      string
		uploadedAt                                      time.Time
	)
	err := row.Scan(&id, &driverID, &docType, &storagePath, &contentType, &version, &expiresAt, &uploadedBy, &uploadedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("kyc document not found")
		}
		return nil, domainerrors.Internal("kyc_document: scan failed").WithMeta("error", err.Error())
	}
	return entity.ReconstituteKYCDocument(id, driverID, entity.DocumentType(docType), storagePath, contentType, version, expiresAt, uploadedBy, uploadedAt.UTC()), nil
}

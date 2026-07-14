package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/fairride/user/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	riderVerificationFields = `id, user_id, full_name, date_of_birth, national_id_number,
		cccd_front_path, cccd_back_path, status, submitted_at, approved_at, rejected_at,
		reviewer, reject_reason, review_mode, ai_confidence, ocr_result, vision_result,
		created_at, updated_at`
	riderVerificationTable  = `rider_verifications`
	riderVerificationSelect = `SELECT ` + riderVerificationFields + ` FROM ` + riderVerificationTable
)

// RiderVerificationRepository is the PostgreSQL implementation of repository.RiderVerificationRepository.
type RiderVerificationRepository struct {
	pool *pgxpool.Pool
}

func NewRiderVerificationRepository(pool *pgxpool.Pool) *RiderVerificationRepository {
	return &RiderVerificationRepository{pool: pool}
}

func (r *RiderVerificationRepository) FindByUserID(ctx context.Context, userID string) (*entity.RiderVerification, error) {
	return r.queryOne(ctx, riderVerificationSelect+` WHERE user_id = $1`, userID)
}

func (r *RiderVerificationRepository) ListByStatus(ctx context.Context, status entity.RiderKYCStatus) ([]*entity.RiderVerification, error) {
	rows, err := r.pool.Query(ctx, riderVerificationSelect+` WHERE status = $1 ORDER BY created_at`, string(status))
	if err != nil {
		return nil, domainerrors.Internal("query rider verifications").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var result []*entity.RiderVerification
	for rows.Next() {
		v, err := scanRiderVerification(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("iterate rider verifications").WithMeta("error", err.Error())
	}
	return result, nil
}

func (r *RiderVerificationRepository) Save(ctx context.Context, v *entity.RiderVerification) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO `+riderVerificationTable+`
			(id, user_id, full_name, date_of_birth, national_id_number,
			 cccd_front_path, cccd_back_path, status, submitted_at, approved_at, rejected_at,
			 reviewer, reject_reason, review_mode, ai_confidence, ocr_result, vision_result,
			 created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		ON CONFLICT (id) DO UPDATE SET
			full_name          = EXCLUDED.full_name,
			date_of_birth      = EXCLUDED.date_of_birth,
			national_id_number = EXCLUDED.national_id_number,
			cccd_front_path    = EXCLUDED.cccd_front_path,
			cccd_back_path     = EXCLUDED.cccd_back_path,
			status             = EXCLUDED.status,
			submitted_at       = EXCLUDED.submitted_at,
			approved_at        = EXCLUDED.approved_at,
			rejected_at        = EXCLUDED.rejected_at,
			reviewer           = EXCLUDED.reviewer,
			reject_reason      = EXCLUDED.reject_reason,
			review_mode        = EXCLUDED.review_mode,
			ai_confidence      = EXCLUDED.ai_confidence,
			ocr_result         = EXCLUDED.ocr_result,
			vision_result      = EXCLUDED.vision_result,
			updated_at         = EXCLUDED.updated_at
	`,
		v.ID, v.UserID, v.FullName, nullableTime(v.DateOfBirth), v.NationalIDNumber,
		v.CCCDFrontPath, v.CCCDBackPath, string(v.Status), v.SubmittedAt, v.ApprovedAt, v.RejectedAt,
		v.Reviewer, v.RejectReason, string(v.ReviewMode), v.AIConfidence, v.OCRResult, v.VisionResult,
		v.CreatedAt, v.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.AlreadyExists("rider verification already exists for this user")
		}
		return domainerrors.Internal("save rider verification").WithMeta("error", err.Error())
	}
	return nil
}

func nullableTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func (r *RiderVerificationRepository) queryOne(ctx context.Context, sql string, args ...any) (*entity.RiderVerification, error) {
	row := r.pool.QueryRow(ctx, sql, args...)
	v, err := scanRiderVerification(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("rider verification not found")
		}
		return nil, err
	}
	return v, nil
}

// rowScanner is satisfied by both pgx.Row (QueryRow) and pgx.Rows (Query) —
// lets ListByStatus and queryOne share one Scan call site.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanRiderVerification(row rowScanner) (*entity.RiderVerification, error) {
	var id, userID, fullName, nationalIDNumber, cccdFrontPath, cccdBackPath, status, reviewer, rejectReason string
	var reviewMode, ocrResult, visionResult string
	var aiConfidence float64
	var dateOfBirth, submittedAt, approvedAt, rejectedAt *time.Time
	var createdAt, updatedAt time.Time

	if err := row.Scan(
		&id, &userID, &fullName, &dateOfBirth, &nationalIDNumber,
		&cccdFrontPath, &cccdBackPath, &status, &submittedAt, &approvedAt, &rejectedAt,
		&reviewer, &rejectReason, &reviewMode, &aiConfidence, &ocrResult, &visionResult,
		&createdAt, &updatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, domainerrors.Internal("scan rider verification").WithMeta("error", err.Error())
	}

	var dob time.Time
	if dateOfBirth != nil {
		dob = dateOfBirth.UTC()
	}

	return entity.ReconstituteRiderVerification(
		id, userID, fullName, dob, nationalIDNumber, cccdFrontPath, cccdBackPath,
		entity.RiderKYCStatus(status), submittedAt, approvedAt, rejectedAt,
		reviewer, rejectReason, entity.ReviewMode(reviewMode), aiConfidence, ocrResult, visionResult,
		createdAt.UTC(), updatedAt.UTC(),
	), nil
}

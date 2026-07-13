package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

const driverVerificationFields = `
	id, driver_id, full_name, date_of_birth, address, national_id_number, license_number,
	status, submitted_at, approved_at, rejected_at, expired_at, reviewer, reject_reason,
	created_at, updated_at`

// DriverVerificationRepository is the PostgreSQL implementation of repository.DriverVerificationRepository.
type DriverVerificationRepository struct {
	pool *pgxpool.Pool
}

var _ repository.DriverVerificationRepository = (*DriverVerificationRepository)(nil)

func NewDriverVerificationRepository(pool *pgxpool.Pool) *DriverVerificationRepository {
	return &DriverVerificationRepository{pool: pool}
}

func (r *DriverVerificationRepository) Save(ctx context.Context, v *entity.DriverVerification) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO driver_verifications (`+driverVerificationFields+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		ON CONFLICT (driver_id) DO UPDATE SET
			full_name          = EXCLUDED.full_name,
			date_of_birth      = EXCLUDED.date_of_birth,
			address            = EXCLUDED.address,
			national_id_number = EXCLUDED.national_id_number,
			license_number     = EXCLUDED.license_number,
			status             = EXCLUDED.status,
			submitted_at       = EXCLUDED.submitted_at,
			approved_at        = EXCLUDED.approved_at,
			rejected_at        = EXCLUDED.rejected_at,
			expired_at         = EXCLUDED.expired_at,
			reviewer           = EXCLUDED.reviewer,
			reject_reason      = EXCLUDED.reject_reason,
			updated_at         = EXCLUDED.updated_at`,
		v.ID, v.DriverID, v.FullName, v.DateOfBirth.UTC(), v.Address, v.NationalIDNumber, v.LicenseNumber,
		string(v.Status), v.SubmittedAt.UTC(), v.ApprovedAt, v.RejectedAt, v.ExpiredAt, v.Reviewer, v.RejectReason,
		v.CreatedAt.UTC(), v.UpdatedAt.UTC(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "driver_verifications_national_id_idx":
				return domainerrors.AlreadyExists("national ID number is already registered to another driver")
			case "driver_verifications_license_number_idx":
				return domainerrors.AlreadyExists("license number is already registered to another driver")
			}
			return domainerrors.AlreadyExists("driver verification already exists")
		}
		return domainerrors.Internal("driver_verification: save failed").WithMeta("error", err.Error())
	}
	return nil
}

func (r *DriverVerificationRepository) FindByDriverID(ctx context.Context, driverID string) (*entity.DriverVerification, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+driverVerificationFields+` FROM driver_verifications WHERE driver_id = $1`, driverID)
	return scanDriverVerification(row)
}

func (r *DriverVerificationRepository) FindByNationalIDNumber(ctx context.Context, nationalIDNumber string) (*entity.DriverVerification, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+driverVerificationFields+` FROM driver_verifications WHERE national_id_number = $1`, nationalIDNumber)
	return scanDriverVerification(row)
}

func (r *DriverVerificationRepository) FindByLicenseNumber(ctx context.Context, licenseNumber string) (*entity.DriverVerification, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+driverVerificationFields+` FROM driver_verifications WHERE license_number = $1`, licenseNumber)
	return scanDriverVerification(row)
}

func (r *DriverVerificationRepository) ListByStatus(ctx context.Context, status entity.KYCStatus, limit int) ([]*entity.DriverVerification, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		SELECT `+driverVerificationFields+` FROM driver_verifications
		WHERE status = $1
		ORDER BY submitted_at DESC
		LIMIT $2`, string(status), limit)
	if err != nil {
		return nil, domainerrors.Internal("driver_verification: list failed").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var out []*entity.DriverVerification
	for rows.Next() {
		v, err := scanDriverVerification(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("driver_verification: rows error").WithMeta("error", err.Error())
	}
	return out, nil
}

func (r *DriverVerificationRepository) CountByStatus(ctx context.Context, status entity.KYCStatus) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM driver_verifications WHERE status = $1`, string(status)).Scan(&count)
	if err != nil {
		return 0, domainerrors.Internal("driver_verification: count failed").WithMeta("error", err.Error())
	}
	return count, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDriverVerification(row rowScanner) (*entity.DriverVerification, error) {
	var (
		id, driverID, fullName, address, nationalIDNumber, licenseNumber string
		dateOfBirth                                                      time.Time
		status                                                           string
		submittedAt                                                      time.Time
		approvedAt, rejectedAt, expiredAt                                *time.Time
		reviewer, rejectReason                                           string
		createdAt, updatedAt                                             time.Time
	)
	err := row.Scan(
		&id, &driverID, &fullName, &dateOfBirth, &address, &nationalIDNumber, &licenseNumber,
		&status, &submittedAt, &approvedAt, &rejectedAt, &expiredAt, &reviewer, &rejectReason,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("driver verification not found")
		}
		return nil, domainerrors.Internal("driver_verification: scan failed").WithMeta("error", err.Error())
	}
	return entity.ReconstituteDriverVerification(
		id, driverID, fullName, dateOfBirth.UTC(), address, nationalIDNumber, licenseNumber,
		entity.KYCStatus(status), submittedAt.UTC(), approvedAt, rejectedAt, expiredAt,
		reviewer, rejectReason, createdAt.UTC(), updatedAt.UTC(),
	), nil
}

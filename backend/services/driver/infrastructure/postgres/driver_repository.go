package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/fairride/driver/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const driverFields = `
	driver_id, user_id, license_number, vehicle_type,
	vehicle_brand, vehicle_model, vehicle_color, plate_number,
	online_status, verification_status, created_at, updated_at`

const driverTable = `driver_profiles`

const driverSelect = `SELECT ` + driverFields + ` FROM ` + driverTable

// DriverRepository is the PostgreSQL implementation of repository.DriverRepository.
type DriverRepository struct {
	pool *pgxpool.Pool
}

func NewDriverRepository(pool *pgxpool.Pool) *DriverRepository {
	return &DriverRepository{pool: pool}
}

// FindByID returns a driver profile by its primary key.
func (r *DriverRepository) FindByID(ctx context.Context, driverID string) (*entity.DriverProfile, error) {
	return r.queryOne(ctx, driverSelect+` WHERE driver_id = $1`, driverID)
}

// FindByUserID returns the driver profile associated with the given user identity.
func (r *DriverRepository) FindByUserID(ctx context.Context, userID string) (*entity.DriverProfile, error) {
	return r.queryOne(ctx, driverSelect+` WHERE user_id = $1`, userID)
}

// Save upserts a driver profile. user_id is immutable after the first insert.
func (r *DriverRepository) Save(ctx context.Context, d *entity.DriverProfile) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO driver_profiles (
			driver_id, user_id, license_number, vehicle_type,
			vehicle_brand, vehicle_model, vehicle_color, plate_number,
			online_status, verification_status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (driver_id) DO UPDATE SET
			license_number      = EXCLUDED.license_number,
			vehicle_type        = EXCLUDED.vehicle_type,
			vehicle_brand       = EXCLUDED.vehicle_brand,
			vehicle_model       = EXCLUDED.vehicle_model,
			vehicle_color       = EXCLUDED.vehicle_color,
			plate_number        = EXCLUDED.plate_number,
			online_status       = EXCLUDED.online_status,
			verification_status = EXCLUDED.verification_status,
			updated_at          = EXCLUDED.updated_at`,
		d.DriverID,
		d.UserID,
		d.LicenseNumber,
		string(d.VehicleType),
		d.VehicleBrand,
		d.VehicleModel,
		d.VehicleColor,
		d.PlateNumber,
		string(d.OnlineStatus),
		string(d.VerificationStatus),
		d.CreatedAt.UTC(),
		d.UpdatedAt.UTC(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.AlreadyExists("driver profile already exists")
		}
		return domainerrors.Internal("save driver profile").WithMeta("error", err.Error())
	}
	return nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func (r *DriverRepository) queryOne(ctx context.Context, sql string, args ...any) (*entity.DriverProfile, error) {
	var (
		driverID, userID, licenseNumber string
		vehicleType                     string
		vehicleBrand, vehicleModel      string
		vehicleColor, plateNumber       string
		onlineStatus, verificationStatus string
		createdAt, updatedAt            time.Time
	)

	err := r.pool.QueryRow(ctx, sql, args...).Scan(
		&driverID, &userID, &licenseNumber, &vehicleType,
		&vehicleBrand, &vehicleModel, &vehicleColor, &plateNumber,
		&onlineStatus, &verificationStatus,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("driver profile not found")
		}
		return nil, domainerrors.Internal("query driver profile").WithMeta("error", err.Error())
	}

	return entity.ReconstituteDriverProfile(
		driverID, userID, licenseNumber,
		entity.VehicleType(vehicleType),
		vehicleBrand, vehicleModel, vehicleColor, plateNumber,
		entity.OnlineStatus(onlineStatus),
		entity.VerificationStatus(verificationStatus),
		createdAt.UTC(), updatedAt.UTC(),
	), nil
}

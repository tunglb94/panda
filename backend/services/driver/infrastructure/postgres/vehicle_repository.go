package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/fairride/driver/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const vehicleFields = `
	vehicle_id, driver_id, type, brand, model, color, plate_number, year, created_at, updated_at`

const vehicleTable = `vehicles`

const vehicleSelect = `SELECT ` + vehicleFields + ` FROM ` + vehicleTable

// VehicleRepository is the PostgreSQL implementation of repository.VehicleRepository.
type VehicleRepository struct {
	pool *pgxpool.Pool
}

func NewVehicleRepository(pool *pgxpool.Pool) *VehicleRepository {
	return &VehicleRepository{pool: pool}
}

// FindByID returns a vehicle by its primary key.
func (r *VehicleRepository) FindByID(ctx context.Context, vehicleID string) (*entity.Vehicle, error) {
	return r.queryOne(ctx, vehicleSelect+` WHERE vehicle_id = $1`, vehicleID)
}

// FindByDriverID returns all vehicles belonging to a driver, ordered by created_at.
func (r *VehicleRepository) FindByDriverID(ctx context.Context, driverID string) ([]*entity.Vehicle, error) {
	rows, err := r.pool.Query(ctx, vehicleSelect+` WHERE driver_id = $1 ORDER BY created_at ASC`, driverID)
	if err != nil {
		return nil, domainerrors.Internal("query vehicles").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var vehicles []*entity.Vehicle
	for rows.Next() {
		v, err := scanVehicle(rows)
		if err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("scan vehicles").WithMeta("error", err.Error())
	}
	return vehicles, nil
}

// Save upserts a vehicle. created_at is immutable after first insert.
func (r *VehicleRepository) Save(ctx context.Context, v *entity.Vehicle) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO vehicles (
			vehicle_id, driver_id, type, brand, model, color, plate_number, year, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (vehicle_id) DO UPDATE SET
			type         = EXCLUDED.type,
			brand        = EXCLUDED.brand,
			model        = EXCLUDED.model,
			color        = EXCLUDED.color,
			plate_number = EXCLUDED.plate_number,
			year         = EXCLUDED.year,
			updated_at   = EXCLUDED.updated_at`,
		v.VehicleID,
		v.DriverID,
		string(v.Type),
		v.Brand,
		v.Model,
		v.Color,
		v.PlateNumber,
		v.Year,
		v.CreatedAt.UTC(),
		v.UpdatedAt.UTC(),
	)
	if err != nil {
		return domainerrors.Internal("save vehicle").WithMeta("error", err.Error())
	}
	return nil
}

// Delete permanently removes a vehicle by its primary key.
// Returns CodeNotFound if no vehicle exists with that ID.
func (r *VehicleRepository) Delete(ctx context.Context, vehicleID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM vehicles WHERE vehicle_id = $1`, vehicleID)
	if err != nil {
		return domainerrors.Internal("delete vehicle").WithMeta("error", err.Error())
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.NotFound("vehicle not found")
	}
	return nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func (r *VehicleRepository) queryOne(ctx context.Context, sql string, args ...any) (*entity.Vehicle, error) {
	var (
		vehicleID, driverID string
		vtype               string
		brand, model, color string
		plateNumber         string
		year                int
		createdAt, updatedAt time.Time
	)
	err := r.pool.QueryRow(ctx, sql, args...).Scan(
		&vehicleID, &driverID, &vtype,
		&brand, &model, &color, &plateNumber,
		&year,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("vehicle not found")
		}
		return nil, domainerrors.Internal("query vehicle").WithMeta("error", err.Error())
	}
	return entity.ReconstituteVehicle(
		vehicleID, driverID,
		entity.VehicleType(vtype),
		brand, model, color, plateNumber,
		year,
		createdAt.UTC(), updatedAt.UTC(),
	), nil
}

// scanVehicle reads one row from an open pgx.Rows into a Vehicle.
type pgxRows interface {
	Scan(dest ...any) error
}

func scanVehicle(row pgxRows) (*entity.Vehicle, error) {
	var (
		vehicleID, driverID string
		vtype               string
		brand, model, color string
		plateNumber         string
		year                int
		createdAt, updatedAt time.Time
	)
	if err := row.Scan(
		&vehicleID, &driverID, &vtype,
		&brand, &model, &color, &plateNumber,
		&year,
		&createdAt, &updatedAt,
	); err != nil {
		return nil, domainerrors.Internal("scan vehicle row").WithMeta("error", err.Error())
	}
	return entity.ReconstituteVehicle(
		vehicleID, driverID,
		entity.VehicleType(vtype),
		brand, model, color, plateNumber,
		year,
		createdAt.UTC(), updatedAt.UTC(),
	), nil
}

package postgres

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

const vehicleVerificationFields = `
	id, driver_id, vehicle_type, service_type, brand, model, year, color,
	plate_number, vin, engine_number, chassis_number, license_class, ride_enabled, delivery_enabled,
	status, submitted_at, approved_at, rejected_at, expired_at, reviewer, reject_reason,
	created_at, updated_at`

// VehicleVerificationRepository is the PostgreSQL implementation of repository.VehicleVerificationRepository.
type VehicleVerificationRepository struct {
	pool *pgxpool.Pool
}

var _ repository.VehicleVerificationRepository = (*VehicleVerificationRepository)(nil)

func NewVehicleVerificationRepository(pool *pgxpool.Pool) *VehicleVerificationRepository {
	return &VehicleVerificationRepository{pool: pool}
}

func (r *VehicleVerificationRepository) Save(ctx context.Context, v *entity.VehicleVerification) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO vehicle_verifications (`+vehicleVerificationFields+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)
		ON CONFLICT (driver_id) DO UPDATE SET
			vehicle_type     = EXCLUDED.vehicle_type,
			service_type     = EXCLUDED.service_type,
			brand            = EXCLUDED.brand,
			model            = EXCLUDED.model,
			year             = EXCLUDED.year,
			color            = EXCLUDED.color,
			plate_number     = EXCLUDED.plate_number,
			vin              = EXCLUDED.vin,
			engine_number    = EXCLUDED.engine_number,
			chassis_number   = EXCLUDED.chassis_number,
			license_class    = EXCLUDED.license_class,
			ride_enabled     = EXCLUDED.ride_enabled,
			delivery_enabled = EXCLUDED.delivery_enabled,
			status           = EXCLUDED.status,
			submitted_at     = EXCLUDED.submitted_at,
			approved_at      = EXCLUDED.approved_at,
			rejected_at      = EXCLUDED.rejected_at,
			expired_at       = EXCLUDED.expired_at,
			reviewer         = EXCLUDED.reviewer,
			reject_reason    = EXCLUDED.reject_reason,
			updated_at       = EXCLUDED.updated_at`,
		v.ID, v.DriverID, string(v.VehicleType), string(v.ServiceType), v.Brand, v.Model, v.Year, v.Color,
		v.PlateNumber, v.VIN, v.EngineNumber, v.ChassisNumber, string(v.LicenseClass), v.RideEnabled, v.DeliveryEnabled,
		string(v.Status), v.SubmittedAt.UTC(), v.ApprovedAt, v.RejectedAt, v.ExpiredAt, v.Reviewer, v.RejectReason,
		v.CreatedAt.UTC(), v.UpdatedAt.UTC(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "vehicle_verifications_plate_idx":
				return domainerrors.AlreadyExists("plate number is already registered to another verification")
			case "vehicle_verifications_vin_idx":
				return domainerrors.AlreadyExists("VIN is already registered to another vehicle")
			case "vehicle_verifications_engine_number_idx":
				return domainerrors.AlreadyExists("engine number is already registered to another vehicle")
			case "vehicle_verifications_chassis_number_idx":
				return domainerrors.AlreadyExists("chassis number is already registered to another vehicle")
			}
			return domainerrors.AlreadyExists("vehicle verification already exists")
		}
		return domainerrors.Internal("vehicle_verification: save failed").WithMeta("error", err.Error())
	}
	return nil
}

func (r *VehicleVerificationRepository) FindByDriverID(ctx context.Context, driverID string) (*entity.VehicleVerification, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+vehicleVerificationFields+` FROM vehicle_verifications WHERE driver_id = $1`, driverID)
	return scanVehicleVerification(row)
}

func (r *VehicleVerificationRepository) FindByPlateNumber(ctx context.Context, plateNumber string) (*entity.VehicleVerification, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+vehicleVerificationFields+` FROM vehicle_verifications WHERE plate_number = $1`, plateNumber)
	return scanVehicleVerification(row)
}

func (r *VehicleVerificationRepository) FindByVIN(ctx context.Context, vin string) (*entity.VehicleVerification, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+vehicleVerificationFields+` FROM vehicle_verifications WHERE vin = $1`, vin)
	return scanVehicleVerification(row)
}

func (r *VehicleVerificationRepository) FindByEngineNumber(ctx context.Context, engineNumber string) (*entity.VehicleVerification, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+vehicleVerificationFields+` FROM vehicle_verifications WHERE engine_number = $1`, engineNumber)
	return scanVehicleVerification(row)
}

func (r *VehicleVerificationRepository) FindByChassisNumber(ctx context.Context, chassisNumber string) (*entity.VehicleVerification, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+vehicleVerificationFields+` FROM vehicle_verifications WHERE chassis_number = $1`, chassisNumber)
	return scanVehicleVerification(row)
}

func (r *VehicleVerificationRepository) ListByFilter(ctx context.Context, status entity.KYCStatus, vehicleType entity.VehicleType, serviceType entity.ServiceType, limit int) ([]*entity.VehicleVerification, error) {
	q, args := buildVehicleFilterQuery(vehicleVerificationFields, status, vehicleType, serviceType, limit)
	q += ` ORDER BY submitted_at DESC LIMIT $` + strconv.Itoa(len(args)+1)
	args = append(args, limit)
	return r.queryVehicleVerifications(ctx, q, args)
}

// ListByFilterSortedByExpiry orders by the nearest upcoming expiry among
// each vehicle's expiry-tracked documents (license/registration/insurance/
// inspection), ascending — vehicles with no expiry-tracked document
// uploaded yet sort last (Phần 12).
func (r *VehicleVerificationRepository) ListByFilterSortedByExpiry(ctx context.Context, status entity.KYCStatus, vehicleType entity.VehicleType, serviceType entity.ServiceType, limit int) ([]*entity.VehicleVerification, error) {
	q := `SELECT ` + vehicleVerificationFields + `
		FROM vehicle_verifications vv
		LEFT JOIN LATERAL (
			SELECT MIN(d.expires_at) AS next_expiry
			FROM kyc_documents d
			WHERE d.driver_id = vv.driver_id
			  AND d.document_type IN ('license','vehicle_registration','vehicle_insurance','vehicle_inspection')
			  AND d.expires_at IS NOT NULL
		) exp ON true
		WHERE 1=1`
	args := []any{}
	if status != "" {
		args = append(args, string(status))
		q += ` AND vv.status = $` + strconv.Itoa(len(args))
	}
	if vehicleType != "" {
		args = append(args, string(vehicleType))
		q += ` AND vv.vehicle_type = $` + strconv.Itoa(len(args))
	}
	if serviceType != "" {
		args = append(args, string(serviceType))
		q += ` AND vv.service_type = $` + strconv.Itoa(len(args))
	}
	q += ` ORDER BY exp.next_expiry ASC NULLS LAST LIMIT $` + strconv.Itoa(len(args)+1)
	args = append(args, limit)
	return r.queryVehicleVerifications(ctx, q, args)
}

func buildVehicleFilterQuery(fields string, status entity.KYCStatus, vehicleType entity.VehicleType, serviceType entity.ServiceType, limit int) (string, []any) {
	q := `SELECT ` + fields + ` FROM vehicle_verifications WHERE 1=1`
	args := []any{}
	if status != "" {
		args = append(args, string(status))
		q += ` AND status = $` + strconv.Itoa(len(args))
	}
	if vehicleType != "" {
		args = append(args, string(vehicleType))
		q += ` AND vehicle_type = $` + strconv.Itoa(len(args))
	}
	if serviceType != "" {
		args = append(args, string(serviceType))
		q += ` AND service_type = $` + strconv.Itoa(len(args))
	}
	return q, args
}

func (r *VehicleVerificationRepository) queryVehicleVerifications(ctx context.Context, q string, args []any) ([]*entity.VehicleVerification, error) {
	if limit, ok := args[len(args)-1].(int); ok && (limit <= 0 || limit > 200) {
		args[len(args)-1] = 50
	}
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, domainerrors.Internal("vehicle_verification: list failed").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var out []*entity.VehicleVerification
	for rows.Next() {
		v, err := scanVehicleVerification(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("vehicle_verification: rows error").WithMeta("error", err.Error())
	}
	return out, nil
}

func scanVehicleVerification(row rowScanner) (*entity.VehicleVerification, error) {
	var (
		id, driverID, vehicleType, serviceType               string
		brand, model                                         string
		year                                                 int
		color, plateNumber, vin, engineNumber, chassisNumber string
		licenseClass                                         string
		rideEnabled, deliveryEnabled                         bool
		status                                               string
		submittedAt                                          time.Time
		approvedAt, rejectedAt, expiredAt                    *time.Time
		reviewer, rejectReason                               string
		createdAt, updatedAt                                 time.Time
	)
	err := row.Scan(
		&id, &driverID, &vehicleType, &serviceType, &brand, &model, &year, &color,
		&plateNumber, &vin, &engineNumber, &chassisNumber, &licenseClass, &rideEnabled, &deliveryEnabled,
		&status, &submittedAt, &approvedAt, &rejectedAt, &expiredAt, &reviewer, &rejectReason,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("vehicle verification not found")
		}
		return nil, domainerrors.Internal("vehicle_verification: scan failed").WithMeta("error", err.Error())
	}
	return entity.ReconstituteVehicleVerification(
		id, driverID, entity.VehicleType(vehicleType), entity.ServiceType(serviceType),
		brand, model, year, color, plateNumber, vin, engineNumber, chassisNumber, entity.LicenseClass(licenseClass),
		rideEnabled, deliveryEnabled, entity.KYCStatus(status), submittedAt.UTC(),
		approvedAt, rejectedAt, expiredAt, reviewer, rejectReason, createdAt.UTC(), updatedAt.UTC(),
	), nil
}

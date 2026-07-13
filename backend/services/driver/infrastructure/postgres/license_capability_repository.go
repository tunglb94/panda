package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// LicenseCapabilityRepository is the PostgreSQL-backed Rule Engine behind
// Phần 1 — reads the license_capabilities table instead of a hardcoded Go
// map, so a change in Vietnamese law is a data UPDATE, not a deploy.
type LicenseCapabilityRepository struct {
	pool *pgxpool.Pool
}

var _ repository.LicenseCapabilityRepository = (*LicenseCapabilityRepository)(nil)

func NewLicenseCapabilityRepository(pool *pgxpool.Pool) *LicenseCapabilityRepository {
	return &LicenseCapabilityRepository{pool: pool}
}

func (r *LicenseCapabilityRepository) IsAllowed(ctx context.Context, licenseClass entity.LicenseClass, serviceType entity.ServiceType) (bool, error) {
	var allowed bool
	err := r.pool.QueryRow(ctx, `
		SELECT allowed FROM license_capabilities
		WHERE license_class = $1 AND service_type = $2`,
		string(licenseClass), string(serviceType),
	).Scan(&allowed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil // unknown pair — deny by default
		}
		return false, domainerrors.Internal("license_capability: query failed").WithMeta("error", err.Error())
	}
	return allowed, nil
}

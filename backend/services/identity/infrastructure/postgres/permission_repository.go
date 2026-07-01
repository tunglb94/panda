// Package postgres contains PostgreSQL implementations of the Identity domain repositories.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/fairride/identity/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	permFields = `id, name, resource, action, description, created_at`
	permTable  = `identity_permissions`
	permSelect = `SELECT ` + permFields + ` FROM ` + permTable
)

// PermissionRepository is the PostgreSQL implementation of repository.PermissionRepository.
type PermissionRepository struct {
	pool *pgxpool.Pool
}

// NewPermissionRepository constructs a PermissionRepository backed by pool.
func NewPermissionRepository(pool *pgxpool.Pool) *PermissionRepository {
	return &PermissionRepository{pool: pool}
}

// FindByID returns the permission with the given ID.
// Returns errors.CodeNotFound if no row exists.
func (r *PermissionRepository) FindByID(ctx context.Context, id string) (*entity.Permission, error) {
	return r.queryOne(ctx, permSelect+` WHERE id = $1`, id)
}

// FindByName returns the permission with the given canonical name.
// Returns errors.CodeNotFound if no row exists.
func (r *PermissionRepository) FindByName(ctx context.Context, name string) (*entity.Permission, error) {
	return r.queryOne(ctx, permSelect+` WHERE name = $1`, name)
}

// FindByResource returns all permissions whose resource component matches.
// Returns an empty slice (not an error) when none are found.
func (r *PermissionRepository) FindByResource(ctx context.Context, resource string) ([]*entity.Permission, error) {
	return r.queryMany(ctx, permSelect+` WHERE resource = $1 ORDER BY name`, resource)
}

// FindAll returns every permission ordered by name.
func (r *PermissionRepository) FindAll(ctx context.Context) ([]*entity.Permission, error) {
	return r.queryMany(ctx, permSelect+` ORDER BY name`)
}

// Save upserts a permission. Existing rows are matched by ID; on conflict the mutable
// fields (name, resource, action, description) are overwritten. created_at is immutable.
func (r *PermissionRepository) Save(ctx context.Context, p *entity.Permission) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO identity_permissions (id, name, resource, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			name        = EXCLUDED.name,
			resource    = EXCLUDED.resource,
			action      = EXCLUDED.action,
			description = EXCLUDED.description
	`, p.ID, p.Name, p.Resource, p.Action, p.Description, p.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.AlreadyExists("permission name already exists: " + p.Name)
		}
		return domainerrors.Internal("save permission").WithMeta("error", err.Error())
	}
	return nil
}

// Delete removes the permission with the given ID.
// Returns errors.CodeNotFound if no row with that ID exists.
func (r *PermissionRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM `+permTable+` WHERE id = $1`, id)
	if err != nil {
		return domainerrors.Internal("delete permission").WithMeta("error", err.Error())
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.NotFound("permission not found: " + id)
	}
	return nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func (r *PermissionRepository) queryOne(ctx context.Context, sql string, args ...any) (*entity.Permission, error) {
	var id, name, resource, action, description string
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, sql, args...).
		Scan(&id, &name, &resource, &action, &description, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("permission not found")
		}
		return nil, domainerrors.Internal("query permission").WithMeta("error", err.Error())
	}
	return entity.ReconstitutePermission(id, name, resource, action, description, createdAt), nil
}

func (r *PermissionRepository) queryMany(ctx context.Context, sql string, args ...any) ([]*entity.Permission, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, domainerrors.Internal("query permissions").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var result []*entity.Permission
	for rows.Next() {
		var id, name, resource, action, description string
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &resource, &action, &description, &createdAt); err != nil {
			return nil, domainerrors.Internal("scan permission").WithMeta("error", err.Error())
		}
		result = append(result, entity.ReconstitutePermission(id, name, resource, action, description, createdAt))
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("iterate permissions").WithMeta("error", err.Error())
	}
	return result, nil
}

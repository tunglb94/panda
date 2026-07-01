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

// roleSelectJoin is the base SELECT that LEFT JOINs roles to their permissions.
// A role with no permissions produces one row with NULL permission columns.
const roleSelectJoin = `
	SELECT r.id, r.name, r.description, r.is_system, r.created_at, r.updated_at,
	       p.id, p.name, p.resource, p.action, p.description, p.created_at
	FROM identity_roles r
	LEFT JOIN identity_role_permissions rp ON rp.role_id = r.id
	LEFT JOIN identity_permissions p      ON p.id = rp.permission_id`

// RoleRepository is the PostgreSQL implementation of repository.RoleRepository.
type RoleRepository struct {
	pool *pgxpool.Pool
}

// NewRoleRepository constructs a RoleRepository backed by pool.
func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

// FindByID returns the role with the given ID, including its full permission set.
// Returns errors.CodeNotFound if no row exists.
func (r *RoleRepository) FindByID(ctx context.Context, id string) (*entity.Role, error) {
	rows, err := r.pool.Query(ctx,
		roleSelectJoin+` WHERE r.id = $1 ORDER BY p.name NULLS LAST`, id)
	if err != nil {
		return nil, domainerrors.Internal("query role by id").WithMeta("error", err.Error())
	}
	roles, err := r.scanRoles(rows)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, domainerrors.NotFound("role not found: " + id)
	}
	return roles[0], nil
}

// FindByName returns the role with the given name, including its full permission set.
// Returns errors.CodeNotFound if no row exists.
func (r *RoleRepository) FindByName(ctx context.Context, name string) (*entity.Role, error) {
	rows, err := r.pool.Query(ctx,
		roleSelectJoin+` WHERE r.name = $1 ORDER BY p.name NULLS LAST`, name)
	if err != nil {
		return nil, domainerrors.Internal("query role by name").WithMeta("error", err.Error())
	}
	roles, err := r.scanRoles(rows)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, domainerrors.NotFound("role not found: " + name)
	}
	return roles[0], nil
}

// FindAll returns every role with its full permission set, ordered by role name.
func (r *RoleRepository) FindAll(ctx context.Context) ([]*entity.Role, error) {
	rows, err := r.pool.Query(ctx,
		roleSelectJoin+` ORDER BY r.name, p.name NULLS LAST`)
	if err != nil {
		return nil, domainerrors.Internal("query all roles").WithMeta("error", err.Error())
	}
	return r.scanRoles(rows)
}

// Save upserts a role and atomically replaces its full permission set.
// Existing roles are matched by ID; on conflict the mutable fields are overwritten.
// created_at is immutable. Permissions not in role.Permissions() are removed.
// The caller MUST ensure all permissions in role.Permissions() are already persisted.
func (r *RoleRepository) Save(ctx context.Context, role *entity.Role) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domainerrors.Internal("begin save role tx").WithMeta("error", err.Error())
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `
		INSERT INTO identity_roles (id, name, description, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			name        = EXCLUDED.name,
			description = EXCLUDED.description,
			is_system   = EXCLUDED.is_system,
			updated_at  = EXCLUDED.updated_at
	`, role.ID, role.Name, role.Description, role.IsSystem, role.CreatedAt, now)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.AlreadyExists("role name already exists: " + role.Name)
		}
		return domainerrors.Internal("upsert role").WithMeta("error", err.Error())
	}

	// Full replacement of the permission set: delete then re-insert.
	_, err = tx.Exec(ctx, `DELETE FROM identity_role_permissions WHERE role_id = $1`, role.ID)
	if err != nil {
		return domainerrors.Internal("clear role permissions").WithMeta("error", err.Error())
	}

	for _, p := range role.Permissions() {
		_, err = tx.Exec(ctx, `
			INSERT INTO identity_role_permissions (role_id, permission_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, role.ID, p.ID)
		if err != nil {
			return domainerrors.Internal("link role permission").
				WithMeta("role_id", role.ID).
				WithMeta("permission_id", p.ID).
				WithMeta("error", err.Error())
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domainerrors.Internal("commit save role").WithMeta("error", err.Error())
	}
	return nil
}

// Delete removes the role with the given ID.
// Callers MUST call role.CanDelete() before invoking this method.
// Returns errors.CodeNotFound if no row with that ID exists.
func (r *RoleRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM identity_roles WHERE id = $1`, id)
	if err != nil {
		return domainerrors.Internal("delete role").WithMeta("error", err.Error())
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.NotFound("role not found: " + id)
	}
	return nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

// roleAcc accumulates permission rows for a single role while scanning a multi-row result.
type roleAcc struct {
	id, name, description string
	isSystem              bool
	createdAt, updatedAt  time.Time
	perms                 []entity.Permission
}

// scanRoles consumes rows from roleSelectJoin and returns assembled Role entities.
// It closes rows before returning.
func (r *RoleRepository) scanRoles(rows pgx.Rows) ([]*entity.Role, error) {
	defer rows.Close()

	byID := make(map[string]*roleAcc)
	order := make([]string, 0) // preserves DB-returned ordering

	for rows.Next() {
		var roleID, roleName, roleDesc string
		var roleIsSystem bool
		var roleCA, roleUA time.Time
		// Permission columns are nullable (LEFT JOIN).
		var permID, permName, permResource, permAction, permDesc *string
		var permCA *time.Time

		if err := rows.Scan(
			&roleID, &roleName, &roleDesc, &roleIsSystem, &roleCA, &roleUA,
			&permID, &permName, &permResource, &permAction, &permDesc, &permCA,
		); err != nil {
			return nil, domainerrors.Internal("scan role row").WithMeta("error", err.Error())
		}

		acc, exists := byID[roleID]
		if !exists {
			acc = &roleAcc{
				id: roleID, name: roleName, description: roleDesc,
				isSystem: roleIsSystem, createdAt: roleCA, updatedAt: roleUA,
			}
			byID[roleID] = acc
			order = append(order, roleID)
		}

		if permID != nil {
			acc.perms = append(acc.perms, *entity.ReconstitutePermission(
				*permID, *permName, *permResource, *permAction, *permDesc, *permCA,
			))
		}
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("iterate role rows").WithMeta("error", err.Error())
	}

	result := make([]*entity.Role, 0, len(order))
	for _, id := range order {
		acc := byID[id]
		result = append(result, entity.ReconstituteRole(
			acc.id, acc.name, acc.description, acc.isSystem,
			acc.perms, acc.createdAt, acc.updatedAt,
		))
	}
	return result, nil
}

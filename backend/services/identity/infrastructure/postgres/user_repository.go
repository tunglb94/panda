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
	userFields = `id, phone_number, name, email, type, status, role_id, created_at, updated_at`
	userTable  = `identity_users`
	userSelect = `SELECT ` + userFields + ` FROM ` + userTable
)

// UserRepository is the PostgreSQL implementation of repository.UserRepository.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository constructs a UserRepository backed by pool.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// FindByID returns the user with the given ID.
// Returns CodeNotFound if no row exists.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	return r.queryOne(ctx, userSelect+` WHERE id = $1`, id)
}

// FindByPhone returns the user with the given phone number.
// Returns CodeNotFound if no row exists.
func (r *UserRepository) FindByPhone(ctx context.Context, phoneNumber string) (*entity.User, error) {
	return r.queryOne(ctx, userSelect+` WHERE phone_number = $1`, phoneNumber)
}

// FindAll returns every user ordered by created_at ascending.
func (r *UserRepository) FindAll(ctx context.Context) ([]*entity.User, error) {
	return r.queryMany(ctx, userSelect+` ORDER BY created_at`)
}

// Save upserts a user. Rows are matched by ID; on conflict the mutable fields are
// overwritten. created_at is immutable. updated_at is taken from user.UpdatedAt,
// which the domain entity updates when status transitions occur.
// Returns CodeAlreadyExists when a different user already holds the same phone number.
func (r *UserRepository) Save(ctx context.Context, user *entity.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO identity_users
			(id, phone_number, name, email, type, status, role_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			phone_number = EXCLUDED.phone_number,
			name         = EXCLUDED.name,
			email        = EXCLUDED.email,
			type         = EXCLUDED.type,
			status       = EXCLUDED.status,
			role_id      = EXCLUDED.role_id,
			updated_at   = EXCLUDED.updated_at
	`, user.ID, user.PhoneNumber, user.Name, user.Email,
		string(user.Type), string(user.Status), user.RoleID,
		user.CreatedAt, user.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.AlreadyExists("user phone number already exists: " + user.PhoneNumber)
		}
		return domainerrors.Internal("save user").WithMeta("error", err.Error())
	}
	return nil
}

// Delete removes the user with the given ID.
// Returns CodeNotFound if no row with that ID exists.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM `+userTable+` WHERE id = $1`, id)
	if err != nil {
		return domainerrors.Internal("delete user").WithMeta("error", err.Error())
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.NotFound("user not found: " + id)
	}
	return nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func (r *UserRepository) queryOne(ctx context.Context, sql string, args ...any) (*entity.User, error) {
	var id, phoneNumber, name, email, userType, status, roleID string
	var createdAt, updatedAt time.Time
	err := r.pool.QueryRow(ctx, sql, args...).
		Scan(&id, &phoneNumber, &name, &email, &userType, &status, &roleID, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("user not found")
		}
		return nil, domainerrors.Internal("query user").WithMeta("error", err.Error())
	}
	return entity.ReconstituteUser(
		id, phoneNumber, name, email,
		entity.UserType(userType), entity.UserStatus(status),
		roleID, createdAt, updatedAt,
	), nil
}

func (r *UserRepository) queryMany(ctx context.Context, sql string, args ...any) ([]*entity.User, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, domainerrors.Internal("query users").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var result []*entity.User
	for rows.Next() {
		var id, phoneNumber, name, email, userType, status, roleID string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&id, &phoneNumber, &name, &email, &userType, &status, &roleID, &createdAt, &updatedAt,
		); err != nil {
			return nil, domainerrors.Internal("scan user").WithMeta("error", err.Error())
		}
		result = append(result, entity.ReconstituteUser(
			id, phoneNumber, name, email,
			entity.UserType(userType), entity.UserStatus(status),
			roleID, createdAt, updatedAt,
		))
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("iterate users").WithMeta("error", err.Error())
	}
	return result, nil
}

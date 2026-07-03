// Package postgres contains the PostgreSQL implementation of the User service repositories.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/fairride/user/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	profileFields = `id, full_name, phone, email, avatar, date_of_birth, gender, status, created_at, updated_at`
	profileTable  = `user_profiles`
	profileSelect = `SELECT ` + profileFields + ` FROM ` + profileTable
)

// ProfileRepository is the PostgreSQL implementation of repository.ProfileRepository.
type ProfileRepository struct {
	pool *pgxpool.Pool
}

// NewProfileRepository constructs a ProfileRepository backed by pool.
func NewProfileRepository(pool *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{pool: pool}
}

// FindByID returns the profile with the given user ID.
// Returns CodeNotFound if no row exists.
func (r *ProfileRepository) FindByID(ctx context.Context, id string) (*entity.UserProfile, error) {
	return r.queryOne(ctx, profileSelect+` WHERE id = $1`, id)
}

// Save upserts a profile. Rows are matched by ID; on conflict the mutable
// fields are overwritten. created_at is immutable after first insert.
func (r *ProfileRepository) Save(ctx context.Context, p *entity.UserProfile) error {
	// date_of_birth is stored as NULL when zero.
	var dob *time.Time
	if !p.DateOfBirth.IsZero() {
		t := p.DateOfBirth.UTC()
		dob = &t
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_profiles
			(id, full_name, phone, email, avatar, date_of_birth, gender, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			full_name    = EXCLUDED.full_name,
			email        = EXCLUDED.email,
			avatar       = EXCLUDED.avatar,
			date_of_birth = EXCLUDED.date_of_birth,
			gender       = EXCLUDED.gender,
			status       = EXCLUDED.status,
			updated_at   = EXCLUDED.updated_at
	`, p.ID, p.FullName, p.Phone, p.Email, p.Avatar,
		dob, string(p.Gender), string(p.Status), p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return domainerrors.Internal("save user profile").WithMeta("error", err.Error())
	}
	return nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func (r *ProfileRepository) queryOne(ctx context.Context, sql string, args ...any) (*entity.UserProfile, error) {
	var id, fullName, phone, email, avatar, gender, status string
	var createdAt, updatedAt time.Time
	var dob *time.Time // nullable in DB

	err := r.pool.QueryRow(ctx, sql, args...).
		Scan(&id, &fullName, &phone, &email, &avatar, &dob, &gender, &status, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("user profile not found")
		}
		return nil, domainerrors.Internal("query user profile").WithMeta("error", err.Error())
	}

	var dobTime time.Time
	if dob != nil {
		dobTime = dob.UTC()
	}

	return entity.ReconstituteUserProfile(
		id, fullName, phone, email, avatar,
		dobTime,
		entity.Gender(gender),
		entity.ProfileStatus(status),
		createdAt.UTC(), updatedAt.UTC(),
	), nil
}

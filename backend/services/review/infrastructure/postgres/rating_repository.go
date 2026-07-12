package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fairride/review/domain/entity"
	"github.com/fairride/review/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// RatingRepository is the PostgreSQL implementation of repository.RatingRepository.
type RatingRepository struct {
	pool *pgxpool.Pool
}

var _ repository.RatingRepository = (*RatingRepository)(nil)

func NewRatingRepository(pool *pgxpool.Pool) *RatingRepository {
	return &RatingRepository{pool: pool}
}

// CreateSchema creates the ratings table if it does not already exist.
// Called once at service startup.
func CreateSchema(ctx context.Context, pool *pgxpool.Pool) error {
	const q = `
		CREATE TABLE IF NOT EXISTS ratings (
			rating_id  TEXT        PRIMARY KEY,
			trip_id    TEXT        NOT NULL,
			rater_id   TEXT        NOT NULL,
			ratee_id   TEXT        NOT NULL,
			role       TEXT        NOT NULL,
			stars      INTEGER     NOT NULL CHECK (stars BETWEEN 1 AND 5),
			comment    TEXT        NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL,
			UNIQUE (trip_id, role)
		)`
	_, err := pool.Exec(ctx, q)
	return err
}

// Save inserts a new rating. Returns CodeAlreadyExists on duplicate (trip_id, role).
func (r *RatingRepository) Save(ctx context.Context, rating *entity.Rating) error {
	const q = `
		INSERT INTO ratings (rating_id, trip_id, rater_id, ratee_id, role, stars, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, q,
		rating.RatingID,
		rating.TripID,
		rating.RaterID,
		rating.RateeID,
		string(rating.Role),
		rating.Stars,
		rating.Comment,
		rating.CreatedAt.UTC(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.AlreadyExists("rating already exists for this trip and role")
		}
		return domainerrors.Internal("rating: save failed").WithMeta("error", err.Error())
	}
	return nil
}

// FindByTripAndRole returns a rating or CodeNotFound.
func (r *RatingRepository) FindByTripAndRole(ctx context.Context, tripID string, role entity.Role) (*entity.Rating, error) {
	const q = `
		SELECT rating_id, trip_id, rater_id, ratee_id, role, stars, comment, created_at
		FROM ratings
		WHERE trip_id = $1 AND role = $2`

	var (
		ratingID, trip, rater, ratee, roleStr, comment string
		stars                                          int32
		createdAt                                      time.Time
	)
	err := r.pool.QueryRow(ctx, q, tripID, string(role)).Scan(
		&ratingID, &trip, &rater, &ratee, &roleStr, &stars, &comment, &createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("rating not found")
		}
		return nil, domainerrors.Internal("rating: find failed").WithMeta("error", err.Error())
	}
	return entity.ReconstituteRating(ratingID, trip, rater, ratee, entity.Role(roleStr), stars, comment, createdAt.UTC()), nil
}

// FindAverageByRatee implements repository.RatingRepository.
func (r *RatingRepository) FindAverageByRatee(ctx context.Context, rateeID string, role entity.Role) (float64, int32, error) {
	const q = `
		SELECT COALESCE(AVG(stars), 0), COUNT(*)
		FROM ratings
		WHERE ratee_id = $1 AND role = $2`
	var (
		avg   float64
		count int32
	)
	if err := r.pool.QueryRow(ctx, q, rateeID, string(role)).Scan(&avg, &count); err != nil {
		return 0, 0, domainerrors.Internal("rating: average failed").WithMeta("error", err.Error())
	}
	return avg, count, nil
}

package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/fairride/promotion/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const voucherFields = `
	id, code, name, description, status, priority,
	start_time, end_time, max_usage, max_usage_per_user,
	budget, remaining_budget, discount_type, discount_value, max_discount, min_order,
	vehicle_types, cities, membership,
	new_user_only, combinable, stackable, promotion_type, usage_count,
	created_at, updated_at`

const voucherTable = `vouchers`

const voucherSelect = `SELECT ` + voucherFields + ` FROM ` + voucherTable

// VoucherRepository is the PostgreSQL implementation of repository.PromotionRepository.
type VoucherRepository struct {
	pool *pgxpool.Pool
}

func NewVoucherRepository(pool *pgxpool.Pool) *VoucherRepository {
	return &VoucherRepository{pool: pool}
}

func (r *VoucherRepository) FindByID(ctx context.Context, id string) (*entity.Voucher, error) {
	return r.queryOne(ctx, voucherSelect+` WHERE id = $1`, id)
}

func (r *VoucherRepository) FindByCode(ctx context.Context, code string) (*entity.Voucher, error) {
	return r.queryOne(ctx, voucherSelect+` WHERE lower(code) = lower($1) AND code <> ''`, code)
}

// FindAutoApplyCandidates returns active, code-less campaigns whose type is
// in types and whose city/vehicle-type restriction (if any) matches. City and
// vehicle-type filtering also accepts campaigns with no restriction set
// (BRB §4.10 nationwide default / §4.11 all-vehicle-classes default).
func (r *VoucherRepository) FindAutoApplyCandidates(ctx context.Context, city, vehicleType string, types []entity.PromotionType) ([]*entity.Voucher, error) {
	typeStrings := make([]string, len(types))
	for i, t := range types {
		typeStrings[i] = string(t)
	}

	rows, err := r.pool.Query(ctx, voucherSelect+`
		WHERE code = ''
		  AND status = 'active'
		  AND promotion_type = ANY($1)
		  AND (cardinality(cities) = 0 OR $2 = ANY(cities))
		  AND (cardinality(vehicle_types) = 0 OR $3 = ANY(vehicle_types))`,
		typeStrings, city, vehicleType,
	)
	if err != nil {
		return nil, domainerrors.Internal("query auto-apply vouchers").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var result []*entity.Voucher
	for rows.Next() {
		v, err := scanVoucher(rows)
		if err != nil {
			return nil, domainerrors.Internal("scan voucher").WithMeta("error", err.Error())
		}
		result = append(result, v)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("iterate vouchers").WithMeta("error", err.Error())
	}
	return result, nil
}

func (r *VoucherRepository) Save(ctx context.Context, v *entity.Voucher) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO vouchers (
			id, code, name, description, status, priority,
			start_time, end_time, max_usage, max_usage_per_user,
			budget, remaining_budget, discount_type, discount_value, max_discount, min_order,
			vehicle_types, cities, membership,
			new_user_only, combinable, stackable, promotion_type, usage_count,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16,
			$17, $18, $19,
			$20, $21, $22, $23, $24,
			$25, $26
		)
		ON CONFLICT (id) DO UPDATE SET
			code                = EXCLUDED.code,
			name                = EXCLUDED.name,
			description         = EXCLUDED.description,
			status              = EXCLUDED.status,
			priority            = EXCLUDED.priority,
			start_time          = EXCLUDED.start_time,
			end_time            = EXCLUDED.end_time,
			max_usage           = EXCLUDED.max_usage,
			max_usage_per_user  = EXCLUDED.max_usage_per_user,
			budget              = EXCLUDED.budget,
			remaining_budget    = EXCLUDED.remaining_budget,
			discount_type       = EXCLUDED.discount_type,
			discount_value      = EXCLUDED.discount_value,
			max_discount        = EXCLUDED.max_discount,
			min_order           = EXCLUDED.min_order,
			vehicle_types       = EXCLUDED.vehicle_types,
			cities              = EXCLUDED.cities,
			membership          = EXCLUDED.membership,
			new_user_only       = EXCLUDED.new_user_only,
			combinable          = EXCLUDED.combinable,
			stackable           = EXCLUDED.stackable,
			promotion_type      = EXCLUDED.promotion_type,
			usage_count         = EXCLUDED.usage_count,
			updated_at          = EXCLUDED.updated_at`,
		v.ID, v.Code, v.Name, v.Description, string(v.Status), v.Priority,
		v.StartTime.UTC(), v.EndTime.UTC(), v.MaxUsage, v.MaxUsagePerUser,
		v.Budget, v.RemainingBudget, string(v.DiscountType), v.DiscountValue, v.MaxDiscount, v.MinOrder,
		nonNilSlice(v.VehicleTypes), nonNilSlice(v.Cities), nonNilSlice(v.Membership),
		v.NewUserOnly, v.Combinable, v.Stackable, string(v.Type), v.UsageCount,
		v.CreatedAt.UTC(), v.UpdatedAt.UTC(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.AlreadyExists("voucher code already in use")
		}
		return domainerrors.Internal("save voucher").WithMeta("error", err.Error())
	}
	return nil
}

func (r *VoucherRepository) UsageCountForRider(ctx context.Context, voucherID, riderID string) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM voucher_redemptions
		WHERE voucher_id = $1 AND rider_id = $2 AND status = 'redeemed'`,
		voucherID, riderID,
	).Scan(&count)
	if err != nil {
		return 0, domainerrors.Internal("query voucher usage count").WithMeta("error", err.Error())
	}
	return count, nil
}

// RecordRedemption atomically reserves budget/usage on the voucher and
// inserts a redemption event, in a single transaction. The UPDATE's WHERE
// clause re-checks budget and usage at commit time so two concurrent
// redemptions cannot both succeed past the voucher's limits (the earlier
// application-layer check in VoucherValidator is necessary but not
// sufficient under concurrency — this is the authoritative guard).
func (r *VoucherRepository) RecordRedemption(ctx context.Context, voucherID, riderID, tripID string, discountAmount int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domainerrors.Internal("begin transaction").WithMeta("error", err.Error())
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	now := time.Now().UTC()
	tag, err := tx.Exec(ctx, `
		UPDATE vouchers
		SET remaining_budget = remaining_budget - $2,
		    usage_count      = usage_count + 1,
		    status           = CASE WHEN remaining_budget - $2 <= 0 THEN 'exhausted' ELSE status END,
		    updated_at       = $3
		WHERE id = $1
		  AND status = 'active'
		  AND remaining_budget >= $2
		  AND (max_usage = 0 OR usage_count < max_usage)`,
		voucherID, discountAmount, now,
	)
	if err != nil {
		return domainerrors.Internal("reserve voucher budget").WithMeta("error", err.Error())
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.ResourceExhausted("voucher budget or usage quota exhausted, or voucher not active")
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO voucher_redemptions (voucher_id, rider_id, trip_id, discount_amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'redeemed', $5, $5)`,
		voucherID, riderID, tripID, discountAmount, now,
	)
	if err != nil {
		return domainerrors.Internal("insert voucher redemption").WithMeta("error", err.Error())
	}

	if err := tx.Commit(ctx); err != nil {
		return domainerrors.Internal("commit voucher redemption").WithMeta("error", err.Error())
	}
	return nil
}

// ReleaseRedemption reverses RecordRedemption for a specific rider/voucher/trip
// (BRB §4.13 Refund Behaviour / §4.14 Cancellation Behaviour).
func (r *VoucherRepository) ReleaseRedemption(ctx context.Context, voucherID, riderID, tripID string, discountAmount int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domainerrors.Internal("begin transaction").WithMeta("error", err.Error())
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	now := time.Now().UTC()
	tag, err := tx.Exec(ctx, `
		UPDATE voucher_redemptions
		SET status = 'released', updated_at = $4
		WHERE voucher_id = $1 AND rider_id = $2 AND trip_id = $3 AND status = 'redeemed'`,
		voucherID, riderID, tripID, now,
	)
	if err != nil {
		return domainerrors.Internal("release voucher redemption").WithMeta("error", err.Error())
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.PreconditionFailed("no active redemption found for this rider/voucher/trip")
	}

	_, err = tx.Exec(ctx, `
		UPDATE vouchers
		SET remaining_budget = LEAST(budget, remaining_budget + $2),
		    usage_count      = GREATEST(usage_count - 1, 0),
		    status           = CASE WHEN status = 'exhausted' AND remaining_budget + $2 > 0 THEN 'active' ELSE status END,
		    updated_at       = $3
		WHERE id = $1`,
		voucherID, discountAmount, now,
	)
	if err != nil {
		return domainerrors.Internal("reinstate voucher budget").WithMeta("error", err.Error())
	}

	if err := tx.Commit(ctx); err != nil {
		return domainerrors.Internal("commit voucher release").WithMeta("error", err.Error())
	}
	return nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *VoucherRepository) queryOne(ctx context.Context, sql string, args ...any) (*entity.Voucher, error) {
	row := r.pool.QueryRow(ctx, sql, args...)
	v, err := scanVoucher(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("voucher not found")
		}
		return nil, domainerrors.Internal("query voucher").WithMeta("error", err.Error())
	}
	return v, nil
}

func scanVoucher(row rowScanner) (*entity.Voucher, error) {
	var (
		id, code, name, description, status string
		priority                            int
		startTime, endTime                  time.Time
		maxUsage, maxUsagePerUser            int64
		budget, remainingBudget              int64
		discountType                        string
		discountValue, maxDiscount, minOrder int64
		vehicleTypes, cities, membership     []string
		newUserOnly, combinable, stackable   bool
		promotionType                        string
		usageCount                           int64
		createdAt, updatedAt                 time.Time
	)

	err := row.Scan(
		&id, &code, &name, &description, &status, &priority,
		&startTime, &endTime, &maxUsage, &maxUsagePerUser,
		&budget, &remainingBudget, &discountType, &discountValue, &maxDiscount, &minOrder,
		&vehicleTypes, &cities, &membership,
		&newUserOnly, &combinable, &stackable, &promotionType, &usageCount,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	return entity.ReconstituteVoucher(
		id, code, name, description,
		entity.VoucherStatus(status),
		priority,
		startTime.UTC(), endTime.UTC(),
		maxUsage, maxUsagePerUser, budget, remainingBudget,
		entity.DiscountType(discountType), discountValue, maxDiscount, minOrder,
		vehicleTypes, cities, membership,
		newUserOnly, combinable, stackable,
		entity.PromotionType(promotionType),
		usageCount,
		createdAt.UTC(), updatedAt.UTC(),
	), nil
}

func nonNilSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

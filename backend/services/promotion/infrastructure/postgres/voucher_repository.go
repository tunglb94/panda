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
	vehicle_types, cities, membership, service_types, trip_types, campaign,
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
			vehicle_types, cities, membership, service_types, trip_types, campaign,
			new_user_only, combinable, stackable, promotion_type, usage_count,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22,
			$23, $24, $25, $26, $27,
			$28, $29
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
			service_types       = EXCLUDED.service_types,
			trip_types          = EXCLUDED.trip_types,
			campaign            = EXCLUDED.campaign,
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
		nonNilSlice(v.ServiceTypes), nonNilSlice(v.TripTypes), v.Campaign,
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

// FindAll returns every voucher campaign, newest first — the Admin app's CRUD list.
func (r *VoucherRepository) FindAll(ctx context.Context) ([]*entity.Voucher, error) {
	rows, err := r.pool.Query(ctx, voucherSelect+` ORDER BY created_at DESC`)
	if err != nil {
		return nil, domainerrors.Internal("query all vouchers").WithMeta("error", err.Error())
	}
	defer rows.Close()
	return scanVouchers(rows)
}

// ListRedemptionsByRider returns riderID's full redemption history, newest
// first — the Rider app's voucher wallet "Used"/"Expired" tabs.
func (r *VoucherRepository) ListRedemptionsByRider(ctx context.Context, riderID string) ([]*entity.RedemptionRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT vr.voucher_id, v.code, v.name, vr.rider_id, vr.trip_id, vr.discount_amount, vr.status, vr.created_at
		FROM voucher_redemptions vr
		JOIN vouchers v ON v.id = vr.voucher_id
		WHERE vr.rider_id = $1
		ORDER BY vr.created_at DESC`, riderID)
	if err != nil {
		return nil, domainerrors.Internal("query rider redemptions").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var result []*entity.RedemptionRecord
	for rows.Next() {
		var rec entity.RedemptionRecord
		if err := rows.Scan(&rec.VoucherID, &rec.VoucherCode, &rec.VoucherName, &rec.RiderID, &rec.TripID, &rec.DiscountAmount, &rec.Status, &rec.RedeemedAt); err != nil {
			return nil, domainerrors.Internal("scan redemption").WithMeta("error", err.Error())
		}
		rec.RedeemedAt = rec.RedeemedAt.UTC()
		result = append(result, &rec)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("iterate redemptions").WithMeta("error", err.Error())
	}
	return result, nil
}

// UsageCountForRider counts 'reserved' + 'redeemed' rows — an in-flight
// reservation already counts against the BRB §4.6 per-rider limit so the
// same rider can't hold two concurrent bookings against a single-use voucher.
func (r *VoucherRepository) UsageCountForRider(ctx context.Context, voucherID, riderID string) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM voucher_redemptions
		WHERE voucher_id = $1 AND rider_id = $2 AND status IN ('reserved', 'redeemed')`,
		voucherID, riderID,
	).Scan(&count)
	if err != nil {
		return 0, domainerrors.Internal("query voucher usage count").WithMeta("error", err.Error())
	}
	return count, nil
}

// Reserve atomically deducts budget/usage on the voucher and inserts a
// 'reserved' redemption row, in a single transaction. The UPDATE's WHERE
// clause re-checks budget and usage at commit time so two concurrent
// reservations cannot both succeed past the voucher's limits (the earlier
// application-layer check in VoucherValidator is necessary but not
// sufficient under concurrency — this is the authoritative guard).
// Idempotent: ON CONFLICT on (voucher_id, rider_id, trip_id) makes a retried
// Reserve for the same trip a no-op rather than a double deduction.
func (r *VoucherRepository) Reserve(ctx context.Context, voucherID, riderID, tripID string, discountAmount int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domainerrors.Internal("begin transaction").WithMeta("error", err.Error())
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var alreadyExists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM voucher_redemptions WHERE voucher_id = $1 AND rider_id = $2 AND trip_id = $3)`,
		voucherID, riderID, tripID,
	).Scan(&alreadyExists); err != nil {
		return domainerrors.Internal("check existing reservation").WithMeta("error", err.Error())
	}
	if alreadyExists {
		return nil // idempotent: already reserved (or further along) for this trip
	}

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
		VALUES ($1, $2, $3, $4, 'reserved', $5, $5)`,
		voucherID, riderID, tripID, discountAmount, now,
	)
	if err != nil {
		return domainerrors.Internal("insert voucher reservation").WithMeta("error", err.Error())
	}

	if err := tx.Commit(ctx); err != nil {
		return domainerrors.Internal("commit voucher reservation").WithMeta("error", err.Error())
	}
	return nil
}

// ConfirmRedeem transitions a 'reserved' row to 'redeemed'. No budget
// change (already deducted at Reserve time). Idempotent.
func (r *VoucherRepository) ConfirmRedeem(ctx context.Context, voucherID, riderID, tripID string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE voucher_redemptions
		SET status = 'redeemed', updated_at = $4
		WHERE voucher_id = $1 AND rider_id = $2 AND trip_id = $3 AND status = 'reserved'`,
		voucherID, riderID, tripID, time.Now().UTC(),
	)
	if err != nil {
		return domainerrors.Internal("confirm voucher redemption").WithMeta("error", err.Error())
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	// 0 rows: either already 'redeemed' (idempotent success) or genuinely missing.
	var status string
	err = r.pool.QueryRow(ctx, `
		SELECT status FROM voucher_redemptions WHERE voucher_id = $1 AND rider_id = $2 AND trip_id = $3`,
		voucherID, riderID, tripID,
	).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainerrors.NotFound("no voucher reservation found for this trip")
		}
		return domainerrors.Internal("check voucher redemption status").WithMeta("error", err.Error())
	}
	if status == "redeemed" {
		return nil
	}
	return domainerrors.PreconditionFailed("voucher reservation is not in a redeemable state: " + status)
}

// Release transitions a 'reserved' row to 'released' and reinstates
// discountAmount to the voucher's budget (BRB §4.13 Refund Behaviour /
// §4.14 Cancellation Behaviour). Idempotent.
func (r *VoucherRepository) Release(ctx context.Context, voucherID, riderID, tripID string, discountAmount int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domainerrors.Internal("begin transaction").WithMeta("error", err.Error())
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	now := time.Now().UTC()
	tag, err := tx.Exec(ctx, `
		UPDATE voucher_redemptions
		SET status = 'released', updated_at = $4
		WHERE voucher_id = $1 AND rider_id = $2 AND trip_id = $3 AND status = 'reserved'`,
		voucherID, riderID, tripID, now,
	)
	if err != nil {
		return domainerrors.Internal("release voucher redemption").WithMeta("error", err.Error())
	}
	if tag.RowsAffected() == 0 {
		var status string
		scanErr := tx.QueryRow(ctx, `
			SELECT status FROM voucher_redemptions WHERE voucher_id = $1 AND rider_id = $2 AND trip_id = $3`,
			voucherID, riderID, tripID,
		).Scan(&status)
		if scanErr == nil && status == "released" {
			return nil // idempotent: already released
		}
		return domainerrors.PreconditionFailed("no reserved voucher redemption found for this rider/voucher/trip")
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

// FindReservationByTrip returns the redemption row for tripID regardless of
// status (one voucher per trip, BRB §4.7).
func (r *VoucherRepository) FindReservationByTrip(ctx context.Context, tripID string) (*entity.RedemptionRecord, error) {
	var rec entity.RedemptionRecord
	err := r.pool.QueryRow(ctx, `
		SELECT vr.voucher_id, v.code, v.name, vr.rider_id, vr.trip_id, vr.discount_amount, vr.status, vr.created_at
		FROM voucher_redemptions vr
		JOIN vouchers v ON v.id = vr.voucher_id
		WHERE vr.trip_id = $1`, tripID,
	).Scan(&rec.VoucherID, &rec.VoucherCode, &rec.VoucherName, &rec.RiderID, &rec.TripID, &rec.DiscountAmount, &rec.Status, &rec.RedeemedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("no voucher reservation found for this trip")
		}
		return nil, domainerrors.Internal("query trip reservation").WithMeta("error", err.Error())
	}
	rec.RedeemedAt = rec.RedeemedAt.UTC()
	return &rec, nil
}

// ─── Per-rider issuance ─────────────────────────────────────────────────────

// IssueToRider grants voucherID to riderID. Idempotent — ON CONFLICT DO
// NOTHING so a re-issue never resets an already-'used' issuance.
func (r *VoucherRepository) IssueToRider(ctx context.Context, voucherID, riderID string, now time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO voucher_issuances (voucher_id, rider_id, status, issued_at)
		VALUES ($1, $2, 'issued', $3)
		ON CONFLICT (voucher_id, rider_id) DO NOTHING`,
		voucherID, riderID, now.UTC(),
	)
	if err != nil {
		return domainerrors.Internal("issue voucher to rider").WithMeta("error", err.Error())
	}
	return nil
}

// ListIssuancesForRider returns every voucher issued to riderID, newest
// first, with Voucher populated via JOIN.
func (r *VoucherRepository) ListIssuancesForRider(ctx context.Context, riderID string) ([]*entity.VoucherIssuance, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT vi.voucher_id, vi.rider_id, vi.status, vi.issued_at, vi.used_at, `+voucherFields+`
		FROM voucher_issuances vi
		JOIN vouchers v ON v.id = vi.voucher_id
		WHERE vi.rider_id = $1
		ORDER BY vi.issued_at DESC`, riderID)
	if err != nil {
		return nil, domainerrors.Internal("query rider issuances").WithMeta("error", err.Error())
	}
	defer rows.Close()

	var result []*entity.VoucherIssuance
	for rows.Next() {
		var iss entity.VoucherIssuance
		var status string
		var usedAt *time.Time
		if err := rows.Scan(&iss.VoucherID, &iss.RiderID, &status, &iss.IssuedAt, &usedAt); err != nil {
			return nil, domainerrors.Internal("scan issuance").WithMeta("error", err.Error())
		}
		v, err := scanVoucher(rows)
		if err != nil {
			return nil, domainerrors.Internal("scan issuance voucher").WithMeta("error", err.Error())
		}
		iss.Status = entity.VoucherIssuanceStatus(status)
		iss.IssuedAt = iss.IssuedAt.UTC()
		if usedAt != nil {
			t := usedAt.UTC()
			iss.UsedAt = &t
		}
		iss.Voucher = v
		result = append(result, &iss)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerrors.Internal("iterate issuances").WithMeta("error", err.Error())
	}
	return result, nil
}

// MarkIssuanceUsed transitions an issuance to 'used' — best-effort, no-op
// if no issuance row exists for this (voucher, rider) pair.
func (r *VoucherRepository) MarkIssuanceUsed(ctx context.Context, voucherID, riderID string, now time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE voucher_issuances SET status = 'used', used_at = $3
		WHERE voucher_id = $1 AND rider_id = $2 AND status = 'issued'`,
		voucherID, riderID, now.UTC(),
	)
	if err != nil {
		return domainerrors.Internal("mark issuance used").WithMeta("error", err.Error())
	}
	return nil
}

// CountIssued returns how many riders voucherID has ever been issued to.
func (r *VoucherRepository) CountIssued(ctx context.Context, voucherID string) (int64, error) {
	return r.countQuery(ctx, `SELECT COUNT(*) FROM voucher_issuances WHERE voucher_id = $1`, voucherID)
}

// CountRedeemed returns how many times voucherID has been permanently
// redeemed (status='redeemed' — reserved-but-not-yet-completed trips don't count).
func (r *VoucherRepository) CountRedeemed(ctx context.Context, voucherID string) (int64, error) {
	return r.countQuery(ctx, `SELECT COUNT(*) FROM voucher_redemptions WHERE voucher_id = $1 AND status = 'redeemed'`, voucherID)
}

// CountExpiredIssuances returns how many riders were issued voucherID, never
// used it, and the campaign's window has now passed.
func (r *VoucherRepository) CountExpiredIssuances(ctx context.Context, voucherID string, now time.Time) (int64, error) {
	return r.countQuery(ctx, `
		SELECT COUNT(*) FROM voucher_issuances vi
		JOIN vouchers v ON v.id = vi.voucher_id
		WHERE vi.voucher_id = $1 AND vi.status = 'issued' AND v.end_time < $2`,
		voucherID, now.UTC())
}

func (r *VoucherRepository) countQuery(ctx context.Context, sql string, args ...any) (int64, error) {
	var count int64
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, domainerrors.Internal("count query failed").WithMeta("error", err.Error())
	}
	return count, nil
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
		id, code, name, description, status  string
		priority                             int
		startTime, endTime                   time.Time
		maxUsage, maxUsagePerUser            int64
		budget, remainingBudget              int64
		discountType                         string
		discountValue, maxDiscount, minOrder int64
		vehicleTypes, cities, membership     []string
		serviceTypes, tripTypes              []string
		campaign                             string
		newUserOnly, combinable, stackable   bool
		promotionType                        string
		usageCount                           int64
		createdAt, updatedAt                 time.Time
	)

	err := row.Scan(
		&id, &code, &name, &description, &status, &priority,
		&startTime, &endTime, &maxUsage, &maxUsagePerUser,
		&budget, &remainingBudget, &discountType, &discountValue, &maxDiscount, &minOrder,
		&vehicleTypes, &cities, &membership, &serviceTypes, &tripTypes, &campaign,
		&newUserOnly, &combinable, &stackable, &promotionType, &usageCount,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	v := entity.ReconstituteVoucher(
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
	)
	// ServiceTypes/TripTypes/Campaign postdate ReconstituteVoucher's original
	// signature — set directly rather than widening that constructor and
	// touching every existing call site (entity fields are exported for
	// exactly this kind of additive extension).
	v.ServiceTypes = serviceTypes
	v.TripTypes = tripTypes
	v.Campaign = campaign
	return v, nil
}

// scanVouchers drains a multi-row query into a slice, sharing scanVoucher's
// column layout — used by FindAll/ListPublicActive.
func scanVouchers(rows pgx.Rows) ([]*entity.Voucher, error) {
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

func nonNilSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

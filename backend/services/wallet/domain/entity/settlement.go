package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// TripType mirrors the Trip service's own TripType wire values ("ride" |
// "delivery") — duplicated here (not imported) because Settlement must not
// depend on the Trip service's package (Ride/Delivery Lifecycle stay
// untouched and unreferenced by Financial Core).
type TripType string

const (
	TripTypeRide     TripType = "ride"
	TripTypeDelivery TripType = "delivery"
)

// PaymentMethod is the Settlement Case discriminator (Phần 2):
// PaymentMethodCash -> Case 1 (driver holds the fare, owes commission);
// PaymentMethodWallet -> Case 2 (platform holds the fare, owes driver income).
type PaymentMethod string

const (
	PaymentMethodCash   PaymentMethod = "cash"
	PaymentMethodWallet PaymentMethod = "wallet"
)

func validPaymentMethod(m PaymentMethod) bool {
	return m == PaymentMethodCash || m == PaymentMethodWallet
}

// SettlementStatus guards against a crash between "claimed this trip_id"
// and "finished posting every ledger entry" (critique #7): the engine
// inserts the Settlement row as StatusPending BEFORE writing any ledger
// entries (using deterministic sub-entity IDs derived from SettlementID),
// so a retry after a mid-way crash can detect the pending row and safely
// resume — re-attempting each write idempotently — rather than either
// re-computing from scratch (risking duplicate postings with fresh random
// IDs) or silently returning a half-posted Settlement as if it were done.
type SettlementStatus string

const (
	SettlementStatusPending SettlementStatus = "pending"
	SettlementStatusPosted  SettlementStatus = "posted"
)

// VoucherStatus distinguishes a confirmed "no voucher on this trip" from
// "Promotion Engine detail was not available at settlement time" — both
// currently produce VoucherCostCents == 0, but they mean very different
// things, and collapsing them made future Promotion-Engine migration harder
// to reason about (critique #2). VoucherStatusUnknown is the case today for
// every trip settled while Pricing runs V2 (no voucher detail computed at
// all); VoucherStatusApplied/None are only ever set when
// CreateSettlementInput.HasCommissionDetail was true (Pricing V3).
type VoucherStatus string

const (
	VoucherStatusUnknown VoucherStatus = "unknown"
	VoucherStatusNone    VoucherStatus = "none"
	VoucherStatusApplied VoucherStatus = "applied"
)

// Settlement is the immutable audit record of one trip's financial
// resolution (Phần 2/13) — created exactly once per trip (enforced by a
// UNIQUE(trip_id) DB constraint, making the Settlement Engine idempotent
// against retries) by the Settlement Engine, never edited afterward.
// TransactionID links to the wallet_transactions row that carries the
// actual LedgerEntry postings this Settlement produced.
type Settlement struct {
	SettlementID          string
	TripID                string
	DriverID              string
	TripType              TripType
	PaymentMethod         PaymentMethod
	FareAmountCents       int64
	CommissionRate        float64 // snapshot at settlement time, e.g. 0.20
	CommissionAmountCents int64
	DriverIncomeCents     int64 // what the driver nets from this trip
	PromotionSubsidyCents int64 // Case 3 — always 0 today, see Known Gap
	VoucherCostCents      int64 // Case 3 — always 0 today, see Known Gap
	Currency              string
	TransactionID         string
	CreatedAt             time.Time
	Status                SettlementStatus
	VoucherStatus         VoucherStatus
}

// NewSettlement creates a validated Settlement.
func NewSettlement(
	settlementID, tripID, driverID string,
	tripType TripType,
	paymentMethod PaymentMethod,
	fareAmountCents int64,
	commissionRate float64,
	commissionAmountCents, driverIncomeCents, promotionSubsidyCents, voucherCostCents int64,
	currency, transactionID string,
	now time.Time,
	status SettlementStatus,
	voucherStatus VoucherStatus,
) (*Settlement, error) {
	if strings.TrimSpace(settlementID) == "" {
		return nil, errors.InvalidArgument("settlement id must not be empty")
	}
	if strings.TrimSpace(tripID) == "" {
		return nil, errors.InvalidArgument("trip id must not be empty")
	}
	if strings.TrimSpace(driverID) == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	if tripType != TripTypeRide && tripType != TripTypeDelivery {
		return nil, errors.InvalidArgument("invalid trip type: " + string(tripType))
	}
	if !validPaymentMethod(paymentMethod) {
		return nil, errors.InvalidArgument("invalid payment method: " + string(paymentMethod))
	}
	if fareAmountCents <= 0 {
		return nil, errors.InvalidArgument("fare_amount_cents must be greater than zero")
	}
	if commissionRate < 0 || commissionRate >= 1 {
		return nil, errors.InvalidArgument("commission_rate must be in [0, 1)")
	}
	if commissionAmountCents < 0 || driverIncomeCents < 0 {
		return nil, errors.InvalidArgument("commission/driver income must not be negative")
	}
	if strings.TrimSpace(currency) == "" {
		return nil, errors.InvalidArgument("currency must not be empty")
	}
	if strings.TrimSpace(transactionID) == "" {
		return nil, errors.InvalidArgument("transaction id must not be empty")
	}
	if status != SettlementStatusPending && status != SettlementStatusPosted {
		status = SettlementStatusPosted
	}
	if voucherStatus != VoucherStatusNone && voucherStatus != VoucherStatusApplied {
		voucherStatus = VoucherStatusUnknown
	}
	return &Settlement{
		SettlementID:          settlementID,
		TripID:                tripID,
		DriverID:              driverID,
		TripType:              tripType,
		PaymentMethod:         paymentMethod,
		FareAmountCents:       fareAmountCents,
		CommissionRate:        commissionRate,
		CommissionAmountCents: commissionAmountCents,
		DriverIncomeCents:     driverIncomeCents,
		PromotionSubsidyCents: promotionSubsidyCents,
		VoucherCostCents:      voucherCostCents,
		Currency:              currency,
		TransactionID:         transactionID,
		CreatedAt:             now,
		Status:                status,
		VoucherStatus:         voucherStatus,
	}, nil
}

// ReconstituteSettlement rebuilds a Settlement from persistence without validation.
func ReconstituteSettlement(
	settlementID, tripID, driverID string,
	tripType TripType,
	paymentMethod PaymentMethod,
	fareAmountCents int64,
	commissionRate float64,
	commissionAmountCents, driverIncomeCents, promotionSubsidyCents, voucherCostCents int64,
	currency, transactionID string,
	createdAt time.Time,
	status SettlementStatus,
	voucherStatus VoucherStatus,
) *Settlement {
	return &Settlement{
		SettlementID:          settlementID,
		TripID:                tripID,
		DriverID:              driverID,
		TripType:              tripType,
		PaymentMethod:         paymentMethod,
		FareAmountCents:       fareAmountCents,
		CommissionRate:        commissionRate,
		CommissionAmountCents: commissionAmountCents,
		DriverIncomeCents:     driverIncomeCents,
		PromotionSubsidyCents: promotionSubsidyCents,
		VoucherCostCents:      voucherCostCents,
		Currency:              currency,
		TransactionID:         transactionID,
		CreatedAt:             createdAt,
		Status:                status,
		VoucherStatus:         voucherStatus,
	}
}

// IsCashCollected reports whether this Settlement is Case 1 (driver
// collected cash, owes the platform its commission — see Phần 4 Outstanding).
func (s *Settlement) IsCashCollected() bool { return s.PaymentMethod == PaymentMethodCash }

package app

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

// DefaultCommissionRate is the FALLBACK take-rate (20%), used only when the
// caller could not resolve a real commission figure from Pricing V3 (i.e.
// CreateSettlementInput.HasCommissionDetail is false — Pricing running V2,
// or an older Trip row completed before commission detail was persisted).
// Whenever HasCommissionDetail is true, Execute uses Pricing's own computed
// Commission/DriverIncome/CommissionRate as-is — Settlement never invents a
// number Pricing already computed (see gateway's SettlementEngine.Settle,
// which resolves this from Trip's persisted CompleteFinancials).
const DefaultCommissionRate = 0.20

// CreateSettlementInput carries everything the Settlement Engine needs to
// resolve one trip's finances (Phần 2). FareAmountCents must already be the
// final, post-trip fare (Trip.FinalFareTotal) — Settlement never computes
// or second-guesses the fare itself (that's Pricing's formula, untouched).
type CreateSettlementInput struct {
	TripID          string
	DriverID        string
	TripType        entity.TripType
	PaymentMethod   entity.PaymentMethod
	FareAmountCents int64
	Currency        string

	// HasCommissionDetail is true when the caller resolved a real commission
	// figure from Pricing V3 (via Trip's persisted CompleteFinancials — see
	// gateway/http/handlers/settlement_engine.go). When true,
	// CommissionCents/DriverIncomeCents are used as-is instead of Settlement
	// computing its own flat-rate commission — Settlement must never invent
	// a number Pricing already computed.
	HasCommissionDetail bool
	CommissionCents     int64
	DriverIncomeCents   int64
	CommissionRate      float64

	// VoucherDiscountCents is Pricing V3's own voucher/promotion discount
	// for this trip — only meaningful when HasCommissionDetail is true
	// (critique #2: distinguish "no voucher" from "unknown, Promotion not
	// wired in for this trip" instead of silently recording 0 either way).
	VoucherDiscountCents int64
}

// CreateSettlementUseCase is the Settlement Engine (Phần 2/13) — the single
// place a completed, paid trip becomes Ledger entries. Idempotent: calling
// Execute twice for the same TripID returns the original Settlement rather
// than double-posting (UNIQUE(trip_id) in the migration is the actual
// guarantee; the pre-check here just avoids an unnecessary round-trip and
// keeps this safe to call from an at-least-once delivery path).
//
// Ledger First: this use case never writes to a "balance" column — it only
// ever appends Transactions + LedgerEntries. Wallet balances are always
// read back by summing the ledger (see get_wallet_summary.go).
type CreateSettlementUseCase struct {
	settlements    repository.SettlementRepository
	wallets        *GetOrCreateWalletUseCase
	ledger         repository.LedgerEntryRepository
	tx             repository.TransactionRepository
	audit          repository.AuditLogRepository
	commissionRate float64
}

func NewCreateSettlementUseCase(
	settlements repository.SettlementRepository,
	wallets *GetOrCreateWalletUseCase,
	ledger repository.LedgerEntryRepository,
	tx repository.TransactionRepository,
	audit repository.AuditLogRepository,
	commissionRate float64,
) *CreateSettlementUseCase {
	if commissionRate <= 0 || commissionRate >= 1 {
		commissionRate = DefaultCommissionRate
	}
	return &CreateSettlementUseCase{
		settlements: settlements, wallets: wallets, ledger: ledger, tx: tx, audit: audit,
		commissionRate: commissionRate,
	}
}

func (uc *CreateSettlementUseCase) Execute(ctx context.Context, in CreateSettlementInput) (*entity.Settlement, error) {
	if strings.TrimSpace(in.TripID) == "" {
		return nil, errors.InvalidArgument("trip_id must not be empty")
	}
	if strings.TrimSpace(in.DriverID) == "" {
		return nil, errors.InvalidArgument("driver_id must not be empty")
	}
	if in.FareAmountCents <= 0 {
		return nil, errors.InvalidArgument("fare_amount_cents must be greater than zero")
	}
	if in.Currency == "" {
		in.Currency = DefaultCurrency
	}
	if in.PaymentMethod != entity.PaymentMethodCash && in.PaymentMethod != entity.PaymentMethodWallet {
		in.PaymentMethod = entity.PaymentMethodCash // real-world default: no payment gateway is wired, cash is what actually happens today
	}
	if in.TripType != entity.TripTypeRide && in.TripType != entity.TripTypeDelivery {
		in.TripType = entity.TripTypeRide
	}

	// Idempotency fast-path: a fully posted Settlement short-circuits
	// immediately. A Pending one (crash between the claim below and the
	// last ledger write) falls through to resume posting, using the exact
	// same deterministic sub-IDs, rather than being returned as if done.
	settlementID := "stl-" + in.TripID
	resuming := false
	if existing, err := uc.settlements.FindByTripID(ctx, in.TripID); err == nil {
		if existing.Status == entity.SettlementStatusPosted {
			return existing, nil
		}
		resuming = true
	} else if errors.GetCode(err) != errors.CodeNotFound {
		return nil, err
	}

	driverWallet, err := uc.wallets.Execute(ctx, in.DriverID, entity.WalletTypeDriver)
	if err != nil {
		return nil, err
	}
	platformWallet, err := uc.wallets.Execute(ctx, PlatformOwnerID, entity.WalletTypePlatform)
	if err != nil {
		return nil, err
	}

	commissionRate := uc.commissionRate
	var commissionAmount, driverIncome int64
	if in.HasCommissionDetail {
		// Read Pricing's own computed number — never invent one when it's
		// available (the fix for critique #1: Settlement must not self-compute
		// commission when Pricing already did).
		commissionRate = in.CommissionRate
		commissionAmount = in.CommissionCents
		driverIncome = in.DriverIncomeCents
	} else {
		// Pricing V2 (or an unresolved Trip financial detail) — fall back to
		// the documented flat rate; this path is a known, labeled gap, not a
		// silent invention.
		commissionAmount = int64(math.Round(float64(in.FareAmountCents) * uc.commissionRate))
		if commissionAmount > in.FareAmountCents {
			commissionAmount = in.FareAmountCents
		}
		driverIncome = in.FareAmountCents - commissionAmount
	}

	incomeType := entity.TypeRideIncome
	if in.TripType == entity.TripTypeDelivery {
		incomeType = entity.TypeDeliveryIncome
	}
	now := time.Now().UTC()

	// Voucher pending marker (critique #2): only Pricing V3 computes a real
	// voucher/promotion discount figure today. Distinguish "confirmed no
	// voucher" from "unknown, Promotion detail unavailable" rather than
	// recording 0 for both — see entity.VoucherStatus.
	voucherCost := int64(0)
	voucherStatus := entity.VoucherStatusUnknown
	if in.HasCommissionDetail {
		voucherCost = in.VoucherDiscountCents
		if voucherCost > 0 {
			voucherStatus = entity.VoucherStatusApplied
		} else {
			voucherStatus = entity.VoucherStatusNone
		}
	}

	// Deterministic sub-entity IDs (derived from settlementID, itself
	// derived from TripID) — critique #7's crash-safety fix. A retry after
	// a mid-way crash reuses the exact same IDs, so re-attempting a write
	// that already landed hits the table's own unique-ID constraint
	// (AlreadyExists, swallowed by save()/saveEntry() below) instead of
	// double-posting money under a fresh random ID.
	incomeTxID := settlementID + "-income-tx"
	commissionTxID := settlementID + "-commission-tx"
	payableTxID := settlementID + "-payable-tx"

	// Insert the Settlement claim BEFORE any ledger entry — this is the
	// crash-safety checkpoint: from this point on, a retry sees Pending and
	// resumes (via idempotent sub-writes) instead of restarting from
	// scratch with new random IDs.
	settlement, err := entity.NewSettlement(
		settlementID, in.TripID, in.DriverID, in.TripType, in.PaymentMethod,
		in.FareAmountCents, commissionRate, commissionAmount, driverIncome, 0, voucherCost,
		in.Currency, incomeTxID, now, entity.SettlementStatusPending, voucherStatus,
	)
	if err != nil {
		return nil, err
	}
	if !resuming {
		if err := uc.settlements.Save(ctx, settlement); err != nil {
			if errors.GetCode(err) == errors.CodeAlreadyExists {
				resuming = true // lost a race with a concurrent caller; resume alongside it
			} else {
				return nil, err
			}
		}
	}

	if err := saveTx(ctx, uc.tx, incomeTxID, incomeType, in, tripIncomeDescription(in), now); err != nil {
		return nil, err
	}
	if err := saveEntry(ctx, uc.ledger, incomeTxID+"-entry", driverWallet.WalletID, incomeTxID, entity.DirectionCredit, driverIncome, in.Currency, tripIncomeDescription(in), now); err != nil {
		return nil, err
	}
	if err := saveTx(ctx, uc.tx, commissionTxID, entity.TypeCommission, in, "Hoa hồng chuyến "+in.TripID, now); err != nil {
		return nil, err
	}

	if in.PaymentMethod == entity.PaymentMethodCash {
		// Case 1 — driver already holds the fare in cash; they now owe the
		// platform its commission share (Outstanding).
		if err := saveEntry(ctx, uc.ledger, commissionTxID+"-driver-debit", driverWallet.WalletID, commissionTxID, entity.DirectionDebit, commissionAmount, in.Currency, "Nợ hoa hồng chuyến "+in.TripID, now); err != nil {
			return nil, err
		}
		if err := saveEntry(ctx, uc.ledger, commissionTxID+"-platform-credit", platformWallet.WalletID, commissionTxID, entity.DirectionCredit, commissionAmount, in.Currency, "Phải thu hoa hồng chuyến "+in.TripID, now); err != nil {
			return nil, err
		}
	} else {
		// Case 2 — the platform collected the fare electronically and now
		// owes the driver their net income (already credited above); book
		// the platform's own commission revenue and discharge its payable.
		if err := saveEntry(ctx, uc.ledger, commissionTxID+"-platform-credit", platformWallet.WalletID, commissionTxID, entity.DirectionCredit, commissionAmount, in.Currency, "Doanh thu hoa hồng chuyến "+in.TripID, now); err != nil {
			return nil, err
		}
		if err := saveTx(ctx, uc.tx, payableTxID, entity.TypePlatformPayable, in, "Panda trả tài xế chuyến "+in.TripID, now); err != nil {
			return nil, err
		}
		if err := saveEntry(ctx, uc.ledger, payableTxID+"-platform-debit", platformWallet.WalletID, payableTxID, entity.DirectionDebit, driverIncome, in.Currency, "Panda trả tài xế chuyến "+in.TripID, now); err != nil {
			return nil, err
		}
	}

	if err := uc.settlements.MarkPosted(ctx, settlementID); err != nil {
		return nil, err
	}
	settlement.Status = entity.SettlementStatusPosted
	_ = recordAudit(ctx, uc.audit, entity.AuditEntitySettlement, settlement.SettlementID, in.DriverID, entity.AuditActionCreate, "system", "", settlementSnapshot(settlement), "")
	return settlement, nil
}

// saveTx creates and saves a Transaction under a deterministic ID,
// tolerating AlreadyExists as "a previous attempt already wrote this" —
// the crash-recovery resume path relies on every sub-write being safely
// re-appliable.
func saveTx(ctx context.Context, repo repository.TransactionRepository, txID string, txType entity.TransactionType, in CreateSettlementInput, description string, now time.Time) error {
	tx, err := entity.NewTransaction(txID, txType, in.TripID, string(in.PaymentMethod), in.Currency, description, now)
	if err != nil {
		return err
	}
	if err := repo.Save(ctx, tx); err != nil && errors.GetCode(err) != errors.CodeAlreadyExists {
		return err
	}
	return nil
}

// saveEntry mirrors saveTx for LedgerEntry.
func saveEntry(ctx context.Context, repo repository.LedgerEntryRepository, entryID, walletID, txID string, direction entity.EntryDirection, amountCents int64, currency, description string, now time.Time) error {
	entry, err := entity.NewLedgerEntry(entryID, walletID, txID, direction, amountCents, currency, description, now)
	if err != nil {
		return err
	}
	if err := repo.Save(ctx, entry); err != nil && errors.GetCode(err) != errors.CodeAlreadyExists {
		return err
	}
	return nil
}

func tripIncomeDescription(in CreateSettlementInput) string {
	label := "chuyến xe"
	if in.TripType == entity.TripTypeDelivery {
		label = "đơn giao hàng"
	}
	return "Thu nhập " + label + " " + in.TripID
}

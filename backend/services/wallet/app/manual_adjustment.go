package app

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

// ManualAdjustmentInput carries an Admin's Manual Adjustment (Phần 10).
// Direction chooses ManualCredit (increases the driver's balance, e.g. a
// goodwill credit or an Outstanding write-off) or ManualDebit (decreases
// it, e.g. correcting an over-payment). Reason is mandatory — Phần 12:
// "Adjustment phải có Audit Log."
type ManualAdjustmentInput struct {
	DriverID    string
	AmountCents int64
	Direction   entity.EntryDirection
	Reason      string
	ActorID     string // admin user id
}

// ManualAdjustmentUseCase lets an Admin post a one-off correction directly
// to a driver's wallet (Phần 10). Ledger First — this never touches a
// balance column, it appends a Transaction + LedgerEntry exactly like every
// other financial event in this module, and is always audit-logged.
type ManualAdjustmentUseCase struct {
	wallets *GetOrCreateWalletUseCase
	ledger  repository.LedgerEntryRepository
	tx      repository.TransactionRepository
	audit   repository.AuditLogRepository
}

func NewManualAdjustmentUseCase(
	wallets *GetOrCreateWalletUseCase,
	ledger repository.LedgerEntryRepository,
	tx repository.TransactionRepository,
	audit repository.AuditLogRepository,
) *ManualAdjustmentUseCase {
	return &ManualAdjustmentUseCase{wallets: wallets, ledger: ledger, tx: tx, audit: audit}
}

func (uc *ManualAdjustmentUseCase) Execute(ctx context.Context, in ManualAdjustmentInput) (*entity.LedgerEntry, error) {
	if strings.TrimSpace(in.DriverID) == "" {
		return nil, errors.InvalidArgument("driver_id must not be empty")
	}
	if in.AmountCents <= 0 {
		return nil, errors.InvalidArgument("amount_cents must be greater than zero")
	}
	if in.Direction != entity.DirectionCredit && in.Direction != entity.DirectionDebit {
		return nil, errors.InvalidArgument("invalid direction: " + string(in.Direction))
	}
	if strings.TrimSpace(in.Reason) == "" {
		return nil, errors.InvalidArgument("reason must not be empty — required for Manual Adjustment audit trail")
	}

	wallet, err := uc.wallets.Execute(ctx, in.DriverID, entity.WalletTypeDriver)
	if err != nil {
		return nil, err
	}
	txType := entity.TypeManualCredit
	if in.Direction == entity.DirectionDebit {
		txType = entity.TypeManualDebit
	}
	now := time.Now().UTC()
	txID := uuid.NewString()
	transaction, err := entity.NewTransaction(txID, txType, "", "", DefaultCurrency, in.Reason, now)
	if err != nil {
		return nil, err
	}
	if err := uc.tx.Save(ctx, transaction); err != nil {
		return nil, err
	}
	entry, err := entity.NewLedgerEntry(uuid.NewString(), wallet.WalletID, txID, in.Direction, in.AmountCents, DefaultCurrency, in.Reason, now)
	if err != nil {
		return nil, err
	}
	if err := uc.ledger.Save(ctx, entry); err != nil {
		return nil, err
	}
	_ = recordAudit(ctx, uc.audit, entity.AuditEntityLedgerEntry, entry.EntryID, in.DriverID, entity.AuditActionManualAdjustment, in.ActorID, "", string(in.Direction)+" "+formatVND(in.AmountCents), in.Reason)
	return entry, nil
}

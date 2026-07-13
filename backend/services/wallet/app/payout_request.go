package app

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

// DefaultMinimumPayoutCents is the Phần 5 "Số tiền ≥ cấu hình tối thiểu"
// floor — 50,000 VND, a realistic minimum for a Vietnamese bank transfer.
// Overridable via NewCreatePayoutRequestUseCase's minAmountCents parameter
// (wired from an env var at the composition root).
const DefaultMinimumPayoutCents = 50_000

// CreatePayoutRequestInput carries Phần 3's "Nhập số tiền" form. KYC
// eligibility (Phần 5 — "Chưa KYC" / "KYC hết hạn") is checked by the
// caller (the gateway, which already owns the driver KYC use cases —
// see the module report's Kiến trúc section for why this use case does
// not import the driver package) before Execute is ever called.
type CreatePayoutRequestInput struct {
	DriverID    string
	AmountCents int64
}

// CreatePayoutRequestUseCase validates and creates a Pending PayoutRequest
// (Phần 5). Every Vietnamese-language rejection reason maps to one of
// Phần 5's explicit checks.
type CreatePayoutRequestUseCase struct {
	payouts        repository.PayoutRequestRepository
	bankAccounts   repository.BankAccountRepository
	walletSummary  *GetWalletSummaryUseCase
	wallets        *GetOrCreateWalletUseCase
	ledger         repository.LedgerEntryRepository
	tx             repository.TransactionRepository
	audit          repository.AuditLogRepository
	minAmountCents int64
}

func NewCreatePayoutRequestUseCase(
	payouts repository.PayoutRequestRepository,
	bankAccounts repository.BankAccountRepository,
	walletSummary *GetWalletSummaryUseCase,
	wallets *GetOrCreateWalletUseCase,
	ledger repository.LedgerEntryRepository,
	tx repository.TransactionRepository,
	audit repository.AuditLogRepository,
	minAmountCents int64,
) *CreatePayoutRequestUseCase {
	if minAmountCents <= 0 {
		minAmountCents = DefaultMinimumPayoutCents
	}
	return &CreatePayoutRequestUseCase{
		payouts: payouts, bankAccounts: bankAccounts, walletSummary: walletSummary,
		wallets: wallets, ledger: ledger, tx: tx, audit: audit,
		minAmountCents: minAmountCents,
	}
}

func (uc *CreatePayoutRequestUseCase) Execute(ctx context.Context, in CreatePayoutRequestInput) (*entity.PayoutRequest, error) {
	if strings.TrimSpace(in.DriverID) == "" {
		return nil, errors.InvalidArgument("driver_id must not be empty")
	}
	if in.AmountCents < uc.minAmountCents {
		return nil, errors.PreconditionFailed("Số tiền rút tối thiểu là " + formatVND(uc.minAmountCents))
	}

	if _, err := uc.payouts.FindInFlightByDriverID(ctx, in.DriverID); err == nil {
		return nil, errors.PreconditionFailed("Bạn đang có yêu cầu rút tiền chờ xử lý")
	} else if errors.GetCode(err) != errors.CodeNotFound {
		return nil, err
	}

	bank, err := uc.bankAccounts.FindByDriverID(ctx, in.DriverID)
	if err != nil {
		if errors.GetCode(err) == errors.CodeNotFound {
			return nil, errors.PreconditionFailed("Bạn chưa thêm tài khoản ngân hàng")
		}
		return nil, err
	}

	summary, err := uc.walletSummary.Execute(ctx, in.DriverID)
	if err != nil {
		return nil, err
	}
	if summary.OutstandingCents > summary.AvailableCents {
		return nil, errors.PreconditionFailed("Bạn đang nợ Panda " + formatVND(summary.OutstandingCents) + ", chưa thể rút tiền")
	}
	if in.AmountCents > summary.NetCents {
		return nil, errors.PreconditionFailed("Số dư khả dụng không đủ")
	}

	now := time.Now().UTC()
	p, err := entity.NewPayoutRequest(uuid.NewString(), in.DriverID, in.AmountCents, summary.Currency, bank.BankAccountID, bank.BankName, bank.MaskedAccountNumber(), now)
	if err != nil {
		return nil, err
	}
	if err := uc.payouts.Save(ctx, p); err != nil {
		return nil, err
	}

	// Frozen Balance (critique #8): reduce Available the moment the request
	// exists, not just at Paid time — so the same money can't double-count
	// toward a second request or any other read of Available while this one
	// is in flight. Deterministic ID (derived from the payout request ID)
	// mirrors the Settlement Engine's own crash-safety pattern.
	wallet, err := uc.wallets.Execute(ctx, in.DriverID, entity.WalletTypeDriver)
	if err != nil {
		return nil, err
	}
	holdTxID := "payout-hold-" + p.PayoutRequestID
	holdTx, err := entity.NewTransaction(holdTxID, entity.TypePayoutHold, p.PayoutRequestID, "", p.Currency, "Giữ tiền chờ rút #"+p.PayoutRequestID, now)
	if err != nil {
		return nil, err
	}
	if err := uc.tx.Save(ctx, holdTx); err != nil {
		return nil, err
	}
	holdEntry, err := entity.NewLedgerEntry(holdTxID+"-entry", wallet.WalletID, holdTxID, entity.DirectionDebit, in.AmountCents, p.Currency, "Giữ tiền chờ rút #"+p.PayoutRequestID, now)
	if err != nil {
		return nil, err
	}
	if err := uc.ledger.Save(ctx, holdEntry); err != nil {
		return nil, err
	}

	_ = recordAudit(ctx, uc.audit, entity.AuditEntityPayoutRequest, p.PayoutRequestID, in.DriverID, entity.AuditActionCreate, in.DriverID, "", payoutRequestSnapshot(p), "")
	return p, nil
}

// reversePayoutHold posts the reversal credit that unfreezes a payout
// request's held amount (Reject) — see TypePayoutHold's doc comment.
func reversePayoutHold(ctx context.Context, wallets *GetOrCreateWalletUseCase, ledger repository.LedgerEntryRepository, tx repository.TransactionRepository, p *entity.PayoutRequest, now time.Time) error {
	wallet, err := wallets.Execute(ctx, p.DriverID, entity.WalletTypeDriver)
	if err != nil {
		return err
	}
	holdTxID := "payout-hold-" + p.PayoutRequestID
	releaseTxID := holdTxID + "-release"
	releaseTx, err := entity.NewTransaction(releaseTxID, entity.TypePayoutHold, p.PayoutRequestID, "", p.Currency, "Hủy giữ tiền #"+p.PayoutRequestID, now)
	if err != nil {
		return err
	}
	if err := tx.Save(ctx, releaseTx); err != nil && errors.GetCode(err) != errors.CodeAlreadyExists {
		return err
	}
	releaseEntry, err := entity.NewLedgerEntry(releaseTxID+"-entry", wallet.WalletID, releaseTxID, entity.DirectionCredit, p.AmountCents, p.Currency, "Hủy giữ tiền #"+p.PayoutRequestID, now)
	if err != nil {
		return err
	}
	if err := ledger.Save(ctx, releaseEntry); err != nil && errors.GetCode(err) != errors.CodeAlreadyExists {
		return err
	}
	return nil
}

// ListMyPayoutRequestsUseCase returns a driver's own payout history.
type ListMyPayoutRequestsUseCase struct {
	payouts repository.PayoutRequestRepository
}

func NewListMyPayoutRequestsUseCase(payouts repository.PayoutRequestRepository) *ListMyPayoutRequestsUseCase {
	return &ListMyPayoutRequestsUseCase{payouts: payouts}
}

func (uc *ListMyPayoutRequestsUseCase) Execute(ctx context.Context, driverID string, limit int) ([]*entity.PayoutRequest, error) {
	if strings.TrimSpace(driverID) == "" {
		return nil, errors.InvalidArgument("driver_id must not be empty")
	}
	return uc.payouts.ListByDriverID(ctx, driverID, limit)
}

// ─── Admin review (Phần 8/10) ────────────────────────────────────────────────

// ReviewPayoutRequestInput carries an admin's approve/reject decision.
type ReviewPayoutRequestInput struct {
	PayoutRequestID string
	Reviewer        string
	Reason          string // required for reject
}

// ApprovePayoutRequestUseCase transitions Pending -> Approved.
type ApprovePayoutRequestUseCase struct {
	payouts repository.PayoutRequestRepository
	audit   repository.AuditLogRepository
}

func NewApprovePayoutRequestUseCase(payouts repository.PayoutRequestRepository, audit repository.AuditLogRepository) *ApprovePayoutRequestUseCase {
	return &ApprovePayoutRequestUseCase{payouts: payouts, audit: audit}
}

func (uc *ApprovePayoutRequestUseCase) Execute(ctx context.Context, in ReviewPayoutRequestInput) (*entity.PayoutRequest, error) {
	p, err := uc.payouts.FindByID(ctx, in.PayoutRequestID)
	if err != nil {
		return nil, err
	}
	oldValue := payoutRequestSnapshot(p)
	now := time.Now().UTC()
	if err := p.Approve(in.Reviewer, now); err != nil {
		return nil, err
	}
	if err := uc.payouts.Save(ctx, p); err != nil {
		return nil, err
	}
	_ = recordAudit(ctx, uc.audit, entity.AuditEntityPayoutRequest, p.PayoutRequestID, p.DriverID, entity.AuditActionApprove, in.Reviewer, oldValue, payoutRequestSnapshot(p), "")
	return p, nil
}

// RejectPayoutRequestUseCase transitions Pending -> Rejected.
type RejectPayoutRequestUseCase struct {
	payouts repository.PayoutRequestRepository
	wallets *GetOrCreateWalletUseCase
	ledger  repository.LedgerEntryRepository
	tx      repository.TransactionRepository
	audit   repository.AuditLogRepository
}

func NewRejectPayoutRequestUseCase(
	payouts repository.PayoutRequestRepository,
	wallets *GetOrCreateWalletUseCase,
	ledger repository.LedgerEntryRepository,
	tx repository.TransactionRepository,
	audit repository.AuditLogRepository,
) *RejectPayoutRequestUseCase {
	return &RejectPayoutRequestUseCase{payouts: payouts, wallets: wallets, ledger: ledger, tx: tx, audit: audit}
}

func (uc *RejectPayoutRequestUseCase) Execute(ctx context.Context, in ReviewPayoutRequestInput) (*entity.PayoutRequest, error) {
	p, err := uc.payouts.FindByID(ctx, in.PayoutRequestID)
	if err != nil {
		return nil, err
	}
	oldValue := payoutRequestSnapshot(p)
	now := time.Now().UTC()
	if err := p.Reject(in.Reviewer, in.Reason, now); err != nil {
		return nil, err
	}
	if err := uc.payouts.Save(ctx, p); err != nil {
		return nil, err
	}
	// Frozen Balance: release the hold this request placed on Available.
	if err := reversePayoutHold(ctx, uc.wallets, uc.ledger, uc.tx, p, now); err != nil {
		return nil, err
	}
	_ = recordAudit(ctx, uc.audit, entity.AuditEntityPayoutRequest, p.PayoutRequestID, p.DriverID, entity.AuditActionReject, in.Reviewer, oldValue, payoutRequestSnapshot(p), in.Reason)
	return p, nil
}

// MarkPayoutPaidUseCase transitions Approved -> Paid (Phần 8 — "Không tự
// Paid": only reachable via this explicit Admin action) and is the ONLY
// place a Withdrawal ledger entry is ever created (Ledger First — the
// wallet's Available balance and Lifetime Withdrawn only change here).
type MarkPayoutPaidUseCase struct {
	payouts repository.PayoutRequestRepository
	wallets *GetOrCreateWalletUseCase
	ledger  repository.LedgerEntryRepository
	tx      repository.TransactionRepository
	audit   repository.AuditLogRepository
}

func NewMarkPayoutPaidUseCase(
	payouts repository.PayoutRequestRepository,
	wallets *GetOrCreateWalletUseCase,
	ledger repository.LedgerEntryRepository,
	tx repository.TransactionRepository,
	audit repository.AuditLogRepository,
) *MarkPayoutPaidUseCase {
	return &MarkPayoutPaidUseCase{payouts: payouts, wallets: wallets, ledger: ledger, tx: tx, audit: audit}
}

func (uc *MarkPayoutPaidUseCase) Execute(ctx context.Context, payoutRequestID, actorID string) (*entity.PayoutRequest, error) {
	p, err := uc.payouts.FindByID(ctx, payoutRequestID)
	if err != nil {
		return nil, err
	}
	oldValue := payoutRequestSnapshot(p)

	wallet, err := uc.wallets.Execute(ctx, p.DriverID, entity.WalletTypeDriver)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	txID := uuid.NewString()
	transaction, err := entity.NewTransaction(txID, entity.TypeWithdrawal, p.PayoutRequestID, "", p.Currency, "Rút tiền về "+p.BankName+" "+p.AccountNumberMasked, now)
	if err != nil {
		return nil, err
	}
	if err := uc.tx.Save(ctx, transaction); err != nil {
		return nil, err
	}
	entry, err := entity.NewLedgerEntry(uuid.NewString(), wallet.WalletID, txID, entity.DirectionDebit, p.AmountCents, p.Currency, "Rút tiền #"+p.PayoutRequestID, now)
	if err != nil {
		return nil, err
	}
	if err := uc.ledger.Save(ctx, entry); err != nil {
		return nil, err
	}
	// Frozen Balance: the hold's job is done — the Withdrawal debit above is
	// now the permanent reduction, so release the hold rather than
	// double-counting both against Available.
	if err := reversePayoutHold(ctx, uc.wallets, uc.ledger, uc.tx, p, now); err != nil {
		return nil, err
	}

	if err := p.MarkPaid(txID, now); err != nil {
		return nil, err
	}
	if err := uc.payouts.Save(ctx, p); err != nil {
		return nil, err
	}
	_ = recordAudit(ctx, uc.audit, entity.AuditEntityPayoutRequest, p.PayoutRequestID, p.DriverID, entity.AuditActionPaid, actorID, oldValue, payoutRequestSnapshot(p), "")
	return p, nil
}

// formatVND renders a whole-VND integer amount for a Vietnamese error
// message, e.g. 50000 -> "50.000 ₫". VND has no decimal subunit.
func formatVND(amount int64) string {
	s := formatThousands(amount)
	return s + " ₫"
}

func formatThousands(n int64) string {
	if n < 0 {
		return "-" + formatThousands(-n)
	}
	s := ""
	for n >= 1000 {
		s = "." + pad3(n%1000) + s
		n /= 1000
	}
	return strconv.FormatInt(n, 10) + s
}

func pad3(n int64) string {
	s := strconv.FormatInt(n, 10)
	for len(s) < 3 {
		s = "0" + s
	}
	return s
}

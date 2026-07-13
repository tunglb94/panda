package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// PayoutStatus is the Phần 5/8 withdrawal workflow state:
// Pending -> Approved -> Paid, or Pending -> Rejected.
type PayoutStatus string

const (
	PayoutPending  PayoutStatus = "pending"
	PayoutApproved PayoutStatus = "approved"
	PayoutRejected PayoutStatus = "rejected"
	PayoutPaid     PayoutStatus = "paid"
)

func validPayoutStatus(s PayoutStatus) bool {
	switch s {
	case PayoutPending, PayoutApproved, PayoutRejected, PayoutPaid:
		return true
	default:
		return false
	}
}

// PayoutRequest is a driver's withdrawal request (Phần 5). No real money
// movement ever happens (Phần 5/8 — "Không chuyển tiền thật. Chỉ mô
// phỏng.") — BankName/AccountNumberMasked are a snapshot of the bank
// account at request time, so later edits to the driver's bank account
// never rewrite payout history.
type PayoutRequest struct {
	PayoutRequestID     string
	DriverID            string
	AmountCents         int64
	Currency            string
	BankAccountID       string
	BankName            string
	AccountNumberMasked string
	Status              PayoutStatus
	RequestedAt         time.Time
	ReviewedAt          *time.Time
	ReviewedBy          string
	RejectReason        string
	PaidAt              *time.Time
	TransactionID       string // set only once Paid — links to the Withdrawal ledger transaction
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// NewPayoutRequest creates a validated PayoutRequest in PayoutPending.
// All eligibility checks (KYC, existing pending request, available vs
// outstanding, minimum amount) are the app layer's responsibility (see
// app/payout_request.go's CreatePayoutRequestUseCase) — this constructor
// only enforces structural invariants.
func NewPayoutRequest(id, driverID string, amountCents int64, currency, bankAccountID, bankName, accountNumberMasked string, now time.Time) (*PayoutRequest, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.InvalidArgument("payout_request_id must not be empty")
	}
	if strings.TrimSpace(driverID) == "" {
		return nil, errors.InvalidArgument("driver_id must not be empty")
	}
	if amountCents <= 0 {
		return nil, errors.InvalidArgument("amount_cents must be greater than zero")
	}
	if strings.TrimSpace(currency) == "" {
		return nil, errors.InvalidArgument("currency must not be empty")
	}
	if strings.TrimSpace(bankAccountID) == "" {
		return nil, errors.InvalidArgument("bank_account_id must not be empty")
	}
	return &PayoutRequest{
		PayoutRequestID:     id,
		DriverID:            driverID,
		AmountCents:         amountCents,
		Currency:            currency,
		BankAccountID:       bankAccountID,
		BankName:            bankName,
		AccountNumberMasked: accountNumberMasked,
		Status:              PayoutPending,
		RequestedAt:         now,
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

// ReconstitutePayoutRequest rebuilds a PayoutRequest from persistence without validation.
func ReconstitutePayoutRequest(
	id, driverID string,
	amountCents int64,
	currency, bankAccountID, bankName, accountNumberMasked string,
	status PayoutStatus,
	requestedAt time.Time,
	reviewedAt *time.Time,
	reviewedBy, rejectReason string,
	paidAt *time.Time,
	transactionID string,
	createdAt, updatedAt time.Time,
) *PayoutRequest {
	return &PayoutRequest{
		PayoutRequestID:     id,
		DriverID:            driverID,
		AmountCents:         amountCents,
		Currency:            currency,
		BankAccountID:       bankAccountID,
		BankName:            bankName,
		AccountNumberMasked: accountNumberMasked,
		Status:              status,
		RequestedAt:         requestedAt,
		ReviewedAt:          reviewedAt,
		ReviewedBy:          reviewedBy,
		RejectReason:        rejectReason,
		PaidAt:              paidAt,
		TransactionID:       transactionID,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
	}
}

// Approve transitions Pending -> Approved.
func (p *PayoutRequest) Approve(reviewer string, now time.Time) error {
	if p.Status != PayoutPending {
		return errors.PreconditionFailed("payout request cannot be approved from status: " + string(p.Status))
	}
	p.Status = PayoutApproved
	p.ReviewedAt = &now
	p.ReviewedBy = reviewer
	p.UpdatedAt = now
	return nil
}

// Reject transitions Pending -> Rejected. reason is required.
func (p *PayoutRequest) Reject(reviewer, reason string, now time.Time) error {
	if p.Status != PayoutPending {
		return errors.PreconditionFailed("payout request cannot be rejected from status: " + string(p.Status))
	}
	if strings.TrimSpace(reason) == "" {
		return errors.InvalidArgument("reject_reason must not be empty")
	}
	p.Status = PayoutRejected
	p.ReviewedAt = &now
	p.ReviewedBy = reviewer
	p.RejectReason = strings.TrimSpace(reason)
	p.UpdatedAt = now
	return nil
}

// MarkPaid transitions Approved -> Paid (Phần 8 — "Không tự Paid": only
// reachable via an explicit Admin action, never automatically).
// transactionID links to the Withdrawal ledger transaction created at this
// exact moment (Ledger First — see app/payout_request.go).
func (p *PayoutRequest) MarkPaid(transactionID string, now time.Time) error {
	if p.Status != PayoutApproved {
		return errors.PreconditionFailed("payout request cannot be marked paid from status: " + string(p.Status))
	}
	if strings.TrimSpace(transactionID) == "" {
		return errors.InvalidArgument("transaction_id must not be empty")
	}
	p.Status = PayoutPaid
	p.PaidAt = &now
	p.TransactionID = transactionID
	p.UpdatedAt = now
	return nil
}

// IsInFlight reports whether this request still blocks a new payout
// request from being created (Phần 5 — "Có payout đang Pending").
func (p *PayoutRequest) IsInFlight() bool {
	return p.Status == PayoutPending || p.Status == PayoutApproved
}

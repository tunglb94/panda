package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// BankAccount is a driver's single default payout destination (Phần 6 —
// "Chỉ cho 1 tài khoản mặc định"). No PIN, no OTP, no real bank
// integration — this is a static record used only to display a masked
// account number and to snapshot onto a PayoutRequest; no money actually
// moves (Phần 5/8 — "Không chuyển khoản thật. Chỉ workflow.").
type BankAccount struct {
	BankAccountID     string
	DriverID          string
	BankName          string
	AccountHolderName string
	AccountNumber     string // raw — NEVER serialized to any API response; see MaskedAccountNumber
	BranchName        string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewBankAccount creates a validated BankAccount. BranchName is optional.
func NewBankAccount(bankAccountID, driverID, bankName, accountHolderName, accountNumber, branchName string, now time.Time) (*BankAccount, error) {
	if strings.TrimSpace(bankAccountID) == "" {
		return nil, errors.InvalidArgument("bank_account_id must not be empty")
	}
	if strings.TrimSpace(driverID) == "" {
		return nil, errors.InvalidArgument("driver_id must not be empty")
	}
	if strings.TrimSpace(bankName) == "" {
		return nil, errors.InvalidArgument("bank_name must not be empty")
	}
	if strings.TrimSpace(accountHolderName) == "" {
		return nil, errors.InvalidArgument("account_holder_name must not be empty")
	}
	if strings.TrimSpace(accountNumber) == "" {
		return nil, errors.InvalidArgument("account_number must not be empty")
	}
	return &BankAccount{
		BankAccountID:     bankAccountID,
		DriverID:          driverID,
		BankName:          strings.TrimSpace(bankName),
		AccountHolderName: strings.TrimSpace(accountHolderName),
		AccountNumber:     strings.TrimSpace(accountNumber),
		BranchName:        strings.TrimSpace(branchName),
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

// ReconstituteBankAccount rebuilds a BankAccount from persistence without validation.
func ReconstituteBankAccount(bankAccountID, driverID, bankName, accountHolderName, accountNumber, branchName string, createdAt, updatedAt time.Time) *BankAccount {
	return &BankAccount{
		BankAccountID:     bankAccountID,
		DriverID:          driverID,
		BankName:          bankName,
		AccountHolderName: accountHolderName,
		AccountNumber:     accountNumber,
		BranchName:        branchName,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}
}

// Update replaces this driver's single default bank account with new
// details (Phần 6 — there is only ever one; "adding" a new account is
// really replacing the existing one).
func (b *BankAccount) Update(bankName, accountHolderName, accountNumber, branchName string, now time.Time) error {
	if strings.TrimSpace(bankName) == "" {
		return errors.InvalidArgument("bank_name must not be empty")
	}
	if strings.TrimSpace(accountHolderName) == "" {
		return errors.InvalidArgument("account_holder_name must not be empty")
	}
	if strings.TrimSpace(accountNumber) == "" {
		return errors.InvalidArgument("account_number must not be empty")
	}
	b.BankName = strings.TrimSpace(bankName)
	b.AccountHolderName = strings.TrimSpace(accountHolderName)
	b.AccountNumber = strings.TrimSpace(accountNumber)
	b.BranchName = strings.TrimSpace(branchName)
	b.UpdatedAt = now
	return nil
}

// MaskedAccountNumber returns only the last 4 digits, e.g. "••••1234"
// (Phần 10 — "Không hiện full account number"). Accounts with 4 or fewer
// characters are masked entirely rather than risk showing the whole thing.
func (b *BankAccount) MaskedAccountNumber() string {
	n := len(b.AccountNumber)
	if n <= 4 {
		return strings.Repeat("•", n)
	}
	return strings.Repeat("•", 4) + b.AccountNumber[n-4:]
}

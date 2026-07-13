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

// SetBankAccountInput carries the fields collected by Phần 6's "thêm ngân
// hàng" form. BranchName is optional.
type SetBankAccountInput struct {
	DriverID          string
	BankName          string
	AccountHolderName string
	AccountNumber     string
	BranchName        string
}

// SetBankAccountUseCase creates or replaces a driver's single default bank
// account (Phần 6 — "Chỉ cho 1 tài khoản mặc định"). Every call is audited.
type SetBankAccountUseCase struct {
	repo  repository.BankAccountRepository
	audit repository.AuditLogRepository
}

func NewSetBankAccountUseCase(repo repository.BankAccountRepository, audit repository.AuditLogRepository) *SetBankAccountUseCase {
	return &SetBankAccountUseCase{repo: repo, audit: audit}
}

func (uc *SetBankAccountUseCase) Execute(ctx context.Context, in SetBankAccountInput) (*entity.BankAccount, error) {
	if strings.TrimSpace(in.DriverID) == "" {
		return nil, errors.InvalidArgument("driver_id must not be empty")
	}
	now := time.Now().UTC()
	existing, err := uc.repo.FindByDriverID(ctx, in.DriverID)
	if err == nil {
		oldValue := bankAccountSnapshot(existing)
		if err := existing.Update(in.BankName, in.AccountHolderName, in.AccountNumber, in.BranchName, now); err != nil {
			return nil, err
		}
		if err := uc.repo.Save(ctx, existing); err != nil {
			return nil, err
		}
		_ = recordAudit(ctx, uc.audit, entity.AuditEntityBankAccount, existing.BankAccountID, in.DriverID, entity.AuditActionCreate, in.DriverID, oldValue, bankAccountSnapshot(existing), "")
		return existing, nil
	}
	if errors.GetCode(err) != errors.CodeNotFound {
		return nil, err
	}
	b, err := entity.NewBankAccount(uuid.NewString(), in.DriverID, in.BankName, in.AccountHolderName, in.AccountNumber, in.BranchName, now)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, b); err != nil {
		return nil, err
	}
	_ = recordAudit(ctx, uc.audit, entity.AuditEntityBankAccount, b.BankAccountID, in.DriverID, entity.AuditActionCreate, in.DriverID, "", bankAccountSnapshot(b), "")
	return b, nil
}

// GetBankAccountUseCase returns a driver's own bank account.
type GetBankAccountUseCase struct {
	repo repository.BankAccountRepository
}

func NewGetBankAccountUseCase(repo repository.BankAccountRepository) *GetBankAccountUseCase {
	return &GetBankAccountUseCase{repo: repo}
}

func (uc *GetBankAccountUseCase) Execute(ctx context.Context, driverID string) (*entity.BankAccount, error) {
	if strings.TrimSpace(driverID) == "" {
		return nil, errors.InvalidArgument("driver_id must not be empty")
	}
	return uc.repo.FindByDriverID(ctx, driverID)
}

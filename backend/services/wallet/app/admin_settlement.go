package app

import (
	"context"

	"github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

// ListSettlementsUseCase powers the admin Settlement List (Phần 10).
// driverID/from/to are optional filters (empty/0 means "any").
type ListSettlementsUseCase struct {
	settlements repository.SettlementRepository
}

func NewListSettlementsUseCase(settlements repository.SettlementRepository) *ListSettlementsUseCase {
	return &ListSettlementsUseCase{settlements: settlements}
}

func (uc *ListSettlementsUseCase) Execute(ctx context.Context, driverID string, from, to int64, limit int) ([]*entity.Settlement, error) {
	return uc.settlements.ListAll(ctx, driverID, from, to, limit)
}

// GetSettlementDetailUseCase powers the admin Settlement Detail view (Phần 10).
type GetSettlementDetailUseCase struct {
	settlements repository.SettlementRepository
}

func NewGetSettlementDetailUseCase(settlements repository.SettlementRepository) *GetSettlementDetailUseCase {
	return &GetSettlementDetailUseCase{settlements: settlements}
}

func (uc *GetSettlementDetailUseCase) Execute(ctx context.Context, settlementID string) (*entity.Settlement, error) {
	if settlementID == "" {
		return nil, errors.InvalidArgument("settlement_id must not be empty")
	}
	return uc.settlements.FindByID(ctx, settlementID)
}

// ListOutstandingDriversUseCase powers the admin Outstanding Drivers view (Phần 10).
type ListOutstandingDriversUseCase struct {
	ledger repository.LedgerEntryRepository
}

func NewListOutstandingDriversUseCase(ledger repository.LedgerEntryRepository) *ListOutstandingDriversUseCase {
	return &ListOutstandingDriversUseCase{ledger: ledger}
}

func (uc *ListOutstandingDriversUseCase) Execute(ctx context.Context, limit int) ([]repository.OutstandingDriver, error) {
	return uc.ledger.ListOutstandingDrivers(ctx, limit)
}

// ListPayoutRequestsUseCase powers the admin Withdrawal List/Search (Phần 10)
// — status/driverID are optional filters.
type ListPayoutRequestsUseCase struct {
	payouts repository.PayoutRequestRepository
}

func NewListPayoutRequestsUseCase(payouts repository.PayoutRequestRepository) *ListPayoutRequestsUseCase {
	return &ListPayoutRequestsUseCase{payouts: payouts}
}

func (uc *ListPayoutRequestsUseCase) Execute(ctx context.Context, status entity.PayoutStatus, driverID string, limit int) ([]*entity.PayoutRequest, error) {
	return uc.payouts.ListByFilter(ctx, status, driverID, limit)
}

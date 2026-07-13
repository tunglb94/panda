package app

import (
	"context"
	"strings"

	"github.com/fairride/shared/errors"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

// StatementLine is one row of a Driver Statement (Phần 9) — one Settlement,
// annotated with the category labels the UI groups by (Ride/Delivery,
// Cash/Electronic).
type StatementLine struct {
	SettlementID          string
	TripID                string
	TripType              entity.TripType
	PaymentMethod         entity.PaymentMethod
	FareAmountCents       int64
	CommissionAmountCents int64
	DriverIncomeCents     int64
	CreatedAt             int64 // unix seconds
}

// Statement is the Phần 9 report — per-category totals plus the line
// items, for a [from, to) window (Ngày/Tuần/Tháng — the caller picks the
// window, this use case just aggregates whatever range it's given).
type Statement struct {
	DriverID              string
	Currency              string
	RideIncomeCents       int64
	DeliveryIncomeCents   int64
	CommissionCents       int64
	PromotionCents        int64 // always 0 today — see Known Gap
	VoucherCents          int64 // always 0 today — see Known Gap
	CashIncomeCents       int64
	ElectronicIncomeCents int64
	WithdrawalCents       int64
	OutstandingCents      int64
	Lines                 []StatementLine
}

// GetStatementUseCase powers Phần 9's Driver Statement export (Ngày/Tuần/Tháng).
type GetStatementUseCase struct {
	settlements   repository.SettlementRepository
	walletSummary *GetWalletSummaryUseCase
}

func NewGetStatementUseCase(settlements repository.SettlementRepository, walletSummary *GetWalletSummaryUseCase) *GetStatementUseCase {
	return &GetStatementUseCase{settlements: settlements, walletSummary: walletSummary}
}

func (uc *GetStatementUseCase) Execute(ctx context.Context, driverID string, from, to int64) (*Statement, error) {
	if strings.TrimSpace(driverID) == "" {
		return nil, errors.InvalidArgument("driver_id must not be empty")
	}
	settlements, err := uc.settlements.ListByDriverID(ctx, driverID, from, to, 500)
	if err != nil {
		return nil, err
	}
	st := &Statement{DriverID: driverID, Currency: DefaultCurrency}
	for _, s := range settlements {
		st.Currency = s.Currency
		st.CommissionCents += s.CommissionAmountCents
		st.PromotionCents += s.PromotionSubsidyCents
		st.VoucherCents += s.VoucherCostCents
		if s.TripType == entity.TripTypeDelivery {
			st.DeliveryIncomeCents += s.DriverIncomeCents
		} else {
			st.RideIncomeCents += s.DriverIncomeCents
		}
		if s.PaymentMethod == entity.PaymentMethodCash {
			st.CashIncomeCents += s.DriverIncomeCents
		} else {
			st.ElectronicIncomeCents += s.DriverIncomeCents
		}
		st.Lines = append(st.Lines, StatementLine{
			SettlementID: s.SettlementID, TripID: s.TripID, TripType: s.TripType, PaymentMethod: s.PaymentMethod,
			FareAmountCents: s.FareAmountCents, CommissionAmountCents: s.CommissionAmountCents,
			DriverIncomeCents: s.DriverIncomeCents, CreatedAt: s.CreatedAt.Unix(),
		})
	}
	summary, err := uc.walletSummary.Execute(ctx, driverID)
	if err == nil {
		st.OutstandingCents = summary.OutstandingCents
		// Known Gap: Withdrawal here is the driver's lifetime total, not
		// filtered to [from, to) — PayoutRequestRepository has no
		// date-range query yet (Phần 9's per-period Withdrawal figure).
		st.WithdrawalCents = summary.LifetimeWithdrawnCents
	}
	return st, nil
}

package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/domain/repository"
)

// recordAudit is the shared helper behind Phần 12 — "Adjustment phải có
// Audit Log". Every mutating use case in this package calls this after a
// successful state change. Its error (if any) is returned like any other
// repository failure — an action isn't considered complete until its audit
// entry is durably recorded.
func recordAudit(ctx context.Context, repo repository.AuditLogRepository, entityType entity.AuditEntityType, entityID, driverID string, action entity.AuditAction, actorID, oldValue, newValue, reason string) error {
	if repo == nil {
		return nil
	}
	log := entity.NewAuditLog(uuid.NewString(), entityType, entityID, driverID, action, actorID, oldValue, newValue, reason, time.Now().UTC())
	return repo.Save(ctx, log)
}

func payoutRequestSnapshot(p *entity.PayoutRequest) string {
	if p == nil {
		return ""
	}
	b, _ := json.Marshal(map[string]any{
		"amount_cents":  p.AmountCents,
		"status":        string(p.Status),
		"reject_reason": p.RejectReason,
	})
	return string(b)
}

func settlementSnapshot(s *entity.Settlement) string {
	if s == nil {
		return ""
	}
	b, _ := json.Marshal(map[string]any{
		"trip_id":        s.TripID,
		"payment_method": string(s.PaymentMethod),
		"fare_amount":    s.FareAmountCents,
		"commission":     s.CommissionAmountCents,
		"driver_income":  s.DriverIncomeCents,
	})
	return string(b)
}

func bankAccountSnapshot(b *entity.BankAccount) string {
	if b == nil {
		return ""
	}
	bytes, _ := json.Marshal(map[string]any{
		"bank_name":             b.BankName,
		"account_holder_name":   b.AccountHolderName,
		"masked_account_number": b.MaskedAccountNumber(),
	})
	return string(bytes)
}

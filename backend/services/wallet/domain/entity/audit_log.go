package entity

import "time"

// AuditAction names the mutating operations Phần 12 requires a trail for
// ("Adjustment phải có Audit Log").
type AuditAction string

const (
	AuditActionCreate           AuditAction = "create"
	AuditActionApprove          AuditAction = "approve"
	AuditActionReject           AuditAction = "reject"
	AuditActionPaid             AuditAction = "paid"
	AuditActionManualAdjustment AuditAction = "manual_adjustment"
)

// AuditEntityType names which aggregate an AuditLog entry describes.
type AuditEntityType string

const (
	AuditEntitySettlement    AuditEntityType = "settlement"
	AuditEntityPayoutRequest AuditEntityType = "payout_request"
	AuditEntityBankAccount   AuditEntityType = "bank_account"
	AuditEntityLedgerEntry   AuditEntityType = "ledger_entry"
)

// AuditLog is one append-only record of a wallet/finance mutation (Phần 12
// — "Settlement immutable. Adjustment phải có Audit Log."). Nothing in this
// codebase ever updates or deletes an AuditLog row once written —
// repository.AuditLogRepository only exposes Save (insert) and List.
type AuditLog struct {
	ID         string
	EntityType AuditEntityType
	EntityID   string
	DriverID   string
	Action     AuditAction
	ActorID    string // driver_id, admin user id, or "system"
	OldValue   string
	NewValue   string
	Reason     string
	CreatedAt  time.Time
}

// NewAuditLog builds an AuditLog entry. No validation beyond requiring the
// identifying fields — this is a factual record of something that already
// happened, not a domain invariant to protect.
func NewAuditLog(id string, entityType AuditEntityType, entityID, driverID string, action AuditAction, actorID, oldValue, newValue, reason string, now time.Time) *AuditLog {
	return &AuditLog{
		ID:         id,
		EntityType: entityType,
		EntityID:   entityID,
		DriverID:   driverID,
		Action:     action,
		ActorID:    actorID,
		OldValue:   oldValue,
		NewValue:   newValue,
		Reason:     reason,
		CreatedAt:  now,
	}
}

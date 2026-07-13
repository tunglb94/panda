package entity

import "time"

// AuditAction names the mutating operations Phần 7 requires a trail for.
type AuditAction string

const (
	AuditActionSubmit  AuditAction = "submit"
	AuditActionModify  AuditAction = "modify"
	AuditActionApprove AuditAction = "approve"
	AuditActionReject  AuditAction = "reject"
	AuditActionExpire  AuditAction = "expire"
)

// AuditEntityType names which aggregate an AuditLog entry describes.
type AuditEntityType string

const (
	AuditEntityDriverVerification  AuditEntityType = "driver_verification"
	AuditEntityVehicleVerification AuditEntityType = "vehicle_verification"
	AuditEntityKYCDocument         AuditEntityType = "kyc_document"
)

// AuditLog is one append-only record of a KYC-related mutation — Phần 7:
// "Không được mất lịch sử". Nothing in this codebase ever updates or
// deletes an AuditLog row once written; repository.AuditLogRepository only
// exposes Save (insert) and List methods.
type AuditLog struct {
	ID         string
	EntityType AuditEntityType
	EntityID   string
	DriverID   string
	Action     AuditAction
	ActorID    string // driver_id, admin user id, or "system" for auto-expiry
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

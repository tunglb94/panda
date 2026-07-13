package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
)

// recordAudit is the shared helper behind Phần 7 — every mutating KYC use
// case calls this after a successful state change, so "Không được mất
// lịch sử" holds for Submit/Approve/Reject/Expire/Modify alike. Its error
// (if any) is returned like any other repository failure — an action isn't
// considered complete until its audit entry is durably recorded.
func recordAudit(ctx context.Context, repo repository.AuditLogRepository, entityType entity.AuditEntityType, entityID, driverID string, action entity.AuditAction, actorID, oldValue, newValue, reason string) error {
	if repo == nil {
		return nil
	}
	log := entity.NewAuditLog(uuid.NewString(), entityType, entityID, driverID, action, actorID, oldValue, newValue, reason, time.Now().UTC())
	return repo.Save(ctx, log)
}

func driverVerificationSnapshot(v *entity.DriverVerification) string {
	if v == nil {
		return ""
	}
	b, _ := json.Marshal(map[string]any{
		"full_name":          v.FullName,
		"address":            v.Address,
		"national_id_number": v.NationalIDNumber,
		"license_number":     v.LicenseNumber,
		"status":             string(v.Status),
	})
	return string(b)
}

func vehicleVerificationSnapshot(v *entity.VehicleVerification) string {
	if v == nil {
		return ""
	}
	b, _ := json.Marshal(map[string]any{
		"vehicle_type":     string(v.VehicleType),
		"service_type":     string(v.ServiceType),
		"plate_number":     v.PlateNumber,
		"vin":              v.VIN,
		"engine_number":    v.EngineNumber,
		"chassis_number":   v.ChassisNumber,
		"license_class":    string(v.LicenseClass),
		"ride_enabled":     v.RideEnabled,
		"delivery_enabled": v.DeliveryEnabled,
		"status":           string(v.Status),
	})
	return string(b)
}

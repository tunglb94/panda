package entity_test

import (
	"testing"
	"time"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/shared/errors"
)

var vvNow = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func TestNewVehicleVerification_OK_Bike(t *testing.T) {
	v, err := entity.NewVehicleVerification(
		"id1", "d1", entity.VehicleTypeMotorcycle, entity.ServiceTypeBike,
		"Honda", "Wave", 2022, "Đỏ", "59H1-12345", "", "", "",
		entity.LicenseClassA1, true, false, vvNow,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Status != entity.KYCPending {
		t.Errorf("status = %v, want pending", v.Status)
	}
}

func TestNewVehicleVerification_RejectsMismatchedVehicleAndServiceType(t *testing.T) {
	// ServiceTypeCarXL requires VehicleTypeVan, not VehicleTypeCar.
	_, err := entity.NewVehicleVerification(
		"id1", "d1", entity.VehicleTypeCar, entity.ServiceTypeCarXL,
		"Toyota", "Vios", 2022, "Trắng", "59H1-99999", "", "", "",
		entity.LicenseClassB1, true, false, vvNow,
	)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument for vehicle/service mismatch, got %v", err)
	}
}

func TestNewVehicleVerification_RequiresAtLeastOneCapability(t *testing.T) {
	_, err := entity.NewVehicleVerification(
		"id1", "d1", entity.VehicleTypeMotorcycle, entity.ServiceTypeBike,
		"Honda", "Wave", 2022, "Đỏ", "59H1-12345", "", "", "",
		"", false, false, vvNow,
	)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument when neither ride nor delivery enabled, got %v", err)
	}
}

func TestNewVehicleVerification_DeliveryOnlyDoesNotRequireLicenseClass(t *testing.T) {
	// Phần 2: "Delivery không dựa vào GPLX mà dựa vào VehicleType" — empty
	// LicenseClass must be accepted when only DeliveryEnabled is true.
	v, err := entity.NewVehicleVerification(
		"id1", "d1", entity.VehicleTypeMotorcycle, entity.ServiceTypeBike,
		"Honda", "Wave", 2022, "Đỏ", "59H1-12345", "", "", "",
		"", false, true, vvNow,
	)
	if err != nil {
		t.Fatalf("delivery-only vehicle should not require a license class: %v", err)
	}
	if v.HasPermission(entity.PermissionRideBike) {
		t.Error("delivery-only vehicle must not carry a ride permission")
	}
	if !v.HasPermission(entity.PermissionDeliveryBike) {
		t.Error("delivery-only vehicle should carry delivery_bike permission")
	}
}

func TestNewVehicleVerification_RideRequiresValidLicenseClass(t *testing.T) {
	_, err := entity.NewVehicleVerification(
		"id1", "d1", entity.VehicleTypeMotorcycle, entity.ServiceTypeBike,
		"Honda", "Wave", 2022, "Đỏ", "59H1-12345", "", "", "",
		"", true, false, vvNow,
	)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument for ride-enabled with no license class, got %v", err)
	}
}

// Whether a LicenseClass actually permits a ServiceType (Phần 1's Rule
// Engine) is no longer an entity-level concern — it's DB-backed and
// checked at the app layer (see app.checkLicenseRule and its tests). This
// entity only validates that LicenseClass is a well-formed value.
func TestNewVehicleVerification_RejectsUnknownLicenseClassWhenRideEnabled(t *testing.T) {
	_, err := entity.NewVehicleVerification(
		"id1", "d1", entity.VehicleTypeMotorcycle, entity.ServiceTypeBike,
		"Honda", "Wave", 2022, "Đỏ", "59H1-12345", "", "", "",
		"Z9", true, false, vvNow,
	)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument for unknown license class, got %v", err)
	}
}

// ─── Phần 6 — vehicle identity fields are optional and carried through ────

func TestNewVehicleVerification_OptionalIdentityFieldsCarried(t *testing.T) {
	v, err := entity.NewVehicleVerification(
		"id1", "d1", entity.VehicleTypeMotorcycle, entity.ServiceTypeBike,
		"Honda", "Wave", 2022, "Đỏ", "59H1-12345", "VIN123", "ENG456", "CHASSIS789",
		entity.LicenseClassA1, true, false, vvNow,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.VIN != "VIN123" || v.EngineNumber != "ENG456" || v.ChassisNumber != "CHASSIS789" {
		t.Errorf("vehicle identity fields not carried through: %+v", v)
	}
}

// ─── lifecycle ──────────────────────────────────────────────────────────────

func TestVehicleVerification_ApproveRejectLifecycle(t *testing.T) {
	v, _ := entity.NewVehicleVerification(
		"id1", "d1", entity.VehicleTypeMotorcycle, entity.ServiceTypeBike,
		"Honda", "Wave", 2022, "Đỏ", "59H1-12345", "", "", "",
		entity.LicenseClassA1, true, false, vvNow,
	)
	if err := v.Reject("admin1", "Biển số không rõ", vvNow); err != nil {
		t.Fatalf("reject should succeed: %v", err)
	}
	if v.Status != entity.KYCRejected {
		t.Errorf("status = %v, want rejected", v.Status)
	}

	if err := v.Resubmit(
		entity.VehicleTypeMotorcycle, entity.ServiceTypeBike, "Honda", "Wave", 2022, "Đỏ",
		"59H1-12345", "", "", "", entity.LicenseClassA1, true, false, vvNow.Add(time.Hour),
	); err != nil {
		t.Fatalf("resubmit should succeed: %v", err)
	}
	if v.Status != entity.KYCPending {
		t.Errorf("status = %v, want pending after resubmit", v.Status)
	}

	if err := v.Approve("admin2", vvNow.Add(2*time.Hour)); err != nil {
		t.Fatalf("approve should succeed: %v", err)
	}
	if !v.IsApproved() {
		t.Error("should be approved")
	}
}

func TestVehicleVerification_CannotResubmitDirectlyFromApproved(t *testing.T) {
	v, _ := entity.NewVehicleVerification(
		"id1", "d1", entity.VehicleTypeMotorcycle, entity.ServiceTypeBike,
		"Honda", "Wave", 2022, "Đỏ", "59H1-12345", "", "", "",
		entity.LicenseClassA1, true, false, vvNow,
	)
	_ = v.Approve("admin1", vvNow)

	err := v.Resubmit(entity.VehicleTypeMotorcycle, entity.ServiceTypeBike, "Honda", "Wave", 2022, "Đỏ",
		"59H1-12345", "", "", "", entity.LicenseClassA1, true, false, vvNow.Add(time.Hour))
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("Resubmit must require Invalidate first from Approved, got %v", err)
	}
}

func TestVehicleVerification_InvalidateResetsApprovedToPending(t *testing.T) {
	v, _ := entity.NewVehicleVerification(
		"id1", "d1", entity.VehicleTypeMotorcycle, entity.ServiceTypeBike,
		"Honda", "Wave", 2022, "Đỏ", "59H1-12345", "", "", "",
		entity.LicenseClassA1, true, false, vvNow,
	)
	_ = v.Approve("admin1", vvNow)

	v.Invalidate(vvNow.Add(time.Hour))
	if v.Status != entity.KYCPending {
		t.Errorf("status = %v, want pending after Invalidate", v.Status)
	}
	if v.ApprovedAt != nil {
		t.Error("Invalidate must clear ApprovedAt")
	}
}

func TestVehicleVerification_ExpireOnlyFromApproved(t *testing.T) {
	v, _ := entity.NewVehicleVerification(
		"id1", "d1", entity.VehicleTypeMotorcycle, entity.ServiceTypeBike,
		"Honda", "Wave", 2022, "Đỏ", "59H1-12345", "", "", "",
		entity.LicenseClassA1, true, false, vvNow,
	)
	if err := v.Expire("GPLX đã hết hạn", vvNow); !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("expiring a non-approved verification should fail, got %v", err)
	}
	_ = v.Approve("admin1", vvNow)
	if err := v.Expire("GPLX đã hết hạn", vvNow.Add(time.Hour)); err != nil {
		t.Fatalf("expire from approved should succeed: %v", err)
	}
	if v.Status != entity.KYCExpired {
		t.Errorf("status = %v, want expired", v.Status)
	}
}

// ─── Phần 8 — ServicePermissions ────────────────────────────────────────────

func TestVehicleVerification_PermissionsRideAndDeliveryBoth(t *testing.T) {
	v, _ := entity.NewVehicleVerification(
		"id1", "d1", entity.VehicleTypeMotorcycle, entity.ServiceTypeBikePlus,
		"Honda", "SH", 2022, "Đen", "59H1-12345", "", "", "",
		entity.LicenseClassA2, true, true, vvNow,
	)
	perms := v.Permissions()
	if len(perms) != 2 {
		t.Fatalf("expected 2 permissions, got %v", perms)
	}
	if !v.HasPermission(entity.PermissionRideBikePlus) || !v.HasPermission(entity.PermissionDeliveryBikePlus) {
		t.Errorf("permissions = %v, want ride_bike_plus and delivery_bike_plus", perms)
	}
	if !v.HasAnyPermission() {
		t.Error("HasAnyPermission should be true")
	}
}

func TestVehicleVerification_CarXLPermissions(t *testing.T) {
	v, _ := entity.NewVehicleVerification(
		"id1", "d1", entity.VehicleTypeVan, entity.ServiceTypeCarXL,
		"Ford", "Transit", 2022, "Trắng", "59H1-77777", "", "", "",
		entity.LicenseClassB1, true, false, vvNow,
	)
	if !v.HasPermission(entity.PermissionRideCarXL) {
		t.Error("expected ride_car_xl permission")
	}
	if v.HasPermission(entity.PermissionRideCar) {
		t.Error("must not carry ride_car (different tier) permission")
	}
}

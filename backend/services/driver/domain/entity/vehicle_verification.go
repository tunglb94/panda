package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// ServicePermission is a granted capability derived from
// (RideEnabled|DeliveryEnabled) × ServiceType — Phần 8 of the Driver KYC
// Hardening spec: "Thay RideEnabled/DeliveryEnabled bằng ServicePermissions".
// RideEnabled/DeliveryEnabled remain the persisted source of truth
// (backward compatible with the existing schema/API); ServicePermission is
// a read-only derived view that the Online Guard and any other consumer
// should check instead of the raw booleans directly.
type ServicePermission string

const (
	PermissionRideBike         ServicePermission = "ride_bike"
	PermissionRideBikePlus     ServicePermission = "ride_bike_plus"
	PermissionRideCar          ServicePermission = "ride_car"
	PermissionRideCarXL        ServicePermission = "ride_car_xl"
	PermissionDeliveryBike     ServicePermission = "delivery_bike"
	PermissionDeliveryBikePlus ServicePermission = "delivery_bike_plus"
	PermissionDeliveryCar      ServicePermission = "delivery_car"
	PermissionDeliveryCarXL    ServicePermission = "delivery_car_xl"
)

func serviceTypePermissionSuffix(st ServiceType) string {
	switch st {
	case ServiceTypeBike:
		return "bike"
	case ServiceTypeBikePlus:
		return "bike_plus"
	case ServiceTypeCar:
		return "car"
	case ServiceTypeCarXL:
		return "car_xl"
	default:
		return string(st)
	}
}

// VehicleVerification is the KYC record for a driver's vehicle — make/
// model/year/color/plate, which ServiceType it's being registered for, and
// (for Ride capability) the driver's license class. One per driver
// (driver_id unique) — a driver with two vehicles registers the one they
// intend to drive with Panda, matching DriverProfile's existing one-vehicle
// model (VehicleType/PlateNumber already there).
//
// VIN/EngineNumber/ChassisNumber (Phần 6) are optional vehicle-identity
// fields — never required, but unique across vehicles when present (see
// the app layer's duplicate checks).
type VehicleVerification struct {
	ID              string
	DriverID        string
	VehicleType     VehicleType
	ServiceType     ServiceType
	Brand           string
	Model           string
	Year            int
	Color           string
	PlateNumber     string
	VIN             string
	EngineNumber    string
	ChassisNumber   string
	LicenseClass    LicenseClass // required only when RideEnabled
	RideEnabled     bool
	DeliveryEnabled bool
	Status          KYCStatus
	SubmittedAt     time.Time
	ApprovedAt      *time.Time
	RejectedAt      *time.Time
	ExpiredAt       *time.Time
	Reviewer        string
	RejectReason    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewVehicleVerification creates a validated VehicleVerification in
// KYCPending. Documents (registration, insurance) are validated as already
// uploaded by the use case layer, not here. Whether LicenseClass actually
// permits ServiceType (Phần 1's Rule Engine) is a DB-backed business rule,
// checked by the app layer (which has repository access), not here — this
// constructor only validates that licenseClass is a well-formed class value
// when RideEnabled.
func NewVehicleVerification(
	id, driverID string,
	vehicleType VehicleType,
	serviceType ServiceType,
	brand, model string,
	year int,
	color, plateNumber, vin, engineNumber, chassisNumber string,
	licenseClass LicenseClass,
	rideEnabled, deliveryEnabled bool,
	now time.Time,
) (*VehicleVerification, error) {
	if id == "" {
		return nil, errors.InvalidArgument("id must not be empty")
	}
	if driverID == "" {
		return nil, errors.InvalidArgument("driver_id must not be empty")
	}
	if err := validateVehicleType(vehicleType); err != nil {
		return nil, err
	}
	if err := validateServiceType(serviceType); err != nil {
		return nil, err
	}
	if serviceType.RequiredVehicleType() != vehicleType {
		return nil, errors.InvalidArgument(
			"service type " + string(serviceType) + " requires vehicle type " +
				string(serviceType.RequiredVehicleType()) + ", got " + string(vehicleType))
	}
	if strings.TrimSpace(plateNumber) == "" {
		return nil, errors.InvalidArgument("plate_number must not be empty")
	}
	if !rideEnabled && !deliveryEnabled {
		return nil, errors.InvalidArgument("at least one of ride_enabled/delivery_enabled must be true")
	}
	if rideEnabled {
		if err := validateLicenseClass(licenseClass); err != nil {
			return nil, err
		}
	}
	return &VehicleVerification{
		ID:              id,
		DriverID:        driverID,
		VehicleType:     vehicleType,
		ServiceType:     serviceType,
		Brand:           strings.TrimSpace(brand),
		Model:           strings.TrimSpace(model),
		Year:            year,
		Color:           strings.TrimSpace(color),
		PlateNumber:     strings.TrimSpace(plateNumber),
		VIN:             strings.TrimSpace(vin),
		EngineNumber:    strings.TrimSpace(engineNumber),
		ChassisNumber:   strings.TrimSpace(chassisNumber),
		LicenseClass:    licenseClass,
		RideEnabled:     rideEnabled,
		DeliveryEnabled: deliveryEnabled,
		Status:          KYCPending,
		SubmittedAt:     now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// ReconstituteVehicleVerification rebuilds a VehicleVerification from a persistence record. No validation.
func ReconstituteVehicleVerification(
	id, driverID string,
	vehicleType VehicleType,
	serviceType ServiceType,
	brand, model string,
	year int,
	color, plateNumber, vin, engineNumber, chassisNumber string,
	licenseClass LicenseClass,
	rideEnabled, deliveryEnabled bool,
	status KYCStatus,
	submittedAt time.Time,
	approvedAt, rejectedAt, expiredAt *time.Time,
	reviewer, rejectReason string,
	createdAt, updatedAt time.Time,
) *VehicleVerification {
	return &VehicleVerification{
		ID:              id,
		DriverID:        driverID,
		VehicleType:     vehicleType,
		ServiceType:     serviceType,
		Brand:           brand,
		Model:           model,
		Year:            year,
		Color:           color,
		PlateNumber:     plateNumber,
		VIN:             vin,
		EngineNumber:    engineNumber,
		ChassisNumber:   chassisNumber,
		LicenseClass:    licenseClass,
		RideEnabled:     rideEnabled,
		DeliveryEnabled: deliveryEnabled,
		Status:          status,
		SubmittedAt:     submittedAt,
		ApprovedAt:      approvedAt,
		RejectedAt:      rejectedAt,
		ExpiredAt:       expiredAt,
		Reviewer:        reviewer,
		RejectReason:    rejectReason,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
}

// Resubmit lets a driver edit and resubmit a Pending or Rejected
// verification. Same validation as NewVehicleVerification. To edit an
// Approved or Expired verification, callers must call Invalidate first
// (Phần 3 — Re-verification): this method deliberately does not allow
// Approved -> Pending directly, so "edit resets approval" is always an
// explicit two-step the app layer controls, never implicit.
func (v *VehicleVerification) Resubmit(
	vehicleType VehicleType,
	serviceType ServiceType,
	brand, model string,
	year int,
	color, plateNumber, vin, engineNumber, chassisNumber string,
	licenseClass LicenseClass,
	rideEnabled, deliveryEnabled bool,
	now time.Time,
) error {
	if v.Status != KYCPending && v.Status != KYCRejected {
		return errors.PreconditionFailed("vehicle verification cannot be edited from status: " + string(v.Status))
	}
	fresh, err := NewVehicleVerification(v.ID, v.DriverID, vehicleType, serviceType, brand, model, year, color, plateNumber, vin, engineNumber, chassisNumber, licenseClass, rideEnabled, deliveryEnabled, now)
	if err != nil {
		return err
	}
	fresh.CreatedAt = v.CreatedAt
	*v = *fresh
	return nil
}

// Invalidate resets an Approved or Expired verification back to Pending —
// Phần 3 (Re-verification): "Không được giữ trạng thái Approved" whenever
// the driver changes vehicle info or re-uploads a document. A no-op from
// any other status (Pending/Rejected/UnderReview have nothing to
// invalidate). Callers that want to also apply new field values should call
// Invalidate first, then Resubmit (now valid, since status becomes Pending).
func (v *VehicleVerification) Invalidate(now time.Time) {
	if v.Status != KYCApproved && v.Status != KYCExpired {
		return
	}
	v.Status = KYCPending
	v.SubmittedAt = now
	v.ApprovedAt = nil
	v.RejectedAt = nil
	v.ExpiredAt = nil
	v.RejectReason = ""
	v.UpdatedAt = now
}

// StartReview transitions Pending -> UnderReview.
func (v *VehicleVerification) StartReview(reviewer string, now time.Time) error {
	if v.Status != KYCPending {
		return errors.PreconditionFailed("only pending verifications can start review")
	}
	v.Status = KYCUnderReview
	v.Reviewer = reviewer
	v.UpdatedAt = now
	return nil
}

// Approve transitions Pending/UnderReview -> Approved.
func (v *VehicleVerification) Approve(reviewer string, now time.Time) error {
	if v.Status != KYCPending && v.Status != KYCUnderReview {
		return errors.PreconditionFailed("vehicle verification cannot be approved from status: " + string(v.Status))
	}
	v.Status = KYCApproved
	v.ApprovedAt = &now
	v.ExpiredAt = nil
	v.Reviewer = reviewer
	v.RejectReason = ""
	v.UpdatedAt = now
	return nil
}

// Reject transitions Pending/UnderReview -> Rejected. reason is required.
func (v *VehicleVerification) Reject(reviewer, reason string, now time.Time) error {
	if v.Status != KYCPending && v.Status != KYCUnderReview {
		return errors.PreconditionFailed("vehicle verification cannot be rejected from status: " + string(v.Status))
	}
	if strings.TrimSpace(reason) == "" {
		return errors.InvalidArgument("reject_reason must not be empty")
	}
	v.Status = KYCRejected
	v.RejectedAt = &now
	v.Reviewer = reviewer
	v.RejectReason = strings.TrimSpace(reason)
	v.UpdatedAt = now
	return nil
}

// Expire transitions Approved -> Expired (Phần 2 — a document backing this
// verification passed its expiry date). Reason is a human-readable note
// (e.g. "GPLX đã hết hạn") recorded for the driver/admin, not a rejection.
func (v *VehicleVerification) Expire(reason string, now time.Time) error {
	if v.Status != KYCApproved {
		return errors.PreconditionFailed("only an approved vehicle verification can expire, current status: " + string(v.Status))
	}
	v.Status = KYCExpired
	v.ExpiredAt = &now
	v.RejectReason = strings.TrimSpace(reason)
	v.UpdatedAt = now
	return nil
}

func (v *VehicleVerification) IsApproved() bool { return v.Status == KYCApproved }

// Permissions returns every ServicePermission this vehicle verification
// currently grants (Phần 8). Derived, not persisted.
func (v *VehicleVerification) Permissions() []ServicePermission {
	var out []ServicePermission
	suffix := serviceTypePermissionSuffix(v.ServiceType)
	if v.RideEnabled {
		out = append(out, ServicePermission("ride_"+suffix))
	}
	if v.DeliveryEnabled {
		out = append(out, ServicePermission("delivery_"+suffix))
	}
	return out
}

// HasPermission reports whether p is among Permissions().
func (v *VehicleVerification) HasPermission(p ServicePermission) bool {
	for _, got := range v.Permissions() {
		if got == p {
			return true
		}
	}
	return false
}

// HasAnyPermission reports whether this vehicle is granted at least one
// ServicePermission — should always be true for any successfully-submitted
// verification (NewVehicleVerification enforces at-least-one-capability),
// re-checked defensively by the Online Guard (Phần 9).
func (v *VehicleVerification) HasAnyPermission() bool { return len(v.Permissions()) > 0 }

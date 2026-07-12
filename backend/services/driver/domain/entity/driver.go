// Package entity defines the Driver Profile domain model for the FAIRRIDE driver service.
package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// VehicleType classifies the vehicle a driver operates.
type VehicleType string

const (
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeMotorcycle VehicleType = "motorcycle"
	VehicleTypeVan        VehicleType = "van"
)

// ServiceType classifies the product/service tier a driver offers — a
// dimension separate from VehicleType (the physical vehicle). A Honda Wave
// and a Honda SH are both VehicleTypeMotorcycle, but the SH might operate
// as ServiceTypeBikePlus while the Wave operates as ServiceTypeBike — same
// vehicle category, different service tier. ServiceType applies to both
// Ride and Delivery trips alike (TripType is the orthogonal dimension that
// already exists on Trip/Dispatch); there is deliberately no
// "delivery_bike"/"delivery_car" ServiceType — encoding "delivery" in both
// TripType and ServiceType would duplicate the same information.
type ServiceType string

const (
	// ServiceTypeBike and ServiceTypeCar reuse VehicleTypeMotorcycle's and
	// VehicleTypeCar's exact wire values (product-facing alias, not a
	// rename) — a motorcycle driver is automatically Bike-eligible with no
	// re-registration.
	ServiceTypeBike     ServiceType = "motorcycle"
	ServiceTypeBikePlus ServiceType = "bike_plus"
	ServiceTypeCar      ServiceType = "car"
	ServiceTypeCarXL    ServiceType = "car_xl"
)

// RequiredVehicleType reports which physical VehicleType a driver must have
// to legitimately offer this ServiceType — Bike/Bike Plus need a
// motorcycle, Car needs a car, Car XL needs a van (BRB's own XL fare tier
// is the van rate — see fare.go's DefaultFareConfig doc comment). Used to
// validate a driver's (VehicleType, ServiceType) combination is coherent,
// e.g. rejecting ServiceType=car_xl on a motorcycle profile.
func (s ServiceType) RequiredVehicleType() VehicleType {
	switch s {
	case ServiceTypeBike, ServiceTypeBikePlus:
		return VehicleTypeMotorcycle
	case ServiceTypeCarXL:
		return VehicleTypeVan
	default:
		return VehicleTypeCar
	}
}

func validateServiceType(s ServiceType) error {
	switch s {
	case ServiceTypeBike, ServiceTypeBikePlus, ServiceTypeCar, ServiceTypeCarXL:
		return nil
	default:
		return errors.InvalidArgument("unknown service type: " + string(s))
	}
}

// OnlineStatus tracks whether a driver is currently accepting trips.
type OnlineStatus string

const (
	OnlineStatusOffline OnlineStatus = "offline"
	OnlineStatusOnline  OnlineStatus = "online"
)

// VerificationStatus tracks the driver approval lifecycle.
// Transitions:
//
//	Pending   → Verified  (admin approves)
//	Pending   → Rejected  (admin rejects)
//	Verified  → Suspended (admin suspends)
//	Suspended → Verified  (admin reinstates)
type VerificationStatus string

const (
	VerificationStatusPending   VerificationStatus = "pending"
	VerificationStatusVerified  VerificationStatus = "verified"
	VerificationStatusRejected  VerificationStatus = "rejected"
	VerificationStatusSuspended VerificationStatus = "suspended"
)

// DriverProfile holds the vehicle, license, and status data for a FAIRRIDE driver.
// DriverID is the service-local primary key; UserID links to the Identity/User services.
type DriverProfile struct {
	DriverID           string
	UserID             string
	LicenseNumber      string
	VehicleType        VehicleType
	VehicleBrand       string
	VehicleModel       string
	VehicleColor       string
	PlateNumber        string
	OnlineStatus       OnlineStatus
	VerificationStatus VerificationStatus
	CreatedAt          time.Time
	UpdatedAt          time.Time

	// ServiceType/RideEnabled/DeliveryEnabled are set independently of
	// vehicle registration, via SetServiceCapability — see that method's
	// doc comment. ServiceType is "" (unset) until the driver declares one.
	ServiceType     ServiceType
	RideEnabled     bool
	DeliveryEnabled bool
}

// NewDriverProfile creates a new DriverProfile.
// New drivers start as VerificationStatusPending and OnlineStatusOffline.
// driverID, userID, licenseNumber, vehicleType, and plateNumber are required.
// vehicleBrand, vehicleModel, vehicleColor are optional (empty allowed).
func NewDriverProfile(
	driverID, userID, licenseNumber string,
	vehicleType VehicleType,
	vehicleBrand, vehicleModel, vehicleColor, plateNumber string,
	now time.Time,
) (*DriverProfile, error) {
	if driverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	if userID == "" {
		return nil, errors.InvalidArgument("user id must not be empty")
	}
	if strings.TrimSpace(licenseNumber) == "" {
		return nil, errors.InvalidArgument("license number must not be empty")
	}
	if err := validateVehicleType(vehicleType); err != nil {
		return nil, err
	}
	if strings.TrimSpace(plateNumber) == "" {
		return nil, errors.InvalidArgument("plate number must not be empty")
	}
	return &DriverProfile{
		DriverID:           driverID,
		UserID:             userID,
		LicenseNumber:      licenseNumber,
		VehicleType:        vehicleType,
		VehicleBrand:       vehicleBrand,
		VehicleModel:       vehicleModel,
		VehicleColor:       vehicleColor,
		PlateNumber:        plateNumber,
		OnlineStatus:       OnlineStatusOffline,
		VerificationStatus: VerificationStatusPending,
		CreatedAt:          now,
		UpdatedAt:          now,
		// Matches migration 008's column defaults: every driver keeps
		// working for Ride exactly as before this field existed; Delivery
		// capability is new and opt-in, never auto-enabled.
		RideEnabled:     true,
		DeliveryEnabled: false,
	}, nil
}

// ReconstituteDriverProfile rebuilds a DriverProfile from a persistence record.
// No validation is applied — data is assumed already valid.
func ReconstituteDriverProfile(
	driverID, userID, licenseNumber string,
	vehicleType VehicleType,
	vehicleBrand, vehicleModel, vehicleColor, plateNumber string,
	onlineStatus OnlineStatus,
	verificationStatus VerificationStatus,
	createdAt, updatedAt time.Time,
	serviceType ServiceType,
	rideEnabled, deliveryEnabled bool,
) *DriverProfile {
	return &DriverProfile{
		DriverID:           driverID,
		UserID:             userID,
		LicenseNumber:      licenseNumber,
		VehicleType:        vehicleType,
		VehicleBrand:       vehicleBrand,
		VehicleModel:       vehicleModel,
		VehicleColor:       vehicleColor,
		PlateNumber:        plateNumber,
		OnlineStatus:       onlineStatus,
		VerificationStatus: verificationStatus,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		ServiceType:        serviceType,
		RideEnabled:        rideEnabled,
		DeliveryEnabled:    deliveryEnabled,
	}
}

// SetServiceCapability declares which ServiceType this driver offers and
// which TripType(s) — Ride and/or Delivery — they're currently enabled
// for. Separate from vehicle registration (NewDriverProfile/Update): a
// driver registers their physical vehicle once, then can declare or change
// their service tier/capability independently (e.g. upgrading from Bike to
// Bike Plus without re-entering vehicle details). Rejects a ServiceType
// whose RequiredVehicleType doesn't match this driver's registered
// VehicleType — e.g. ServiceType=car_xl on a motorcycle profile.
func (d *DriverProfile) SetServiceCapability(serviceType ServiceType, rideEnabled, deliveryEnabled bool, now time.Time) error {
	if err := validateServiceType(serviceType); err != nil {
		return err
	}
	if serviceType.RequiredVehicleType() != d.VehicleType {
		return errors.InvalidArgument(
			"service type " + string(serviceType) + " requires vehicle type " +
				string(serviceType.RequiredVehicleType()) + ", driver has " + string(d.VehicleType))
	}
	d.ServiceType = serviceType
	d.RideEnabled = rideEnabled
	d.DeliveryEnabled = deliveryEnabled
	d.UpdatedAt = now
	return nil
}

// Update replaces vehicle information and license number.
// Callers should note that changing these fields may warrant re-verification
// in a production workflow; that logic lives at the application layer.
func (d *DriverProfile) Update(
	licenseNumber string,
	vehicleType VehicleType,
	vehicleBrand, vehicleModel, vehicleColor, plateNumber string,
	now time.Time,
) error {
	if strings.TrimSpace(licenseNumber) == "" {
		return errors.InvalidArgument("license number must not be empty")
	}
	if err := validateVehicleType(vehicleType); err != nil {
		return err
	}
	if strings.TrimSpace(plateNumber) == "" {
		return errors.InvalidArgument("plate number must not be empty")
	}
	d.LicenseNumber = licenseNumber
	d.VehicleType = vehicleType
	d.VehicleBrand = vehicleBrand
	d.VehicleModel = vehicleModel
	d.VehicleColor = vehicleColor
	d.PlateNumber = plateNumber
	d.UpdatedAt = now
	return nil
}

// GoOnline transitions the driver to OnlineStatusOnline.
// The driver must be VerificationStatusVerified; any other status returns CodePreconditionFailed.
func (d *DriverProfile) GoOnline(now time.Time) error {
	if d.VerificationStatus != VerificationStatusVerified {
		return errors.PreconditionFailed("driver must be verified before going online")
	}
	if d.OnlineStatus == OnlineStatusOnline {
		return errors.PreconditionFailed("driver is already online")
	}
	d.OnlineStatus = OnlineStatusOnline
	d.UpdatedAt = now
	return nil
}

// GoOffline transitions the driver to OnlineStatusOffline.
// Returns CodePreconditionFailed if already offline.
func (d *DriverProfile) GoOffline(now time.Time) error {
	if d.OnlineStatus == OnlineStatusOffline {
		return errors.PreconditionFailed("driver is already offline")
	}
	d.OnlineStatus = OnlineStatusOffline
	d.UpdatedAt = now
	return nil
}

// Verify approves a pending driver.
// Returns CodePreconditionFailed if not in VerificationStatusPending.
func (d *DriverProfile) Verify(now time.Time) error {
	if d.VerificationStatus != VerificationStatusPending {
		return errors.PreconditionFailed("only pending drivers can be verified")
	}
	d.VerificationStatus = VerificationStatusVerified
	d.UpdatedAt = now
	return nil
}

// Reject moves a pending driver to VerificationStatusRejected.
// Returns CodePreconditionFailed if the driver is not pending.
func (d *DriverProfile) Reject(now time.Time) error {
	if d.VerificationStatus != VerificationStatusPending {
		return errors.PreconditionFailed("only pending drivers can be rejected")
	}
	d.VerificationStatus = VerificationStatusRejected
	d.UpdatedAt = now
	return nil
}

// Suspend moves a verified driver to VerificationStatusSuspended and forces offline.
// Returns CodePreconditionFailed if the driver is not verified.
func (d *DriverProfile) Suspend(now time.Time) error {
	if d.VerificationStatus != VerificationStatusVerified {
		return errors.PreconditionFailed("only verified drivers can be suspended")
	}
	d.VerificationStatus = VerificationStatusSuspended
	d.OnlineStatus = OnlineStatusOffline
	d.UpdatedAt = now
	return nil
}

// Reinstate moves a suspended driver back to VerificationStatusVerified.
// Returns CodePreconditionFailed if the driver is not suspended.
func (d *DriverProfile) Reinstate(now time.Time) error {
	if d.VerificationStatus != VerificationStatusSuspended {
		return errors.PreconditionFailed("only suspended drivers can be reinstated")
	}
	d.VerificationStatus = VerificationStatusVerified
	d.UpdatedAt = now
	return nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func validateVehicleType(vt VehicleType) error {
	switch vt {
	case VehicleTypeCar, VehicleTypeMotorcycle, VehicleTypeVan:
		return nil
	default:
		return errors.InvalidArgument("unknown vehicle type: " + string(vt))
	}
}

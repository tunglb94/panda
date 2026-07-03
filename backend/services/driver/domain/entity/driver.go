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
	}
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

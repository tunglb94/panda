package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

const (
	vehicleYearMin = 1900
	vehicleYearMax = 1 // added to current year; enforced in New/Update callers
)

// Vehicle is a vehicle registered by a driver.
// A driver may own multiple vehicles.
// VehicleType reuses the same enum defined in driver.go.
type Vehicle struct {
	VehicleID   string
	DriverID    string
	Type        VehicleType
	Brand       string
	Model       string
	Color       string
	PlateNumber string
	Year        int // 0 = not provided
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewVehicle creates a validated Vehicle.
// Required: vehicleID, driverID, vehicleType, plateNumber.
// Optional: brand, model, color, year (0 = not provided).
// now is used to bound the manufacturing year upper limit.
func NewVehicle(
	vehicleID, driverID string,
	vehicleType VehicleType,
	brand, model, color, plateNumber string,
	year int,
	now time.Time,
) (*Vehicle, error) {
	if vehicleID == "" {
		return nil, errors.InvalidArgument("vehicle id must not be empty")
	}
	if driverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	if err := validateVehicleType(vehicleType); err != nil {
		return nil, err
	}
	if strings.TrimSpace(plateNumber) == "" {
		return nil, errors.InvalidArgument("plate number must not be empty")
	}
	if err := validateYear(year, now); err != nil {
		return nil, err
	}
	return &Vehicle{
		VehicleID:   vehicleID,
		DriverID:    driverID,
		Type:        vehicleType,
		Brand:       brand,
		Model:       model,
		Color:       color,
		PlateNumber: plateNumber,
		Year:        year,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// ReconstituteVehicle rebuilds a Vehicle from a persistence record.
// No validation is applied — data is assumed already valid.
func ReconstituteVehicle(
	vehicleID, driverID string,
	vehicleType VehicleType,
	brand, model, color, plateNumber string,
	year int,
	createdAt, updatedAt time.Time,
) *Vehicle {
	return &Vehicle{
		VehicleID:   vehicleID,
		DriverID:    driverID,
		Type:        vehicleType,
		Brand:       brand,
		Model:       model,
		Color:       color,
		PlateNumber: plateNumber,
		Year:        year,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// Update replaces mutable vehicle fields.
// Required: vehicleType, plateNumber. Optional: brand, model, color, year.
// now is used to bound the manufacturing year upper limit.
func (v *Vehicle) Update(
	vehicleType VehicleType,
	brand, model, color, plateNumber string,
	year int,
	now time.Time,
) error {
	if err := validateVehicleType(vehicleType); err != nil {
		return err
	}
	if strings.TrimSpace(plateNumber) == "" {
		return errors.InvalidArgument("plate number must not be empty")
	}
	if err := validateYear(year, now); err != nil {
		return err
	}
	v.Type = vehicleType
	v.Brand = brand
	v.Model = model
	v.Color = color
	v.PlateNumber = plateNumber
	v.Year = year
	v.UpdatedAt = now
	return nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func validateYear(year int, now time.Time) error {
	if year == 0 {
		return nil // not provided
	}
	if year < vehicleYearMin {
		return errors.InvalidArgument("vehicle year must be 1900 or later")
	}
	maxYear := now.Year() + vehicleYearMax
	if year > maxYear {
		return errors.InvalidArgument("vehicle year must not be in the future")
	}
	return nil
}

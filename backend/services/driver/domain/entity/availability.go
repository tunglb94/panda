package entity

import "time"

// AvailabilityState is a value object representing a driver's real-time
// presence in the Redis availability store.
// IsOnline reflects whether the online key currently exists (has not expired).
// LastSeen is zero if the driver has never been seen.
type AvailabilityState struct {
	DriverID string
	IsOnline bool
	LastSeen time.Time // zero = never seen
}

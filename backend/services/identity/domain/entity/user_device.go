package entity

import (
	"time"

	"github.com/fairride/shared/errors"
)

// UserDevice is one (user, physical device) pairing — upserted on every
// successful login so the platform knows what's currently signed in
// (push targeting via FCMToken, support/fraud investigation via
// Platform/Model/AppVersion/LastSeen).
type UserDevice struct {
	UserID     string
	DeviceID   string
	Platform   string
	Model      string
	AppVersion string
	FCMToken   string
	LastSeen   time.Time
	CreatedAt  time.Time
}

// NewUserDevice creates a device row as of now (used for both first-seen
// inserts and upsert-refreshes — the repository's Upsert only actually
// writes CreatedAt on first insert, see infrastructure/postgres).
// userID and deviceID are required; the rest are best-effort and may be
// empty (a platform that can't determine its own model, no FCM token yet).
func NewUserDevice(userID, deviceID, platform, model, appVersion, fcmToken string, now time.Time) (*UserDevice, error) {
	if userID == "" {
		return nil, errors.InvalidArgument("user_id must not be empty")
	}
	if deviceID == "" {
		return nil, errors.InvalidArgument("device_id must not be empty")
	}
	return &UserDevice{
		UserID:     userID,
		DeviceID:   deviceID,
		Platform:   platform,
		Model:      model,
		AppVersion: appVersion,
		FCMToken:   fcmToken,
		LastSeen:   now,
		CreatedAt:  now,
	}, nil
}

// ReconstituteUserDevice rebuilds a UserDevice from a persistence record. No validation.
func ReconstituteUserDevice(userID, deviceID, platform, model, appVersion, fcmToken string, lastSeen, createdAt time.Time) *UserDevice {
	return &UserDevice{
		UserID:     userID,
		DeviceID:   deviceID,
		Platform:   platform,
		Model:      model,
		AppVersion: appVersion,
		FCMToken:   fcmToken,
		LastSeen:   lastSeen,
		CreatedAt:  createdAt,
	}
}

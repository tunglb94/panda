package entity

import (
	"time"

	"github.com/fairride/shared/errors"
)

// LoginMethod records how a login attempt was made.
type LoginMethod string

const (
	LoginMethodOTP    LoginMethod = "otp"
	LoginMethodGoogle LoginMethod = "google"
)

// LoginRecord is one append-only login-history row. UserID may be empty —
// a wrong OTP code or an unrecognized Google account can fail before any
// account is resolved, and the attempt is still worth recording.
type LoginRecord struct {
	ID          string
	UserID      string
	LoginTime   time.Time
	IP          string
	DeviceID    string
	Platform    string
	LoginMethod LoginMethod
	Success     bool
}

// NewLoginRecord creates a LoginRecord. id is required (caller-generated);
// every other field is best-effort — a login attempt with no resolvable
// user, IP, or device is still recorded with those fields empty.
func NewLoginRecord(id, userID string, loginTime time.Time, ip, deviceID, platform string, method LoginMethod, success bool) (*LoginRecord, error) {
	if id == "" {
		return nil, errors.InvalidArgument("id must not be empty")
	}
	return &LoginRecord{
		ID:          id,
		UserID:      userID,
		LoginTime:   loginTime,
		IP:          ip,
		DeviceID:    deviceID,
		Platform:    platform,
		LoginMethod: method,
		Success:     success,
	}, nil
}

// ReconstituteLoginRecord rebuilds a LoginRecord from a persistence record. No validation.
func ReconstituteLoginRecord(id, userID string, loginTime time.Time, ip, deviceID, platform string, method LoginMethod, success bool) *LoginRecord {
	return &LoginRecord{
		ID:          id,
		UserID:      userID,
		LoginTime:   loginTime,
		IP:          ip,
		DeviceID:    deviceID,
		Platform:    platform,
		LoginMethod: method,
		Success:     success,
	}
}

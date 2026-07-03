package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// UserType distinguishes the functional role of a User in the FAIRRIDE platform.
// Values map directly to the user segments defined in DOC-0002 §6.12.
type UserType string

const (
	TypeRider         UserType = "rider"
	TypeDriver        UserType = "driver"
	TypeFleetOperator UserType = "fleet_operator"
	TypeAdmin         UserType = "admin"
)

// UserStatus tracks the lifecycle state of a User account.
// Valid transitions:
//
//	PendingVerification → Active   (phone OTP verified)
//	Active              → Suspended
//	Suspended           → Active
//	Active              → Deactivated (terminal)
//	Suspended           → Deactivated (terminal)
type UserStatus string

const (
	StatusPendingVerification UserStatus = "pending_verification"
	StatusActive              UserStatus = "active"
	StatusSuspended           UserStatus = "suspended"
	StatusDeactivated         UserStatus = "deactivated"
)

// User is the platform account entity. One User exists per unique phone number.
// Email is optional; an empty string means not provided.
// RoleID links to an identity.Role that governs permissions.
type User struct {
	ID          string
	PhoneNumber string
	Name        string
	Email       string
	Type        UserType
	Status      UserStatus
	RoleID      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewUser creates a new User account in StatusPendingVerification.
// id, phoneNumber, name, and roleID must be non-empty.
// email is optional — pass an empty string to omit it; if non-empty it must be valid.
// userType must be one of the defined UserType constants.
func NewUser(
	id, phoneNumber, name, email string,
	userType UserType,
	roleID string,
	now time.Time,
) (*User, error) {
	if id == "" {
		return nil, errors.InvalidArgument("user id must not be empty")
	}
	if strings.TrimSpace(phoneNumber) == "" {
		return nil, errors.InvalidArgument("user phone number must not be empty")
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.InvalidArgument("user name must not be empty")
	}
	if err := validateUserType(userType); err != nil {
		return nil, err
	}
	if email != "" {
		if err := validateEmail(email); err != nil {
			return nil, err
		}
	}
	if roleID == "" {
		return nil, errors.InvalidArgument("user role id must not be empty")
	}
	return &User{
		ID:          id,
		PhoneNumber: phoneNumber,
		Name:        name,
		Email:       email,
		Type:        userType,
		Status:      StatusPendingVerification,
		RoleID:      roleID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// ReconstituteUser rebuilds a User from a persistence record.
// No validation is applied — data is assumed already valid.
func ReconstituteUser(
	id, phoneNumber, name, email string,
	userType UserType,
	status UserStatus,
	roleID string,
	createdAt, updatedAt time.Time,
) *User {
	return &User{
		ID:          id,
		PhoneNumber: phoneNumber,
		Name:        name,
		Email:       email,
		Type:        userType,
		Status:      status,
		RoleID:      roleID,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// Activate transitions the user to StatusActive.
// Allowed from StatusPendingVerification (initial phone verification) or StatusSuspended (reinstatement).
// Returns CodePreconditionFailed for any other starting status.
func (u *User) Activate(now time.Time) error {
	if u.Status != StatusPendingVerification && u.Status != StatusSuspended {
		return errors.PreconditionFailed("cannot activate user with status: " + string(u.Status))
	}
	u.Status = StatusActive
	u.UpdatedAt = now
	return nil
}

// Suspend transitions the user from StatusActive to StatusSuspended.
// Returns CodePreconditionFailed if the current status is not Active.
func (u *User) Suspend(now time.Time) error {
	if u.Status != StatusActive {
		return errors.PreconditionFailed("cannot suspend user with status: " + string(u.Status))
	}
	u.Status = StatusSuspended
	u.UpdatedAt = now
	return nil
}

// Deactivate permanently deactivates the user. This is a terminal transition.
// Allowed from StatusActive or StatusSuspended.
// Returns CodePreconditionFailed if the user is already Deactivated or still PendingVerification.
func (u *User) Deactivate(now time.Time) error {
	switch u.Status {
	case StatusDeactivated:
		return errors.PreconditionFailed("user is already deactivated")
	case StatusPendingVerification:
		return errors.PreconditionFailed("cannot deactivate user with status: " + string(u.Status))
	}
	u.Status = StatusDeactivated
	u.UpdatedAt = now
	return nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func validateUserType(t UserType) error {
	switch t {
	case TypeRider, TypeDriver, TypeFleetOperator, TypeAdmin:
		return nil
	default:
		return errors.InvalidArgument("unknown user type: " + string(t))
	}
}

// validateEmail performs a minimal structural check:
// non-empty local part, "@", non-empty domain containing at least one interior dot.
func validateEmail(email string) error {
	atIdx := strings.Index(email, "@")
	if atIdx < 1 {
		return errors.InvalidArgument("invalid email address: " + email)
	}
	domain := email[atIdx+1:]
	dotIdx := strings.Index(domain, ".")
	if dotIdx < 1 || dotIdx == len(domain)-1 {
		return errors.InvalidArgument("invalid email address: " + email)
	}
	return nil
}

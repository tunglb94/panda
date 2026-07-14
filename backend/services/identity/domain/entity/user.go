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

// User is the platform account entity. PhoneNumber, Email, and GoogleSub are
// each optional — empty string means "not provided" — but whichever ones
// are set must be unique platform-wide (enforced by the persistence layer,
// see migration 020). A phone-OTP signup sets PhoneNumber only; a Google
// signup sets Email + GoogleSub only; either can later gain the other via
// FindOrCreateUserUseCase's auto-link (matching by email) or a future
// explicit account-merge flow (not implemented yet — see plan's Known Gaps).
//
// Every account defaults to TypeRider/RoleRider regardless of which app
// first created it (Driver app included) — DriverEnabled is the capability
// flag that actually gates the Driver app, independent of Type/RoleID. One
// account can be Rider and Driver-enabled at the same time.
// RoleID links to an identity.Role that governs permissions.
type User struct {
	ID            string
	PhoneNumber   string
	Name          string
	Email         string
	GoogleSub     string
	Type          UserType
	Status        UserStatus
	RoleID        string
	DriverEnabled bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
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

// NewGoogleUser creates a new User account from a verified Google Sign-In —
// no phone number, since Google never provides one. email and googleSub
// must both be non-empty (GoogleLoginUseCase already requires a verified
// email before calling this). Always TypeRider — see the User doc comment
// on why Type/RoleID no longer vary by which app first created the account.
func NewGoogleUser(id, email, googleSub, name, roleID string, now time.Time) (*User, error) {
	if id == "" {
		return nil, errors.InvalidArgument("user id must not be empty")
	}
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if strings.TrimSpace(googleSub) == "" {
		return nil, errors.InvalidArgument("google sub must not be empty")
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.InvalidArgument("user name must not be empty")
	}
	if roleID == "" {
		return nil, errors.InvalidArgument("user role id must not be empty")
	}
	return &User{
		ID:        id,
		Name:      name,
		Email:     email,
		GoogleSub: googleSub,
		Type:      TypeRider,
		Status:    StatusPendingVerification,
		RoleID:    roleID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ReconstituteUser rebuilds a User from a persistence record.
// No validation is applied — data is assumed already valid.
func ReconstituteUser(
	id, phoneNumber, name, email, googleSub string,
	userType UserType,
	status UserStatus,
	roleID string,
	driverEnabled bool,
	createdAt, updatedAt time.Time,
) *User {
	return &User{
		ID:            id,
		PhoneNumber:   phoneNumber,
		Name:          name,
		Email:         email,
		GoogleSub:     googleSub,
		Type:          userType,
		Status:        status,
		RoleID:        roleID,
		DriverEnabled: driverEnabled,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
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

// EnableDriverCapability idempotently flips DriverEnabled on — called the
// first time an account logs into the Driver app. Does not touch Type or
// RoleID (a Driver-enabled account is still, and remains, a Rider account —
// see the User doc comment).
func (u *User) EnableDriverCapability(now time.Time) {
	if u.DriverEnabled {
		return
	}
	u.DriverEnabled = true
	u.UpdatedAt = now
}

// LinkGoogleSub attaches a verified Google subject ID to an account that
// doesn't have one yet — used when GoogleLoginUseCase resolves an existing
// phone/email account by email match rather than by GoogleSub (auto-link,
// not a full account merge; see the User doc comment).
func (u *User) LinkGoogleSub(sub string, now time.Time) error {
	if strings.TrimSpace(sub) == "" {
		return errors.InvalidArgument("google sub must not be empty")
	}
	if u.GoogleSub != "" && u.GoogleSub != sub {
		return errors.PreconditionFailed("account is already linked to a different Google account")
	}
	u.GoogleSub = sub
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

// Package entity defines the User Profile domain model for the FAIRRIDE user service.
package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// Gender represents the user's self-reported gender.
type Gender string

const (
	GenderMale        Gender = "male"
	GenderFemale      Gender = "female"
	GenderOther       Gender = "other"
	GenderUnspecified Gender = "unspecified"
)

// ProfileStatus represents the visibility and usability state of a user profile.
type ProfileStatus string

const (
	ProfileStatusActive    ProfileStatus = "active"
	ProfileStatusSuspended ProfileStatus = "suspended"
	ProfileStatusDeleted   ProfileStatus = "deleted"
)

// UserProfile holds the public-facing data for a FAIRRIDE user.
// Phone is authoritative from the Identity service and is not updatable here.
// DateOfBirth zero value means not provided.
type UserProfile struct {
	ID          string
	FullName    string
	Phone       string
	Email       string
	Avatar      string
	DateOfBirth time.Time     // zero = not provided
	Gender      Gender
	Status      ProfileStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewUserProfile creates a new UserProfile in ProfileStatusActive.
// id, fullName, and phone must be non-empty.
// email is optional — empty string is accepted; if non-empty it must be valid.
// avatar is optional — any non-empty string is accepted (URL validated by caller).
// dateOfBirth zero means not provided; if non-zero it must be in the past.
// gender must be one of the defined constants.
func NewUserProfile(
	id, fullName, phone, email, avatar string,
	dateOfBirth time.Time,
	gender Gender,
	now time.Time,
) (*UserProfile, error) {
	if id == "" {
		return nil, errors.InvalidArgument("profile id must not be empty")
	}
	if strings.TrimSpace(fullName) == "" {
		return nil, errors.InvalidArgument("full name must not be empty")
	}
	if strings.TrimSpace(phone) == "" {
		return nil, errors.InvalidArgument("phone must not be empty")
	}
	if email != "" {
		if err := validateEmail(email); err != nil {
			return nil, err
		}
	}
	if err := validateGender(gender); err != nil {
		return nil, err
	}
	if !dateOfBirth.IsZero() {
		if err := validateDateOfBirth(dateOfBirth, now); err != nil {
			return nil, err
		}
	}
	return &UserProfile{
		ID:          id,
		FullName:    fullName,
		Phone:       phone,
		Email:       email,
		Avatar:      avatar,
		DateOfBirth: dateOfBirth,
		Gender:      gender,
		Status:      ProfileStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// ReconstituteUserProfile rebuilds a UserProfile from a persistence record.
// No validation is applied — data is assumed already valid.
func ReconstituteUserProfile(
	id, fullName, phone, email, avatar string,
	dateOfBirth time.Time,
	gender Gender,
	status ProfileStatus,
	createdAt, updatedAt time.Time,
) *UserProfile {
	return &UserProfile{
		ID:          id,
		FullName:    fullName,
		Phone:       phone,
		Email:       email,
		Avatar:      avatar,
		DateOfBirth: dateOfBirth,
		Gender:      gender,
		Status:      status,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// Update replaces the mutable fields of a UserProfile.
// fullName is required; email and avatar are optional (empty string clears them).
// dateOfBirth zero clears the field; non-zero must be in the past.
// gender must be a valid constant.
func (p *UserProfile) Update(
	fullName, email, avatar string,
	dateOfBirth time.Time,
	gender Gender,
	now time.Time,
) error {
	if strings.TrimSpace(fullName) == "" {
		return errors.InvalidArgument("full name must not be empty")
	}
	if email != "" {
		if err := validateEmail(email); err != nil {
			return err
		}
	}
	if err := validateGender(gender); err != nil {
		return err
	}
	if !dateOfBirth.IsZero() {
		if err := validateDateOfBirth(dateOfBirth, now); err != nil {
			return err
		}
	}
	p.FullName = fullName
	p.Email = email
	p.Avatar = avatar
	p.DateOfBirth = dateOfBirth
	p.Gender = gender
	p.UpdatedAt = now
	return nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func validateGender(g Gender) error {
	switch g {
	case GenderMale, GenderFemale, GenderOther, GenderUnspecified:
		return nil
	default:
		return errors.InvalidArgument("unknown gender value: " + string(g))
	}
}

func validateDateOfBirth(dob, now time.Time) error {
	if !dob.Before(now) {
		return errors.InvalidArgument("date of birth must be in the past")
	}
	if now.Year()-dob.Year() > 150 {
		return errors.InvalidArgument("date of birth is implausibly far in the past")
	}
	return nil
}

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

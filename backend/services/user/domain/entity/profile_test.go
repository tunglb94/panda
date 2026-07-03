package entity

import (
	"testing"
	"time"

	domainerrors "github.com/fairride/shared/errors"
)

var testNow = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

func validProfile(t *testing.T) *UserProfile {
	t.Helper()
	p, err := NewUserProfile("u-1", "Nguyen Van A", "+84901234567", "", "", time.Time{}, GenderUnspecified, testNow)
	if err != nil {
		t.Fatalf("validProfile: %v", err)
	}
	return p
}

// ─── NewUserProfile ───────────────────────────────────────────────────────────

func TestNewUserProfile_Valid_Minimal(t *testing.T) {
	p, err := NewUserProfile("u-1", "Nguyen Van A", "+84901234567", "", "", time.Time{}, GenderUnspecified, testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != "u-1" {
		t.Errorf("ID: got %q", p.ID)
	}
	if p.FullName != "Nguyen Van A" {
		t.Errorf("FullName: got %q", p.FullName)
	}
	if p.Phone != "+84901234567" {
		t.Errorf("Phone: got %q", p.Phone)
	}
	if p.Status != ProfileStatusActive {
		t.Errorf("Status: got %q, want %q", p.Status, ProfileStatusActive)
	}
	if !p.CreatedAt.Equal(testNow) {
		t.Errorf("CreatedAt: got %v", p.CreatedAt)
	}
	if !p.UpdatedAt.Equal(testNow) {
		t.Errorf("UpdatedAt: got %v", p.UpdatedAt)
	}
}

func TestNewUserProfile_Valid_AllFields(t *testing.T) {
	dob := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	p, err := NewUserProfile("u-1", "Ahmad", "+601112345678", "ahmad@example.com", "https://cdn.example.com/avatar.jpg", dob, GenderMale, testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Email != "ahmad@example.com" {
		t.Errorf("Email: got %q", p.Email)
	}
	if p.Avatar != "https://cdn.example.com/avatar.jpg" {
		t.Errorf("Avatar: got %q", p.Avatar)
	}
	if !p.DateOfBirth.Equal(dob) {
		t.Errorf("DateOfBirth: got %v", p.DateOfBirth)
	}
	if p.Gender != GenderMale {
		t.Errorf("Gender: got %q", p.Gender)
	}
}

func TestNewUserProfile_EmptyID(t *testing.T) {
	_, err := NewUserProfile("", "Name", "+1", "", "", time.Time{}, GenderUnspecified, testNow)
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestNewUserProfile_EmptyFullName(t *testing.T) {
	_, err := NewUserProfile("u-1", "", "+1", "", "", time.Time{}, GenderUnspecified, testNow)
	if err == nil {
		t.Fatal("expected error for empty full name")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestNewUserProfile_WhitespaceFullName(t *testing.T) {
	_, err := NewUserProfile("u-1", "  ", "+1", "", "", time.Time{}, GenderUnspecified, testNow)
	if err == nil {
		t.Fatal("expected error for whitespace-only full name")
	}
}

func TestNewUserProfile_EmptyPhone(t *testing.T) {
	_, err := NewUserProfile("u-1", "Name", "", "", "", time.Time{}, GenderUnspecified, testNow)
	if err == nil {
		t.Fatal("expected error for empty phone")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestNewUserProfile_InvalidGender(t *testing.T) {
	_, err := NewUserProfile("u-1", "Name", "+1", "", "", time.Time{}, "unknown", testNow)
	if err == nil {
		t.Fatal("expected error for invalid gender")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestNewUserProfile_AllGenders(t *testing.T) {
	genders := []Gender{GenderMale, GenderFemale, GenderOther, GenderUnspecified}
	for _, g := range genders {
		_, err := NewUserProfile("u-1", "Name", "+1", "", "", time.Time{}, g, testNow)
		if err != nil {
			t.Errorf("gender %q: unexpected error: %v", g, err)
		}
	}
}

func TestNewUserProfile_EmailValidation(t *testing.T) {
	cases := []struct {
		email   string
		wantErr bool
	}{
		{"user@example.com", false},
		{"", false},
		{"notanemail", true},
		{"@domain.com", true},
		{"user@nodot", true},
	}
	for _, tc := range cases {
		_, err := NewUserProfile("u-1", "Name", "+1", tc.email, "", time.Time{}, GenderUnspecified, testNow)
		if tc.wantErr && err == nil {
			t.Errorf("email %q: expected error, got nil", tc.email)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("email %q: unexpected error: %v", tc.email, err)
		}
	}
}

func TestNewUserProfile_DateOfBirth_InFuture(t *testing.T) {
	future := testNow.Add(24 * time.Hour)
	_, err := NewUserProfile("u-1", "Name", "+1", "", "", future, GenderUnspecified, testNow)
	if err == nil {
		t.Fatal("expected error for future date of birth")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestNewUserProfile_DateOfBirth_TooOld(t *testing.T) {
	tooOld := testNow.AddDate(-151, 0, 0)
	_, err := NewUserProfile("u-1", "Name", "+1", "", "", tooOld, GenderUnspecified, testNow)
	if err == nil {
		t.Fatal("expected error for implausibly old date of birth")
	}
}

func TestNewUserProfile_DateOfBirth_Valid(t *testing.T) {
	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := NewUserProfile("u-1", "Name", "+1", "", "", dob, GenderUnspecified, testNow)
	if err != nil {
		t.Errorf("valid dob: unexpected error: %v", err)
	}
}

func TestNewUserProfile_DateOfBirth_Zero_Allowed(t *testing.T) {
	_, err := NewUserProfile("u-1", "Name", "+1", "", "", time.Time{}, GenderUnspecified, testNow)
	if err != nil {
		t.Errorf("zero dob should be allowed: %v", err)
	}
}

// ─── ReconstituteUserProfile ─────────────────────────────────────────────────

func TestReconstituteUserProfile_NoValidation(t *testing.T) {
	// Reconstitute must accept any values, even ones that would fail NewUserProfile.
	p := ReconstituteUserProfile(
		"u-42", "Ahmad", "+601112345678", "", "", time.Time{},
		GenderOther, ProfileStatusSuspended,
		testNow, testNow.Add(time.Hour),
	)
	if p.Status != ProfileStatusSuspended {
		t.Errorf("Status: got %q, want Suspended", p.Status)
	}
	if p.ID != "u-42" {
		t.Errorf("ID: got %q", p.ID)
	}
}

// ─── Update ──────────────────────────────────────────────────────────────────

func TestUpdate_ValidFields(t *testing.T) {
	p := validProfile(t)
	later := testNow.Add(time.Hour)
	dob := time.Date(1995, 3, 20, 0, 0, 0, 0, time.UTC)

	err := p.Update("New Name", "new@example.com", "https://cdn.example.com/a.jpg", dob, GenderFemale, later)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if p.FullName != "New Name" {
		t.Errorf("FullName: got %q", p.FullName)
	}
	if p.Email != "new@example.com" {
		t.Errorf("Email: got %q", p.Email)
	}
	if p.Avatar != "https://cdn.example.com/a.jpg" {
		t.Errorf("Avatar: got %q", p.Avatar)
	}
	if !p.DateOfBirth.Equal(dob) {
		t.Errorf("DateOfBirth: got %v", p.DateOfBirth)
	}
	if p.Gender != GenderFemale {
		t.Errorf("Gender: got %q", p.Gender)
	}
	if !p.UpdatedAt.Equal(later) {
		t.Errorf("UpdatedAt: got %v, want %v", p.UpdatedAt, later)
	}
}

func TestUpdate_ClearsEmail(t *testing.T) {
	p := validProfile(t)
	_ = p.Update("Name", "initial@example.com", "", time.Time{}, GenderUnspecified, testNow)
	_ = p.Update("Name", "", "", time.Time{}, GenderUnspecified, testNow.Add(time.Hour))
	if p.Email != "" {
		t.Errorf("Email should be cleared, got %q", p.Email)
	}
}

func TestUpdate_ClearsDateOfBirth(t *testing.T) {
	p := validProfile(t)
	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	_ = p.Update("Name", "", "", dob, GenderUnspecified, testNow)
	_ = p.Update("Name", "", "", time.Time{}, GenderUnspecified, testNow.Add(time.Hour))
	if !p.DateOfBirth.IsZero() {
		t.Errorf("DateOfBirth should be cleared, got %v", p.DateOfBirth)
	}
}

func TestUpdate_EmptyFullName_Error(t *testing.T) {
	p := validProfile(t)
	err := p.Update("", "", "", time.Time{}, GenderUnspecified, testNow)
	if err == nil {
		t.Fatal("expected error for empty full name in update")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestUpdate_InvalidEmail_Error(t *testing.T) {
	p := validProfile(t)
	err := p.Update("Name", "bademail", "", time.Time{}, GenderUnspecified, testNow)
	if err == nil {
		t.Fatal("expected error for invalid email in update")
	}
}

func TestUpdate_FutureDateOfBirth_Error(t *testing.T) {
	p := validProfile(t)
	future := testNow.Add(24 * time.Hour)
	err := p.Update("Name", "", "", future, GenderUnspecified, testNow)
	if err == nil {
		t.Fatal("expected error for future date of birth in update")
	}
}

func TestUpdate_PhoneNotChanged(t *testing.T) {
	p := validProfile(t)
	originalPhone := p.Phone
	_ = p.Update("New Name", "", "", time.Time{}, GenderUnspecified, testNow.Add(time.Hour))
	if p.Phone != originalPhone {
		t.Errorf("Phone should not change during update: got %q, want %q", p.Phone, originalPhone)
	}
}

func TestUpdate_StatusNotChanged(t *testing.T) {
	p := validProfile(t)
	originalStatus := p.Status
	_ = p.Update("New Name", "", "", time.Time{}, GenderUnspecified, testNow.Add(time.Hour))
	if p.Status != originalStatus {
		t.Errorf("Status should not change during update: got %q, want %q", p.Status, originalStatus)
	}
}

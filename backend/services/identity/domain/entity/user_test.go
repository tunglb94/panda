package entity

import (
	"testing"
	"time"

	domainerrors "github.com/fairride/shared/errors"
)

var testUserNow = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

func newValidUser(t *testing.T) *User {
	t.Helper()
	u, err := NewUser("u-1", "+84901234567", "Nguyen Van A", "", TypeRider, "role-1", testUserNow)
	if err != nil {
		t.Fatalf("newValidUser: %v", err)
	}
	return u
}

// ─── NewUser construction ─────────────────────────────────────────────────────

func TestNewUser_Valid(t *testing.T) {
	u, err := NewUser("u-1", "+84901234567", "Nguyen Van A", "", TypeRider, "role-1", testUserNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != "u-1" {
		t.Errorf("ID: got %q, want %q", u.ID, "u-1")
	}
	if u.PhoneNumber != "+84901234567" {
		t.Errorf("PhoneNumber: got %q", u.PhoneNumber)
	}
	if u.Name != "Nguyen Van A" {
		t.Errorf("Name: got %q", u.Name)
	}
	if u.Type != TypeRider {
		t.Errorf("Type: got %q, want %q", u.Type, TypeRider)
	}
	if u.Status != StatusPendingVerification {
		t.Errorf("Status: got %q, want %q", u.Status, StatusPendingVerification)
	}
	if u.RoleID != "role-1" {
		t.Errorf("RoleID: got %q, want %q", u.RoleID, "role-1")
	}
	if !u.CreatedAt.Equal(testUserNow) {
		t.Errorf("CreatedAt: got %v, want %v", u.CreatedAt, testUserNow)
	}
	if !u.UpdatedAt.Equal(testUserNow) {
		t.Errorf("UpdatedAt: got %v, want %v", u.UpdatedAt, testUserNow)
	}
}

func TestNewUser_WithEmail(t *testing.T) {
	u, err := NewUser("u-1", "+84901234567", "Alex", "alex@example.com", TypeRider, "role-1", testUserNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Email != "alex@example.com" {
		t.Errorf("Email: got %q", u.Email)
	}
}

func TestNewUser_EmptyEmail_Allowed(t *testing.T) {
	_, err := NewUser("u-1", "+84901234567", "Alex", "", TypeRider, "role-1", testUserNow)
	if err != nil {
		t.Errorf("empty email should be allowed, got error: %v", err)
	}
}

func TestNewUser_EmptyID(t *testing.T) {
	_, err := NewUser("", "+84901234567", "Alex", "", TypeRider, "role-1", testUserNow)
	if err == nil {
		t.Fatal("expected error for empty id, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestNewUser_EmptyPhone(t *testing.T) {
	_, err := NewUser("u-1", "", "Alex", "", TypeRider, "role-1", testUserNow)
	if err == nil {
		t.Fatal("expected error for empty phone, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestNewUser_WhitespacePhone(t *testing.T) {
	_, err := NewUser("u-1", "   ", "Alex", "", TypeRider, "role-1", testUserNow)
	if err == nil {
		t.Fatal("expected error for whitespace-only phone, got nil")
	}
}

func TestNewUser_EmptyName(t *testing.T) {
	_, err := NewUser("u-1", "+84901234567", "", "", TypeRider, "role-1", testUserNow)
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestNewUser_WhitespaceName(t *testing.T) {
	_, err := NewUser("u-1", "+84901234567", "  ", "", TypeRider, "role-1", testUserNow)
	if err == nil {
		t.Fatal("expected error for whitespace-only name, got nil")
	}
}

func TestNewUser_InvalidType(t *testing.T) {
	_, err := NewUser("u-1", "+84901234567", "Alex", "", "unknown_type", "role-1", testUserNow)
	if err == nil {
		t.Fatal("expected error for invalid type, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestNewUser_EmptyRoleID(t *testing.T) {
	_, err := NewUser("u-1", "+84901234567", "Alex", "", TypeRider, "", testUserNow)
	if err == nil {
		t.Fatal("expected error for empty roleID, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestNewUser_AllTypes(t *testing.T) {
	types := []UserType{TypeRider, TypeDriver, TypeFleetOperator, TypeAdmin}
	for _, ut := range types {
		_, err := NewUser("u-1", "+1", "Name", "", ut, "role-1", testUserNow)
		if err != nil {
			t.Errorf("type %q: unexpected error: %v", ut, err)
		}
	}
}

// ─── Email validation ─────────────────────────────────────────────────────────

func TestNewUser_EmailValidation(t *testing.T) {
	cases := []struct {
		email   string
		wantErr bool
	}{
		{"user@example.com", false},
		{"u@a.io", false},
		{"user.name+tag@sub.domain.org", false},
		{"", false},           // empty is allowed
		{"notanemail", true},  // no @
		{"@domain.com", true}, // empty local part
		{"user@", true},       // empty domain
		{"user@nodot", true},  // domain without dot
		{"user@.start", true}, // domain starts with dot
		{"user@end.", true},   // domain ends with dot
	}
	for _, tc := range cases {
		_, err := NewUser("u-1", "+1", "Name", tc.email, TypeRider, "role-1", testUserNow)
		if tc.wantErr && err == nil {
			t.Errorf("email %q: expected error, got nil", tc.email)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("email %q: unexpected error: %v", tc.email, err)
		}
	}
}

// ─── ReconstituteUser ────────────────────────────────────────────────────────

func TestReconstituteUser_ArbitraryState(t *testing.T) {
	// Reconstitute must accept any status without validation.
	u := ReconstituteUser(
		"u-42", "+84999999999", "Ahmad", "ahmad@example.com", "",
		TypeDriver, StatusSuspended, "role-driver", false,
		testUserNow, testUserNow.Add(24*time.Hour),
	)
	if u.ID != "u-42" {
		t.Errorf("ID: got %q", u.ID)
	}
	if u.Status != StatusSuspended {
		t.Errorf("Status: got %q, want %q", u.Status, StatusSuspended)
	}
	if u.Type != TypeDriver {
		t.Errorf("Type: got %q, want %q", u.Type, TypeDriver)
	}
}

// ─── Activate ────────────────────────────────────────────────────────────────

func TestUser_Activate_FromPendingVerification(t *testing.T) {
	u := newValidUser(t) // starts as PendingVerification
	later := testUserNow.Add(time.Minute)

	if err := u.Activate(later); err != nil {
		t.Fatalf("Activate from PendingVerification: %v", err)
	}
	if u.Status != StatusActive {
		t.Errorf("Status: got %q, want %q", u.Status, StatusActive)
	}
	if !u.UpdatedAt.Equal(later) {
		t.Errorf("UpdatedAt not updated: got %v, want %v", u.UpdatedAt, later)
	}
}

func TestUser_Activate_FromSuspended(t *testing.T) {
	u := newValidUser(t)
	later := testUserNow.Add(time.Minute)
	_ = u.Activate(later)       // PendingVerification → Active
	_ = u.Suspend(later)        // Active → Suspended
	even := later.Add(time.Minute)

	if err := u.Activate(even); err != nil {
		t.Fatalf("Activate from Suspended: %v", err)
	}
	if u.Status != StatusActive {
		t.Errorf("Status: got %q, want %q", u.Status, StatusActive)
	}
}

func TestUser_Activate_FromActive_Error(t *testing.T) {
	u := newValidUser(t)
	later := testUserNow.Add(time.Minute)
	_ = u.Activate(later)

	err := u.Activate(later)
	if err == nil {
		t.Fatal("expected PreconditionFailed, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected CodePreconditionFailed, got %v", domainerrors.GetCode(err))
	}
}

func TestUser_Activate_FromDeactivated_Error(t *testing.T) {
	u := newValidUser(t)
	later := testUserNow.Add(time.Minute)
	_ = u.Activate(later)
	_ = u.Deactivate(later)

	err := u.Activate(later)
	if err == nil {
		t.Fatal("expected PreconditionFailed, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected CodePreconditionFailed, got %v", domainerrors.GetCode(err))
	}
}

// ─── Suspend ─────────────────────────────────────────────────────────────────

func TestUser_Suspend_FromActive(t *testing.T) {
	u := newValidUser(t)
	later := testUserNow.Add(time.Minute)
	_ = u.Activate(later)
	even := later.Add(time.Minute)

	if err := u.Suspend(even); err != nil {
		t.Fatalf("Suspend from Active: %v", err)
	}
	if u.Status != StatusSuspended {
		t.Errorf("Status: got %q, want %q", u.Status, StatusSuspended)
	}
	if !u.UpdatedAt.Equal(even) {
		t.Errorf("UpdatedAt: got %v, want %v", u.UpdatedAt, even)
	}
}

func TestUser_Suspend_FromPendingVerification_Error(t *testing.T) {
	u := newValidUser(t)
	err := u.Suspend(testUserNow)
	if err == nil {
		t.Fatal("expected PreconditionFailed, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected CodePreconditionFailed, got %v", domainerrors.GetCode(err))
	}
}

func TestUser_Suspend_FromSuspended_Error(t *testing.T) {
	u := newValidUser(t)
	later := testUserNow.Add(time.Minute)
	_ = u.Activate(later)
	_ = u.Suspend(later)

	err := u.Suspend(later)
	if err == nil {
		t.Fatal("expected PreconditionFailed, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected CodePreconditionFailed, got %v", domainerrors.GetCode(err))
	}
}

func TestUser_Suspend_FromDeactivated_Error(t *testing.T) {
	u := newValidUser(t)
	later := testUserNow.Add(time.Minute)
	_ = u.Activate(later)
	_ = u.Deactivate(later)

	err := u.Suspend(later)
	if err == nil {
		t.Fatal("expected PreconditionFailed, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected CodePreconditionFailed, got %v", domainerrors.GetCode(err))
	}
}

// ─── Deactivate ──────────────────────────────────────────────────────────────

func TestUser_Deactivate_FromActive(t *testing.T) {
	u := newValidUser(t)
	later := testUserNow.Add(time.Minute)
	_ = u.Activate(later)
	even := later.Add(time.Minute)

	if err := u.Deactivate(even); err != nil {
		t.Fatalf("Deactivate from Active: %v", err)
	}
	if u.Status != StatusDeactivated {
		t.Errorf("Status: got %q, want %q", u.Status, StatusDeactivated)
	}
	if !u.UpdatedAt.Equal(even) {
		t.Errorf("UpdatedAt: got %v, want %v", u.UpdatedAt, even)
	}
}

func TestUser_Deactivate_FromSuspended(t *testing.T) {
	u := newValidUser(t)
	later := testUserNow.Add(time.Minute)
	_ = u.Activate(later)
	_ = u.Suspend(later)

	if err := u.Deactivate(later.Add(time.Minute)); err != nil {
		t.Fatalf("Deactivate from Suspended: %v", err)
	}
	if u.Status != StatusDeactivated {
		t.Errorf("Status: got %q, want %q", u.Status, StatusDeactivated)
	}
}

func TestUser_Deactivate_FromPendingVerification_Error(t *testing.T) {
	u := newValidUser(t)
	err := u.Deactivate(testUserNow)
	if err == nil {
		t.Fatal("expected PreconditionFailed, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected CodePreconditionFailed, got %v", domainerrors.GetCode(err))
	}
}

func TestUser_Deactivate_AlreadyDeactivated_Error(t *testing.T) {
	u := newValidUser(t)
	later := testUserNow.Add(time.Minute)
	_ = u.Activate(later)
	_ = u.Deactivate(later)

	err := u.Deactivate(later)
	if err == nil {
		t.Fatal("expected PreconditionFailed for double-deactivate, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected CodePreconditionFailed, got %v", domainerrors.GetCode(err))
	}
}

// ─── Full lifecycle ───────────────────────────────────────────────────────────

func TestUser_FullLifecycle_ActiveToSuspendedToActive(t *testing.T) {
	u := newValidUser(t)
	t1 := testUserNow.Add(1 * time.Minute)
	t2 := testUserNow.Add(2 * time.Minute)
	t3 := testUserNow.Add(3 * time.Minute)

	if err := u.Activate(t1); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if err := u.Suspend(t2); err != nil {
		t.Fatalf("Suspend: %v", err)
	}
	if err := u.Activate(t3); err != nil {
		t.Fatalf("re-Activate: %v", err)
	}
	if u.Status != StatusActive {
		t.Errorf("final Status: got %q, want Active", u.Status)
	}
	if !u.UpdatedAt.Equal(t3) {
		t.Errorf("UpdatedAt: got %v, want %v", u.UpdatedAt, t3)
	}
}

func TestUser_FullLifecycle_DeactivatedIsTerminal(t *testing.T) {
	u := newValidUser(t)
	later := testUserNow.Add(time.Minute)
	_ = u.Activate(later)
	_ = u.Deactivate(later)

	// None of the other transitions may succeed from Deactivated.
	if err := u.Activate(later); err == nil {
		t.Error("Activate from Deactivated should fail")
	}
	if err := u.Suspend(later); err == nil {
		t.Error("Suspend from Deactivated should fail")
	}
	if err := u.Deactivate(later); err == nil {
		t.Error("Deactivate from Deactivated should fail")
	}
}

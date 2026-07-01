package entity_test

import (
	"testing"
	"time"

	"github.com/fairride/identity/domain/entity"
	"github.com/fairride/shared/errors"
)

var testNow = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

func TestNewPermission_Valid(t *testing.T) {
	p, err := entity.NewPermission("perm-1", "trips:read", "Read access to trips", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != "perm-1" {
		t.Errorf("ID: got %q, want %q", p.ID, "perm-1")
	}
	if p.Name != "trips:read" {
		t.Errorf("Name: got %q, want %q", p.Name, "trips:read")
	}
	if p.Resource != "trips" {
		t.Errorf("Resource: got %q, want %q", p.Resource, "trips")
	}
	if p.Action != "read" {
		t.Errorf("Action: got %q, want %q", p.Action, "read")
	}
	if !p.CreatedAt.Equal(testNow) {
		t.Errorf("CreatedAt: got %v, want %v", p.CreatedAt, testNow)
	}
}

func TestNewPermission_EmptyID(t *testing.T) {
	_, err := entity.NewPermission("", "trips:read", "", testNow)
	if err == nil {
		t.Fatal("expected error for empty id, got nil")
	}
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", errors.GetCode(err))
	}
}

func TestNewPermission_MissingColon(t *testing.T) {
	_, err := entity.NewPermission("id", "tripsread", "", testNow)
	if err == nil {
		t.Fatal("expected error for missing colon, got nil")
	}
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", errors.GetCode(err))
	}
}

func TestNewPermission_EmptyResource(t *testing.T) {
	_, err := entity.NewPermission("id", ":read", "", testNow)
	if err == nil {
		t.Fatal("expected error for empty resource, got nil")
	}
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", errors.GetCode(err))
	}
}

func TestNewPermission_EmptyAction(t *testing.T) {
	_, err := entity.NewPermission("id", "trips:", "", testNow)
	if err == nil {
		t.Fatal("expected error for empty action, got nil")
	}
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", errors.GetCode(err))
	}
}

func TestNewPermission_ColonInAction(t *testing.T) {
	// Extra colons are allowed — only the first split matters.
	p, err := entity.NewPermission("id", "admin:read:all", "", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Resource != "admin" {
		t.Errorf("Resource: got %q, want %q", p.Resource, "admin")
	}
	if p.Action != "read:all" {
		t.Errorf("Action: got %q, want %q", p.Action, "read:all")
	}
}

func TestReconstitutePermission(t *testing.T) {
	p := entity.ReconstitutePermission("id", "drivers:write", "drivers", "write", "desc", testNow)
	if p.ID != "id" || p.Resource != "drivers" || p.Action != "write" {
		t.Errorf("unexpected reconstituted permission: %+v", p)
	}
}

func TestPermissionConstants_ValidFormat(t *testing.T) {
	constants := []string{
		entity.PermTripsRead, entity.PermTripsWrite, entity.PermTripsManage,
		entity.PermDriversRead, entity.PermDriversWrite, entity.PermDriversManage,
		entity.PermRidersRead, entity.PermRidersWrite, entity.PermRidersManage,
		entity.PermWalletRead, entity.PermWalletWrite,
		entity.PermPaymentsRead, entity.PermPaymentsWrite,
		entity.PermDispatchRead, entity.PermDispatchWrite,
		entity.PermReviewsRead, entity.PermReviewsWrite,
		entity.PermReportsRead,
		entity.PermSupportRead, entity.PermSupportWrite,
		entity.PermAdminRead, entity.PermAdminWrite, entity.PermAdminManage,
	}

	for _, name := range constants {
		_, err := entity.NewPermission("test-id", name, "", testNow)
		if err != nil {
			t.Errorf("constant %q failed validation: %v", name, err)
		}
	}
}

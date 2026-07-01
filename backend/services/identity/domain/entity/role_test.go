package entity_test

import (
	"testing"

	"github.com/fairride/identity/domain/entity"
	"github.com/fairride/shared/errors"
)

func makePermission(id, name string) entity.Permission {
	p, err := entity.NewPermission(id, name, "", testNow)
	if err != nil {
		panic("makePermission: " + err.Error())
	}
	return *p
}

// ─── NewRole ─────────────────────────────────────────────────────────────────

func TestNewRole_Valid(t *testing.T) {
	r, err := entity.NewRole("role-1", "rider", "Standard rider", false, testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ID != "role-1" {
		t.Errorf("ID: got %q, want %q", r.ID, "role-1")
	}
	if r.Name != "rider" {
		t.Errorf("Name: got %q, want %q", r.Name, "rider")
	}
	if r.IsSystem {
		t.Error("IsSystem: got true, want false")
	}
	if r.PermissionCount() != 0 {
		t.Errorf("PermissionCount: got %d, want 0", r.PermissionCount())
	}
}

func TestNewRole_EmptyID(t *testing.T) {
	_, err := entity.NewRole("", "rider", "", false, testNow)
	if err == nil {
		t.Fatal("expected error for empty id, got nil")
	}
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", errors.GetCode(err))
	}
}

func TestNewRole_EmptyName(t *testing.T) {
	_, err := entity.NewRole("id", "", "", false, testNow)
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", errors.GetCode(err))
	}
}

func TestNewRole_SystemFlag(t *testing.T) {
	r, err := entity.NewRole("id", entity.RoleSuperAdmin, "Super admin", true, testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.IsSystem {
		t.Error("IsSystem: got false, want true")
	}
}

// ─── AddPermission / RemovePermission ────────────────────────────────────────

func TestRole_AddPermission(t *testing.T) {
	r, _ := entity.NewRole("role-1", "rider", "", false, testNow)
	p := makePermission("perm-1", entity.PermTripsRead)

	r.AddPermission(p)

	if r.PermissionCount() != 1 {
		t.Errorf("PermissionCount: got %d, want 1", r.PermissionCount())
	}
}

func TestRole_AddPermission_Idempotent(t *testing.T) {
	r, _ := entity.NewRole("role-1", "rider", "", false, testNow)
	p := makePermission("perm-1", entity.PermTripsRead)

	r.AddPermission(p)
	r.AddPermission(p) // add same permission twice

	if r.PermissionCount() != 1 {
		t.Errorf("PermissionCount after double add: got %d, want 1", r.PermissionCount())
	}
}

func TestRole_RemovePermission(t *testing.T) {
	r, _ := entity.NewRole("role-1", "rider", "", false, testNow)
	p := makePermission("perm-1", entity.PermTripsRead)
	r.AddPermission(p)

	r.RemovePermission("perm-1")

	if r.PermissionCount() != 0 {
		t.Errorf("PermissionCount after remove: got %d, want 0", r.PermissionCount())
	}
}

func TestRole_RemovePermission_NoOp(t *testing.T) {
	r, _ := entity.NewRole("role-1", "rider", "", false, testNow)

	// removing a non-existent permission must not panic or error
	r.RemovePermission("does-not-exist")

	if r.PermissionCount() != 0 {
		t.Errorf("PermissionCount: got %d, want 0", r.PermissionCount())
	}
}

// ─── HasPermission ───────────────────────────────────────────────────────────

func TestRole_HasPermission_Present(t *testing.T) {
	r, _ := entity.NewRole("role-1", "rider", "", false, testNow)
	r.AddPermission(makePermission("perm-1", entity.PermTripsRead))

	if !r.HasPermission(entity.PermTripsRead) {
		t.Error("HasPermission: expected true for granted permission")
	}
}

func TestRole_HasPermission_NotPresent(t *testing.T) {
	r, _ := entity.NewRole("role-1", "rider", "", false, testNow)

	if r.HasPermission(entity.PermTripsRead) {
		t.Error("HasPermission: expected false for ungrant permission")
	}
}

func TestRole_HasPermission_AfterRemove(t *testing.T) {
	r, _ := entity.NewRole("role-1", "rider", "", false, testNow)
	p := makePermission("perm-1", entity.PermTripsRead)
	r.AddPermission(p)
	r.RemovePermission("perm-1")

	if r.HasPermission(entity.PermTripsRead) {
		t.Error("HasPermission: expected false after remove")
	}
}

// ─── Permissions snapshot ────────────────────────────────────────────────────

func TestRole_Permissions_ReturnsCopy(t *testing.T) {
	r, _ := entity.NewRole("role-1", "rider", "", false, testNow)
	r.AddPermission(makePermission("p1", entity.PermTripsRead))
	r.AddPermission(makePermission("p2", entity.PermTripsWrite))

	snapshot := r.Permissions()
	if len(snapshot) != 2 {
		t.Fatalf("len(Permissions): got %d, want 2", len(snapshot))
	}

	// Mutating the snapshot must not affect the role.
	snapshot[0] = entity.Permission{ID: "injected"}
	if r.HasPermission("") {
		t.Error("Permissions snapshot mutation leaked into role")
	}
	if r.PermissionCount() != 2 {
		t.Errorf("PermissionCount after snapshot mutation: got %d, want 2", r.PermissionCount())
	}
}

// ─── CanDelete ───────────────────────────────────────────────────────────────

func TestRole_CanDelete_NonSystem(t *testing.T) {
	r, _ := entity.NewRole("role-1", "custom", "", false, testNow)
	if err := r.CanDelete(); err != nil {
		t.Errorf("CanDelete: expected nil for non-system role, got %v", err)
	}
}

func TestRole_CanDelete_System(t *testing.T) {
	r, _ := entity.NewRole("role-1", entity.RoleSuperAdmin, "", true, testNow)
	err := r.CanDelete()
	if err == nil {
		t.Fatal("CanDelete: expected error for system role, got nil")
	}
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("expected CodePreconditionFailed, got %v", errors.GetCode(err))
	}
}

// ─── ReconstituteRole ────────────────────────────────────────────────────────

func TestReconstituteRole_RestoresPermissions(t *testing.T) {
	perms := []entity.Permission{
		makePermission("p1", entity.PermTripsRead),
		makePermission("p2", entity.PermDriversRead),
	}
	r := entity.ReconstituteRole("role-1", "rider", "Rider role", true, perms, testNow, testNow)

	if r.PermissionCount() != 2 {
		t.Errorf("PermissionCount: got %d, want 2", r.PermissionCount())
	}
	if !r.HasPermission(entity.PermTripsRead) {
		t.Error("missing PermTripsRead after reconstitution")
	}
	if !r.HasPermission(entity.PermDriversRead) {
		t.Error("missing PermDriversRead after reconstitution")
	}
}

func TestReconstituteRole_EmptyPermissions(t *testing.T) {
	r := entity.ReconstituteRole("role-1", "rider", "", false, nil, testNow, testNow)
	if r.PermissionCount() != 0 {
		t.Errorf("PermissionCount: got %d, want 0", r.PermissionCount())
	}
}

// ─── System role constants ───────────────────────────────────────────────────

func TestSystemRoleConstants_NonEmpty(t *testing.T) {
	roles := []string{
		entity.RoleRider,
		entity.RoleDriver,
		entity.RoleFleetOperator,
		entity.RoleCityManager,
		entity.RoleSupportAgent,
		entity.RoleSuperAdmin,
	}
	for _, name := range roles {
		if name == "" {
			t.Errorf("system role constant is empty")
		}
		_, err := entity.NewRole("id", name, "", true, testNow)
		if err != nil {
			t.Errorf("system role %q failed NewRole: %v", name, err)
		}
	}
}

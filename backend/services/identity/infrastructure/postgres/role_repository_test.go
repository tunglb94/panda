package postgres_test

import (
	"context"
	"testing"

	"github.com/fairride/identity/domain/entity"
	"github.com/fairride/identity/infrastructure/postgres"
	domainerrors "github.com/fairride/shared/errors"
)

func newRoleRepo() *postgres.RoleRepository {
	return postgres.NewRoleRepository(testPool)
}

func makeTestRole(id, name string) *entity.Role {
	r, err := entity.NewRole(id, name, "test role "+name, false, testNow)
	if err != nil {
		panic("makeTestRole: " + err.Error())
	}
	return r
}

// savePerm is a convenience that saves a permission and fatals on error.
func savePerm(t *testing.T, repo *postgres.PermissionRepository, id, name string) *entity.Permission {
	t.Helper()
	p := makeTestPermission(id, name)
	if err := repo.Save(context.Background(), p); err != nil {
		t.Fatalf("savePerm %s: %v", name, err)
	}
	return p
}

// ─── Save ────────────────────────────────────────────────────────────────────

func TestRoleRepository_Save_New(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	roleRepo := newRoleRepo()

	role := makeTestRole("role-1", entity.RoleRider)
	if err := roleRepo.Save(ctx, role); err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := roleRepo.FindByID(ctx, "role-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != entity.RoleRider {
		t.Errorf("Name: got %q, want %q", found.Name, entity.RoleRider)
	}
	if found.IsSystem {
		t.Error("IsSystem: got true, want false")
	}
	if found.PermissionCount() != 0 {
		t.Errorf("PermissionCount: got %d, want 0", found.PermissionCount())
	}
}

func TestRoleRepository_Save_SystemRole(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	roleRepo := newRoleRepo()

	role, _ := entity.NewRole("role-sys", entity.RoleSuperAdmin, "super admin", true, testNow)
	if err := roleRepo.Save(ctx, role); err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := roleRepo.FindByID(ctx, "role-sys")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if !found.IsSystem {
		t.Error("IsSystem: got false, want true")
	}
}

func TestRoleRepository_Save_WithPermissions(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	permRepo := newPermRepo()
	roleRepo := newRoleRepo()

	// Permissions must exist before they can be linked to a role (FK constraint).
	tripsRead := savePerm(t, permRepo, "perm-tr", entity.PermTripsRead)
	driversRead := savePerm(t, permRepo, "perm-dr", entity.PermDriversRead)

	role := makeTestRole("role-1", entity.RoleRider)
	role.AddPermission(*tripsRead)
	role.AddPermission(*driversRead)

	if err := roleRepo.Save(ctx, role); err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := roleRepo.FindByID(ctx, "role-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.PermissionCount() != 2 {
		t.Errorf("PermissionCount: got %d, want 2", found.PermissionCount())
	}
	if !found.HasPermission(entity.PermTripsRead) {
		t.Error("expected trips:read permission")
	}
	if !found.HasPermission(entity.PermDriversRead) {
		t.Error("expected drivers:read permission")
	}
}

func TestRoleRepository_Save_ReplacesPermissions(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	permRepo := newPermRepo()
	roleRepo := newRoleRepo()

	perm1 := savePerm(t, permRepo, "perm-1", entity.PermTripsRead)
	perm2 := savePerm(t, permRepo, "perm-2", entity.PermTripsWrite)

	// First save: role has perm1.
	role := makeTestRole("role-1", entity.RoleRider)
	role.AddPermission(*perm1)
	if err := roleRepo.Save(ctx, role); err != nil {
		t.Fatalf("initial Save: %v", err)
	}

	// Second save: role now has only perm2 — perm1 must be removed.
	role.RemovePermission("perm-1")
	role.AddPermission(*perm2)
	if err := roleRepo.Save(ctx, role); err != nil {
		t.Fatalf("update Save: %v", err)
	}

	found, err := roleRepo.FindByID(ctx, "role-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.PermissionCount() != 1 {
		t.Errorf("PermissionCount: got %d, want 1", found.PermissionCount())
	}
	if found.HasPermission(entity.PermTripsRead) {
		t.Error("trips:read should have been removed")
	}
	if !found.HasPermission(entity.PermTripsWrite) {
		t.Error("expected trips:write permission")
	}
}

func TestRoleRepository_Save_DuplicateName(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	roleRepo := newRoleRepo()

	r1 := makeTestRole("role-1", entity.RoleRider)
	if err := roleRepo.Save(ctx, r1); err != nil {
		t.Fatalf("Save r1: %v", err)
	}

	// Different ID, same name → unique violation.
	r2 := makeTestRole("role-2", entity.RoleRider)
	err := roleRepo.Save(ctx, r2)
	if err == nil {
		t.Fatal("expected AlreadyExists, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeAlreadyExists) {
		t.Errorf("expected CodeAlreadyExists, got %v", domainerrors.GetCode(err))
	}
}

// ─── FindByID ────────────────────────────────────────────────────────────────

func TestRoleRepository_FindByID_NotFound(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	roleRepo := newRoleRepo()

	_, err := roleRepo.FindByID(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected NotFound, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

// ─── FindByName ──────────────────────────────────────────────────────────────

func TestRoleRepository_FindByName(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	roleRepo := newRoleRepo()

	role := makeTestRole("role-1", entity.RoleDriver)
	if err := roleRepo.Save(ctx, role); err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := roleRepo.FindByName(ctx, entity.RoleDriver)
	if err != nil {
		t.Fatalf("FindByName: %v", err)
	}
	if found.ID != "role-1" {
		t.Errorf("ID: got %q, want %q", found.ID, "role-1")
	}
}

func TestRoleRepository_FindByName_NotFound(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	roleRepo := newRoleRepo()

	_, err := roleRepo.FindByName(ctx, "nonexistent-role")
	if err == nil {
		t.Fatal("expected NotFound, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

// ─── FindAll ─────────────────────────────────────────────────────────────────

func TestRoleRepository_FindAll(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	roleRepo := newRoleRepo()

	for _, name := range []string{entity.RoleRider, entity.RoleDriver, entity.RoleCityManager} {
		if err := roleRepo.Save(ctx, makeTestRole("role-"+name, name)); err != nil {
			t.Fatalf("Save %s: %v", name, err)
		}
	}

	all, err := roleRepo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("len: got %d, want 3", len(all))
	}
}

func TestRoleRepository_FindAll_IncludesPermissions(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	permRepo := newPermRepo()
	roleRepo := newRoleRepo()

	perm := savePerm(t, permRepo, "perm-1", entity.PermTripsRead)

	role := makeTestRole("role-1", entity.RoleRider)
	role.AddPermission(*perm)
	if err := roleRepo.Save(ctx, role); err != nil {
		t.Fatalf("Save: %v", err)
	}

	all, err := roleRepo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("len: got %d, want 1", len(all))
	}
	if !all[0].HasPermission(entity.PermTripsRead) {
		t.Error("expected trips:read in FindAll result")
	}
}

func TestRoleRepository_FindAll_Empty(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	roleRepo := newRoleRepo()

	all, err := roleRepo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected empty slice, got %d", len(all))
	}
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func TestRoleRepository_Delete(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	roleRepo := newRoleRepo()

	role := makeTestRole("role-1", entity.RoleRider)
	if err := roleRepo.Save(ctx, role); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := roleRepo.Delete(ctx, "role-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := roleRepo.FindByID(ctx, "role-1")
	if err == nil {
		t.Fatal("expected NotFound after delete, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

func TestRoleRepository_Delete_CascadesPermissions(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	permRepo := newPermRepo()
	roleRepo := newRoleRepo()

	perm := savePerm(t, permRepo, "perm-1", entity.PermTripsRead)
	role := makeTestRole("role-1", entity.RoleRider)
	role.AddPermission(*perm)
	if err := roleRepo.Save(ctx, role); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Deleting the role should cascade-delete identity_role_permissions rows.
	if err := roleRepo.Delete(ctx, "role-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// The standalone permission itself must still exist.
	if _, err := permRepo.FindByID(ctx, "perm-1"); err != nil {
		t.Errorf("permission should still exist after role delete: %v", err)
	}
}

func TestRoleRepository_Delete_NotFound(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	roleRepo := newRoleRepo()

	err := roleRepo.Delete(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected NotFound, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

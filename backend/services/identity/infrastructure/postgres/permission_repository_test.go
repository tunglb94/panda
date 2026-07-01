package postgres_test

import (
	"context"
	"testing"

	"github.com/fairride/identity/domain/entity"
	"github.com/fairride/identity/infrastructure/postgres"
	domainerrors "github.com/fairride/shared/errors"
)

func newPermRepo() *postgres.PermissionRepository {
	return postgres.NewPermissionRepository(testPool)
}

func makeTestPermission(id, name string) *entity.Permission {
	p, err := entity.NewPermission(id, name, "test permission "+name, testNow)
	if err != nil {
		panic("makeTestPermission: " + err.Error())
	}
	return p
}

// ─── Save ────────────────────────────────────────────────────────────────────

func TestPermissionRepository_Save_New(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newPermRepo()

	p := makeTestPermission("perm-1", entity.PermTripsRead)
	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := repo.FindByID(ctx, "perm-1")
	if err != nil {
		t.Fatalf("FindByID after save: %v", err)
	}
	if found.Name != entity.PermTripsRead {
		t.Errorf("Name: got %q, want %q", found.Name, entity.PermTripsRead)
	}
	if found.Resource != "trips" {
		t.Errorf("Resource: got %q, want %q", found.Resource, "trips")
	}
	if found.Action != "read" {
		t.Errorf("Action: got %q, want %q", found.Action, "read")
	}
}

func TestPermissionRepository_Save_Update(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newPermRepo()

	p := makeTestPermission("perm-1", entity.PermTripsRead)
	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("initial Save: %v", err)
	}

	// Update description by re-saving with same ID but new description.
	updated, _ := entity.NewPermission("perm-1", entity.PermTripsWrite, "updated description", testNow)
	if err := repo.Save(ctx, updated); err != nil {
		t.Fatalf("update Save: %v", err)
	}

	found, err := repo.FindByID(ctx, "perm-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != entity.PermTripsWrite {
		t.Errorf("Name after update: got %q, want %q", found.Name, entity.PermTripsWrite)
	}
	if found.Description != "updated description" {
		t.Errorf("Description: got %q, want %q", found.Description, "updated description")
	}
}

func TestPermissionRepository_Save_DuplicateName(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newPermRepo()

	p1 := makeTestPermission("perm-1", entity.PermTripsRead)
	if err := repo.Save(ctx, p1); err != nil {
		t.Fatalf("Save p1: %v", err)
	}

	// Different ID, same name → unique constraint violation.
	p2 := makeTestPermission("perm-2", entity.PermTripsRead)
	err := repo.Save(ctx, p2)
	if err == nil {
		t.Fatal("expected AlreadyExists error, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeAlreadyExists) {
		t.Errorf("expected CodeAlreadyExists, got %v", domainerrors.GetCode(err))
	}
}

// ─── FindByID ────────────────────────────────────────────────────────────────

func TestPermissionRepository_FindByID_NotFound(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newPermRepo()

	_, err := repo.FindByID(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected NotFound error, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

// ─── FindByName ──────────────────────────────────────────────────────────────

func TestPermissionRepository_FindByName(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newPermRepo()

	p := makeTestPermission("perm-1", entity.PermDriversRead)
	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := repo.FindByName(ctx, entity.PermDriversRead)
	if err != nil {
		t.Fatalf("FindByName: %v", err)
	}
	if found.ID != "perm-1" {
		t.Errorf("ID: got %q, want %q", found.ID, "perm-1")
	}
}

func TestPermissionRepository_FindByName_NotFound(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newPermRepo()

	_, err := repo.FindByName(ctx, "nonexistent:perm")
	if err == nil {
		t.Fatal("expected NotFound error, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

// ─── FindByResource ──────────────────────────────────────────────────────────

func TestPermissionRepository_FindByResource(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newPermRepo()

	// Save three trip permissions and one driver permission.
	for _, perm := range []struct{ id, name string }{
		{"p1", entity.PermTripsRead},
		{"p2", entity.PermTripsWrite},
		{"p3", entity.PermTripsManage},
		{"p4", entity.PermDriversRead},
	} {
		if err := repo.Save(ctx, makeTestPermission(perm.id, perm.name)); err != nil {
			t.Fatalf("Save %s: %v", perm.name, err)
		}
	}

	results, err := repo.FindByResource(ctx, "trips")
	if err != nil {
		t.Fatalf("FindByResource: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("len: got %d, want 3", len(results))
	}
	for _, p := range results {
		if p.Resource != "trips" {
			t.Errorf("unexpected resource %q in result", p.Resource)
		}
	}
}

func TestPermissionRepository_FindByResource_Empty(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newPermRepo()

	results, err := repo.FindByResource(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("FindByResource: unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty slice, got %d results", len(results))
	}
}

// ─── FindAll ─────────────────────────────────────────────────────────────────

func TestPermissionRepository_FindAll(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newPermRepo()

	perms := []struct{ id, name string }{
		{"p1", entity.PermTripsRead},
		{"p2", entity.PermDriversRead},
		{"p3", entity.PermWalletRead},
	}
	for _, perm := range perms {
		if err := repo.Save(ctx, makeTestPermission(perm.id, perm.name)); err != nil {
			t.Fatalf("Save %s: %v", perm.name, err)
		}
	}

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("len: got %d, want 3", len(all))
	}
}

func TestPermissionRepository_FindAll_Empty(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newPermRepo()

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll empty: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected empty slice, got %d", len(all))
	}
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func TestPermissionRepository_Delete(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newPermRepo()

	p := makeTestPermission("perm-1", entity.PermTripsRead)
	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := repo.Delete(ctx, "perm-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, "perm-1")
	if err == nil {
		t.Fatal("expected NotFound after delete, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

func TestPermissionRepository_Delete_NotFound(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newPermRepo()

	err := repo.Delete(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected NotFound error, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

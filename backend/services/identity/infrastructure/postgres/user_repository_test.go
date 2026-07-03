package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/fairride/identity/domain/entity"
	"github.com/fairride/identity/infrastructure/postgres"
	domainerrors "github.com/fairride/shared/errors"
)

func newUserRepo() *postgres.UserRepository {
	return postgres.NewUserRepository(testPool)
}

func makeTestUser(id, phone string) *entity.User {
	u, err := entity.NewUser(id, phone, "Test User", "", entity.TypeRider, "role-rider", testNow)
	if err != nil {
		panic("makeTestUser: " + err.Error())
	}
	return u
}

// ─── Save (Create) ────────────────────────────────────────────────────────────

func TestUserRepository_Save_New(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newUserRepo()

	u := makeTestUser("user-1", "+84901111111")
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := repo.FindByID(ctx, "user-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.ID != "user-1" {
		t.Errorf("ID: got %q, want %q", found.ID, "user-1")
	}
	if found.PhoneNumber != "+84901111111" {
		t.Errorf("PhoneNumber: got %q", found.PhoneNumber)
	}
	if found.Name != "Test User" {
		t.Errorf("Name: got %q", found.Name)
	}
	if found.Type != entity.TypeRider {
		t.Errorf("Type: got %q, want %q", found.Type, entity.TypeRider)
	}
	if found.Status != entity.StatusPendingVerification {
		t.Errorf("Status: got %q, want %q", found.Status, entity.StatusPendingVerification)
	}
	if found.RoleID != "role-rider" {
		t.Errorf("RoleID: got %q", found.RoleID)
	}
}

func TestUserRepository_Save_WithEmail(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newUserRepo()

	u, err := entity.NewUser("user-1", "+84901111111", "Alex", "alex@example.com", entity.TypeRider, "role-1", testNow)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := repo.FindByID(ctx, "user-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Email != "alex@example.com" {
		t.Errorf("Email: got %q, want %q", found.Email, "alex@example.com")
	}
}

func TestUserRepository_Save_WithoutEmail(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newUserRepo()

	u := makeTestUser("user-1", "+84901111111") // email is ""
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := repo.FindByID(ctx, "user-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Email != "" {
		t.Errorf("Email: got %q, want empty string", found.Email)
	}
}

func TestUserRepository_Save_DuplicatePhone(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newUserRepo()

	u1 := makeTestUser("user-1", "+84901111111")
	if err := repo.Save(ctx, u1); err != nil {
		t.Fatalf("Save u1: %v", err)
	}

	// Different ID, same phone number → unique constraint violation.
	u2 := makeTestUser("user-2", "+84901111111")
	err := repo.Save(ctx, u2)
	if err == nil {
		t.Fatal("expected AlreadyExists, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeAlreadyExists) {
		t.Errorf("expected CodeAlreadyExists, got %v", domainerrors.GetCode(err))
	}
}

// ─── Save (Update) ───────────────────────────────────────────────────────────

func TestUserRepository_Save_UpdateStatus(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newUserRepo()

	u := makeTestUser("user-1", "+84901111111")
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("initial Save: %v", err)
	}

	// Activate via domain method, then persist.
	activatedAt := testNow.Add(time.Minute)
	if err := u.Activate(activatedAt); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("Save after Activate: %v", err)
	}

	found, err := repo.FindByID(ctx, "user-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Status != entity.StatusActive {
		t.Errorf("Status: got %q, want %q", found.Status, entity.StatusActive)
	}
	if !found.UpdatedAt.Equal(activatedAt) {
		t.Errorf("UpdatedAt: got %v, want %v", found.UpdatedAt, activatedAt)
	}
}

func TestUserRepository_Save_StatusTransitions(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newUserRepo()

	u := makeTestUser("user-1", "+84901111111")
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("initial Save: %v", err)
	}

	t1 := testNow.Add(1 * time.Minute)
	t2 := testNow.Add(2 * time.Minute)
	t3 := testNow.Add(3 * time.Minute)

	// PendingVerification → Active → Suspended → Active
	_ = u.Activate(t1)
	_ = repo.Save(ctx, u)
	_ = u.Suspend(t2)
	_ = repo.Save(ctx, u)
	_ = u.Activate(t3)
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("Save after re-Activate: %v", err)
	}

	found, err := repo.FindByID(ctx, "user-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Status != entity.StatusActive {
		t.Errorf("Status: got %q, want %q", found.Status, entity.StatusActive)
	}
	if !found.UpdatedAt.Equal(t3) {
		t.Errorf("UpdatedAt: got %v, want %v", found.UpdatedAt, t3)
	}
}

func TestUserRepository_Save_CreatedAtImmutable(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newUserRepo()

	u := makeTestUser("user-1", "+84901111111")
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("initial Save: %v", err)
	}

	// Re-save with same ID but the entity has the same createdAt.
	later := testNow.Add(time.Hour)
	_ = u.Activate(later)
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("update Save: %v", err)
	}

	found, err := repo.FindByID(ctx, "user-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if !found.CreatedAt.Equal(testNow) {
		t.Errorf("CreatedAt mutated: got %v, want %v", found.CreatedAt, testNow)
	}
}

// ─── FindByID ────────────────────────────────────────────────────────────────

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newUserRepo()

	_, err := repo.FindByID(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected NotFound, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

// ─── FindByPhone ─────────────────────────────────────────────────────────────

func TestUserRepository_FindByPhone(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newUserRepo()

	u := makeTestUser("user-1", "+84901234567")
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := repo.FindByPhone(ctx, "+84901234567")
	if err != nil {
		t.Fatalf("FindByPhone: %v", err)
	}
	if found.ID != "user-1" {
		t.Errorf("ID: got %q, want %q", found.ID, "user-1")
	}
}

func TestUserRepository_FindByPhone_NotFound(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newUserRepo()

	_, err := repo.FindByPhone(ctx, "+84999999999")
	if err == nil {
		t.Fatal("expected NotFound, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

// ─── FindAll ─────────────────────────────────────────────────────────────────

func TestUserRepository_FindAll(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newUserRepo()

	phones := []struct{ id, phone string }{
		{"user-1", "+84901111111"},
		{"user-2", "+84902222222"},
		{"user-3", "+84903333333"},
	}
	for _, p := range phones {
		if err := repo.Save(ctx, makeTestUser(p.id, p.phone)); err != nil {
			t.Fatalf("Save %s: %v", p.id, err)
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

func TestUserRepository_FindAll_Empty(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newUserRepo()

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected empty slice, got %d", len(all))
	}
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func TestUserRepository_Delete(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newUserRepo()

	u := makeTestUser("user-1", "+84901111111")
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := repo.Delete(ctx, "user-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, "user-1")
	if err == nil {
		t.Fatal("expected NotFound after delete, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

func TestUserRepository_Delete_NotFound(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newUserRepo()

	err := repo.Delete(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected NotFound, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

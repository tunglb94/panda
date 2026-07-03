package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/fairride/user/domain/entity"
	"github.com/fairride/user/infrastructure/postgres"
	domainerrors "github.com/fairride/shared/errors"
)

func newRepo() *postgres.ProfileRepository {
	return postgres.NewProfileRepository(testPool)
}

func makeProfile(id, fullName, phone string) *entity.UserProfile {
	p, err := entity.NewUserProfile(id, fullName, phone, "", "", time.Time{}, entity.GenderUnspecified, testNow)
	if err != nil {
		panic("makeProfile: " + err.Error())
	}
	return p
}

// ─── Save (Create) ────────────────────────────────────────────────────────────

func TestProfileRepository_Save_New(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newRepo()

	p := makeProfile("u-1", "Nguyen Van A", "+84901234567")
	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := repo.FindByID(ctx, "u-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.FullName != "Nguyen Van A" {
		t.Errorf("FullName: got %q", found.FullName)
	}
	if found.Phone != "+84901234567" {
		t.Errorf("Phone: got %q", found.Phone)
	}
	if found.Status != entity.ProfileStatusActive {
		t.Errorf("Status: got %q", found.Status)
	}
	if found.Gender != entity.GenderUnspecified {
		t.Errorf("Gender: got %q", found.Gender)
	}
	if !found.DateOfBirth.IsZero() {
		t.Errorf("DateOfBirth: expected zero, got %v", found.DateOfBirth)
	}
}

func TestProfileRepository_Save_WithAllFields(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newRepo()

	dob := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	p, err := entity.NewUserProfile(
		"u-1", "Ahmad", "+601112345678",
		"ahmad@example.com", "https://cdn.example.com/avatar.jpg",
		dob, entity.GenderMale, testNow,
	)
	if err != nil {
		t.Fatalf("NewUserProfile: %v", err)
	}
	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := repo.FindByID(ctx, "u-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Email != "ahmad@example.com" {
		t.Errorf("Email: got %q", found.Email)
	}
	if found.Avatar != "https://cdn.example.com/avatar.jpg" {
		t.Errorf("Avatar: got %q", found.Avatar)
	}
	if !found.DateOfBirth.Equal(dob) {
		t.Errorf("DateOfBirth: got %v, want %v", found.DateOfBirth, dob)
	}
	if found.Gender != entity.GenderMale {
		t.Errorf("Gender: got %q", found.Gender)
	}
}

// ─── Save (Update) ────────────────────────────────────────────────────────────

func TestProfileRepository_Save_Update(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newRepo()

	p := makeProfile("u-1", "Old Name", "+84901234567")
	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("initial Save: %v", err)
	}

	later := testNow.Add(time.Hour)
	if err := p.Update("New Name", "new@example.com", "", time.Time{}, entity.GenderFemale, later); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("update Save: %v", err)
	}

	found, err := repo.FindByID(ctx, "u-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.FullName != "New Name" {
		t.Errorf("FullName: got %q", found.FullName)
	}
	if found.Email != "new@example.com" {
		t.Errorf("Email: got %q", found.Email)
	}
	if found.Gender != entity.GenderFemale {
		t.Errorf("Gender: got %q", found.Gender)
	}
}

func TestProfileRepository_Save_CreatedAtImmutable(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newRepo()

	p := makeProfile("u-1", "Name", "+1")
	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("initial Save: %v", err)
	}

	_ = p.Update("Name2", "", "", time.Time{}, entity.GenderUnspecified, testNow.Add(time.Hour))
	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("update Save: %v", err)
	}

	found, err := repo.FindByID(ctx, "u-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if !found.CreatedAt.Equal(testNow) {
		t.Errorf("CreatedAt mutated: got %v, want %v", found.CreatedAt, testNow)
	}
}

func TestProfileRepository_Save_SetAndClearDateOfBirth(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newRepo()

	p := makeProfile("u-1", "Name", "+1")
	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("initial Save: %v", err)
	}

	// Set DOB.
	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	_ = p.Update("Name", "", "", dob, entity.GenderUnspecified, testNow.Add(time.Minute))
	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("Save with dob: %v", err)
	}
	found, _ := repo.FindByID(ctx, "u-1")
	if !found.DateOfBirth.Equal(dob) {
		t.Errorf("DateOfBirth: got %v, want %v", found.DateOfBirth, dob)
	}

	// Clear DOB.
	_ = p.Update("Name", "", "", time.Time{}, entity.GenderUnspecified, testNow.Add(time.Hour))
	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("Save clearing dob: %v", err)
	}
	found, _ = repo.FindByID(ctx, "u-1")
	if !found.DateOfBirth.IsZero() {
		t.Errorf("DateOfBirth should be zero after clear, got %v", found.DateOfBirth)
	}
}

// ─── FindByID ─────────────────────────────────────────────────────────────────

func TestProfileRepository_FindByID_NotFound(t *testing.T) {
	setupTest(t)
	ctx := context.Background()
	repo := newRepo()

	_, err := repo.FindByID(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected NotFound, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

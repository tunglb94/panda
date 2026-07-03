package app

import (
	"context"
	"testing"
	"time"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/user/domain/entity"
)

// ─── In-memory stub ──────────────────────────────────────────────────────────

// stubProfileRepo is a simple in-memory ProfileRepository for use-case tests.
type stubProfileRepo struct {
	store map[string]*entity.UserProfile
	// saveErr, if set, is returned by Save.
	saveErr error
}

func newStubRepo() *stubProfileRepo {
	return &stubProfileRepo{store: make(map[string]*entity.UserProfile)}
}

func (r *stubProfileRepo) FindByID(_ context.Context, id string) (*entity.UserProfile, error) {
	p, ok := r.store[id]
	if !ok {
		return nil, domainerrors.NotFound("profile not found: " + id)
	}
	// Return a copy so mutations in tests don't affect the store unless Save is called.
	cp := *p
	return &cp, nil
}

func (r *stubProfileRepo) Save(_ context.Context, p *entity.UserProfile) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	cp := *p
	r.store[p.ID] = &cp
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

var testNow = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

func seedProfile(t *testing.T, repo *stubProfileRepo, id, fullName, phone string) *entity.UserProfile {
	t.Helper()
	p, err := entity.NewUserProfile(id, fullName, phone, "", "", time.Time{}, entity.GenderUnspecified, testNow)
	if err != nil {
		t.Fatalf("seedProfile: %v", err)
	}
	repo.store[id] = p
	return p
}

// ─── GetProfileUseCase ───────────────────────────────────────────────────────

func TestGetProfile_Success(t *testing.T) {
	repo := newStubRepo()
	seedProfile(t, repo, "u-1", "Nguyen Van A", "+84901234567")
	uc := NewGetProfileUseCase(repo)

	profile, err := uc.Execute(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if profile.ID != "u-1" {
		t.Errorf("ID: got %q, want %q", profile.ID, "u-1")
	}
	if profile.FullName != "Nguyen Van A" {
		t.Errorf("FullName: got %q", profile.FullName)
	}
}

func TestGetProfile_EmptyUserID(t *testing.T) {
	uc := NewGetProfileUseCase(newStubRepo())
	_, err := uc.Execute(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty user id")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	uc := NewGetProfileUseCase(newStubRepo())
	_, err := uc.Execute(context.Background(), "does-not-exist")
	if err == nil {
		t.Fatal("expected NotFound, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

func TestGetProfile_NilRepo_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil repo")
		}
	}()
	NewGetProfileUseCase(nil)
}

// ─── UpdateProfileUseCase ────────────────────────────────────────────────────

func TestUpdateProfile_Success(t *testing.T) {
	repo := newStubRepo()
	seedProfile(t, repo, "u-1", "Old Name", "+84901234567")
	uc := NewUpdateProfileUseCase(repo)

	result, err := uc.Execute(context.Background(), UpdateProfileInput{
		UserID:   "u-1",
		FullName: "New Name",
		Email:    "new@example.com",
		Gender:   entity.GenderFemale,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.FullName != "New Name" {
		t.Errorf("FullName: got %q, want %q", result.FullName, "New Name")
	}
	if result.Email != "new@example.com" {
		t.Errorf("Email: got %q", result.Email)
	}
	if result.Gender != entity.GenderFemale {
		t.Errorf("Gender: got %q", result.Gender)
	}
}

func TestUpdateProfile_Persists(t *testing.T) {
	repo := newStubRepo()
	seedProfile(t, repo, "u-1", "Old Name", "+84901234567")
	uc := NewUpdateProfileUseCase(repo)

	_, err := uc.Execute(context.Background(), UpdateProfileInput{
		UserID:   "u-1",
		FullName: "Persisted Name",
		Gender:   entity.GenderOther,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Verify the change was persisted via a fresh FindByID.
	saved, err := repo.FindByID(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if saved.FullName != "Persisted Name" {
		t.Errorf("FullName in store: got %q", saved.FullName)
	}
}

func TestUpdateProfile_EmptyUserID(t *testing.T) {
	uc := NewUpdateProfileUseCase(newStubRepo())
	_, err := uc.Execute(context.Background(), UpdateProfileInput{UserID: "", FullName: "Name", Gender: entity.GenderUnspecified})
	if err == nil {
		t.Fatal("expected error for empty user id")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestUpdateProfile_NotFound(t *testing.T) {
	uc := NewUpdateProfileUseCase(newStubRepo())
	_, err := uc.Execute(context.Background(), UpdateProfileInput{
		UserID: "ghost", FullName: "Name", Gender: entity.GenderUnspecified,
	})
	if err == nil {
		t.Fatal("expected NotFound, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected CodeNotFound, got %v", domainerrors.GetCode(err))
	}
}

func TestUpdateProfile_ValidationError(t *testing.T) {
	repo := newStubRepo()
	seedProfile(t, repo, "u-1", "Name", "+1")
	uc := NewUpdateProfileUseCase(repo)

	_, err := uc.Execute(context.Background(), UpdateProfileInput{
		UserID:   "u-1",
		FullName: "", // invalid — empty
		Gender:   entity.GenderUnspecified,
	})
	if err == nil {
		t.Fatal("expected error for empty full name")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestUpdateProfile_WithDateOfBirth(t *testing.T) {
	repo := newStubRepo()
	seedProfile(t, repo, "u-1", "Name", "+1")
	uc := NewUpdateProfileUseCase(repo)

	dob := time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC)
	result, err := uc.Execute(context.Background(), UpdateProfileInput{
		UserID:      "u-1",
		FullName:    "Name",
		DateOfBirth: dob,
		Gender:      entity.GenderMale,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !result.DateOfBirth.Equal(dob) {
		t.Errorf("DateOfBirth: got %v, want %v", result.DateOfBirth, dob)
	}
}

func TestUpdateProfile_NilRepo_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil repo")
		}
	}()
	NewUpdateProfileUseCase(nil)
}

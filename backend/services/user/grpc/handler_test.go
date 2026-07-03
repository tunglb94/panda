package grpc

import (
	"context"
	"testing"
	"time"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/user/app"
	"github.com/fairride/user/domain/entity"
	"github.com/fairride/user/grpc/userpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─── stub repository ─────────────────────────────────────────────────────────

type stubRepo struct {
	store map[string]*entity.UserProfile
}

func newStubRepo() *stubRepo {
	return &stubRepo{store: make(map[string]*entity.UserProfile)}
}

func (r *stubRepo) FindByID(_ context.Context, id string) (*entity.UserProfile, error) {
	p, ok := r.store[id]
	if !ok {
		return nil, domainerrors.NotFound("profile not found: " + id)
	}
	cp := *p
	return &cp, nil
}

func (r *stubRepo) Save(_ context.Context, p *entity.UserProfile) error {
	cp := *p
	r.store[p.ID] = &cp
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

var testNow = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

func newHandler(repo *stubRepo) *Handler {
	get := app.NewGetProfileUseCase(repo)
	update := app.NewUpdateProfileUseCase(repo)
	return NewHandler(get, update)
}

func seedProfile(repo *stubRepo, id, fullName, phone string) {
	p, _ := entity.NewUserProfile(id, fullName, phone, "", "", time.Time{}, entity.GenderUnspecified, testNow)
	repo.store[id] = p
}

func grpcCode(err error) codes.Code {
	s, _ := status.FromError(err)
	return s.Code()
}

// ─── GetProfile ──────────────────────────────────────────────────────────────

func TestHandler_GetProfile_Success(t *testing.T) {
	repo := newStubRepo()
	seedProfile(repo, "u-1", "Nguyen Van A", "+84901234567")
	h := newHandler(repo)

	resp, err := h.GetProfile(context.Background(), &userpb.GetProfileRequest{UserId: "u-1"})
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if resp.Profile.Id != "u-1" {
		t.Errorf("ID: got %q", resp.Profile.Id)
	}
	if resp.Profile.FullName != "Nguyen Van A" {
		t.Errorf("FullName: got %q", resp.Profile.FullName)
	}
	if resp.Profile.Phone != "+84901234567" {
		t.Errorf("Phone: got %q", resp.Profile.Phone)
	}
	if resp.Profile.Status != string(entity.ProfileStatusActive) {
		t.Errorf("Status: got %q", resp.Profile.Status)
	}
}

func TestHandler_GetProfile_EmptyUserID(t *testing.T) {
	h := newHandler(newStubRepo())
	_, err := h.GetProfile(context.Background(), &userpb.GetProfileRequest{UserId: ""})
	if err == nil {
		t.Fatal("expected error for empty user_id")
	}
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", grpcCode(err))
	}
}

func TestHandler_GetProfile_NotFound(t *testing.T) {
	h := newHandler(newStubRepo())
	_, err := h.GetProfile(context.Background(), &userpb.GetProfileRequest{UserId: "ghost"})
	if err == nil {
		t.Fatal("expected NotFound error")
	}
	if grpcCode(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %v", grpcCode(err))
	}
}

func TestHandler_GetProfile_ProtoFields(t *testing.T) {
	repo := newStubRepo()
	seedProfile(repo, "u-1", "Ahmad", "+601112345678")
	h := newHandler(repo)

	resp, err := h.GetProfile(context.Background(), &userpb.GetProfileRequest{UserId: "u-1"})
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	proto := resp.Profile
	if proto.CreatedAt == nil {
		t.Error("created_at must not be nil")
	}
	if proto.UpdatedAt == nil {
		t.Error("updated_at must not be nil")
	}
	if proto.DateOfBirth != nil {
		t.Error("date_of_birth should be nil when not set")
	}
}

// ─── UpdateProfile ───────────────────────────────────────────────────────────

func TestHandler_UpdateProfile_Success(t *testing.T) {
	repo := newStubRepo()
	seedProfile(repo, "u-1", "Old Name", "+84901234567")
	h := newHandler(repo)

	resp, err := h.UpdateProfile(context.Background(), &userpb.UpdateProfileRequest{
		UserId:   "u-1",
		FullName: "New Name",
		Email:    "new@example.com",
		Gender:   string(entity.GenderFemale),
	})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if resp.Profile.FullName != "New Name" {
		t.Errorf("FullName: got %q", resp.Profile.FullName)
	}
	if resp.Profile.Email != "new@example.com" {
		t.Errorf("Email: got %q", resp.Profile.Email)
	}
	if resp.Profile.Gender != string(entity.GenderFemale) {
		t.Errorf("Gender: got %q", resp.Profile.Gender)
	}
}

func TestHandler_UpdateProfile_EmptyUserID(t *testing.T) {
	h := newHandler(newStubRepo())
	_, err := h.UpdateProfile(context.Background(), &userpb.UpdateProfileRequest{
		UserId: "", FullName: "Name",
	})
	if err == nil {
		t.Fatal("expected error for empty user_id")
	}
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", grpcCode(err))
	}
}

func TestHandler_UpdateProfile_NotFound(t *testing.T) {
	h := newHandler(newStubRepo())
	_, err := h.UpdateProfile(context.Background(), &userpb.UpdateProfileRequest{
		UserId: "ghost", FullName: "Name", Gender: string(entity.GenderUnspecified),
	})
	if err == nil {
		t.Fatal("expected NotFound error")
	}
	if grpcCode(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %v", grpcCode(err))
	}
}

func TestHandler_UpdateProfile_ValidationError(t *testing.T) {
	repo := newStubRepo()
	seedProfile(repo, "u-1", "Name", "+1")
	h := newHandler(repo)

	_, err := h.UpdateProfile(context.Background(), &userpb.UpdateProfileRequest{
		UserId:   "u-1",
		FullName: "", // invalid
		Gender:   string(entity.GenderUnspecified),
	})
	if err == nil {
		t.Fatal("expected InvalidArgument, got nil")
	}
	if grpcCode(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", grpcCode(err))
	}
}

func TestHandler_UpdateProfile_PhoneNotUpdatable(t *testing.T) {
	repo := newStubRepo()
	seedProfile(repo, "u-1", "Name", "+84901234567")
	h := newHandler(repo)

	resp, err := h.UpdateProfile(context.Background(), &userpb.UpdateProfileRequest{
		UserId:   "u-1",
		FullName: "New Name",
		Gender:   string(entity.GenderUnspecified),
	})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	// Phone must remain unchanged.
	if resp.Profile.Phone != "+84901234567" {
		t.Errorf("Phone changed: got %q", resp.Profile.Phone)
	}
}

// ─── toGRPCError mapping ─────────────────────────────────────────────────────

func TestToGRPCError_Mapping(t *testing.T) {
	cases := []struct {
		domainCode domainerrors.Code
		grpcCode   codes.Code
	}{
		{domainerrors.CodeNotFound, codes.NotFound},
		{domainerrors.CodeInvalidArgument, codes.InvalidArgument},
		{domainerrors.CodeAlreadyExists, codes.AlreadyExists},
		{domainerrors.CodeUnauthenticated, codes.Unauthenticated},
		{domainerrors.CodePermissionDenied, codes.PermissionDenied},
		{domainerrors.CodePreconditionFailed, codes.FailedPrecondition},
		{domainerrors.CodeUnavailable, codes.Unavailable},
		{domainerrors.CodeInternalError, codes.Internal},
	}
	for _, tc := range cases {
		err := domainerrors.New(tc.domainCode, "test")
		got := grpcCode(toGRPCError(err))
		if got != tc.grpcCode {
			t.Errorf("domain %v → got %v, want %v", tc.domainCode, got, tc.grpcCode)
		}
	}
}

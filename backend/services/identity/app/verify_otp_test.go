package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/fairride/identity/app"
	"github.com/fairride/identity/domain/entity"
	"github.com/fairride/shared/errors"
)

var appTestNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// ─── in-memory fakes ────────────────────────────────────────────────────────

type fakeOTPRepo struct {
	byPhone map[string]*entity.OTPChallenge
}

func newFakeOTPRepo() *fakeOTPRepo { return &fakeOTPRepo{byPhone: map[string]*entity.OTPChallenge{}} }

func (r *fakeOTPRepo) Save(_ context.Context, c *entity.OTPChallenge) error {
	cp := *c
	r.byPhone[c.PhoneNumber] = &cp
	return nil
}

func (r *fakeOTPRepo) FindLatestByPhone(_ context.Context, phone string) (*entity.OTPChallenge, error) {
	c, ok := r.byPhone[phone]
	if !ok {
		return nil, errors.NotFound("no otp challenge found for phone")
	}
	cp := *c
	return &cp, nil
}

type fakeUserRepo struct {
	byID        map[string]*entity.User
	byPhone     map[string]*entity.User
	byEmail     map[string]*entity.User
	byGoogleSub map[string]*entity.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		byID:        map[string]*entity.User{},
		byPhone:     map[string]*entity.User{},
		byEmail:     map[string]*entity.User{},
		byGoogleSub: map[string]*entity.User{},
	}
}

func (r *fakeUserRepo) FindByID(_ context.Context, id string) (*entity.User, error) {
	if u, ok := r.byID[id]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, errors.NotFound("user not found")
}

func (r *fakeUserRepo) FindByPhone(_ context.Context, phone string) (*entity.User, error) {
	if u, ok := r.byPhone[phone]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, errors.NotFound("user not found")
}

func (r *fakeUserRepo) FindByEmail(_ context.Context, email string) (*entity.User, error) {
	if u, ok := r.byEmail[email]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, errors.NotFound("user not found")
}

func (r *fakeUserRepo) FindByGoogleSub(_ context.Context, sub string) (*entity.User, error) {
	if u, ok := r.byGoogleSub[sub]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, errors.NotFound("user not found")
}

func (r *fakeUserRepo) FindAll(_ context.Context) ([]*entity.User, error) {
	var out []*entity.User
	for _, u := range r.byID {
		out = append(out, u)
	}
	return out, nil
}

func (r *fakeUserRepo) Save(_ context.Context, u *entity.User) error {
	cp := *u
	r.byID[u.ID] = &cp
	if u.PhoneNumber != "" {
		r.byPhone[u.PhoneNumber] = &cp
	}
	if u.Email != "" {
		r.byEmail[u.Email] = &cp
	}
	if u.GoogleSub != "" {
		r.byGoogleSub[u.GoogleSub] = &cp
	}
	return nil
}

func (r *fakeUserRepo) Delete(_ context.Context, id string) error {
	delete(r.byID, id)
	return nil
}

type fakeRoleRepo struct {
	byName map[string]*entity.Role
}

func newFakeRoleRepo() *fakeRoleRepo {
	repo := &fakeRoleRepo{byName: map[string]*entity.Role{}}
	riderRole, _ := entity.NewRole("role-rider", entity.RoleRider, "", true, appTestNow)
	driverRole, _ := entity.NewRole("role-driver", entity.RoleDriver, "", true, appTestNow)
	repo.byName[entity.RoleRider] = riderRole
	repo.byName[entity.RoleDriver] = driverRole
	return repo
}

func (r *fakeRoleRepo) FindByID(_ context.Context, id string) (*entity.Role, error) {
	for _, role := range r.byName {
		if role.ID == id {
			return role, nil
		}
	}
	return nil, errors.NotFound("role not found")
}

func (r *fakeRoleRepo) FindByName(_ context.Context, name string) (*entity.Role, error) {
	if role, ok := r.byName[name]; ok {
		return role, nil
	}
	return nil, errors.NotFound("role not found")
}

func (r *fakeRoleRepo) FindAll(_ context.Context) ([]*entity.Role, error) {
	var out []*entity.Role
	for _, role := range r.byName {
		out = append(out, role)
	}
	return out, nil
}

func (r *fakeRoleRepo) Save(_ context.Context, role *entity.Role) error {
	r.byName[role.Name] = role
	return nil
}

func (r *fakeRoleRepo) Delete(_ context.Context, id string) error { return nil }

// ─── tests ──────────────────────────────────────────────────────────────────

func TestVerifyOTP_WrongCodeRejected(t *testing.T) {
	otpRepo := newFakeOTPRepo()
	userRepo := newFakeUserRepo()
	roleRepo := newFakeRoleRepo()

	challenge, _ := entity.NewOTPChallenge("otp-1", "+84901111111", "123456", "login", time.Now())
	_ = otpRepo.Save(context.Background(), challenge)

	findOrCreate := app.NewFindOrCreateUserUseCase(userRepo, roleRepo)
	uc := app.NewVerifyOTPUseCase(otpRepo, findOrCreate)

	_, err := uc.Execute(context.Background(), app.VerifyOTPInput{
		PhoneNumber: "+84901111111",
		Code:        "000000",
		UserType:    entity.TypeRider,
	})
	if err == nil {
		t.Fatal("expected error for wrong code")
	}
	if len(userRepo.byID) != 0 {
		t.Fatal("no user should have been created on a failed verify")
	}
}

func TestVerifyOTP_CorrectCode_CreatesNewActiveUser(t *testing.T) {
	otpRepo := newFakeOTPRepo()
	userRepo := newFakeUserRepo()
	roleRepo := newFakeRoleRepo()

	// Verify() checks expiry against real wall-clock time (VerifyOTPUseCase
	// has no injectable clock), so the challenge must be created "now" —
	// appTestNow is a fixed historical date and would already read as expired.
	challenge, _ := entity.NewOTPChallenge("otp-1", "+84901111111", "123456", "login", time.Now())
	_ = otpRepo.Save(context.Background(), challenge)

	findOrCreate := app.NewFindOrCreateUserUseCase(userRepo, roleRepo)
	uc := app.NewVerifyOTPUseCase(otpRepo, findOrCreate)

	result, err := uc.Execute(context.Background(), app.VerifyOTPInput{
		PhoneNumber: "+84901111111",
		Code:        "123456",
		UserType:    entity.TypeRider,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNewUser {
		t.Fatal("expected IsNewUser=true for a first-time login")
	}
	if result.User.Status != entity.StatusActive {
		t.Fatalf("expected new user to be Active (OTP verified), got %s", result.User.Status)
	}
	if result.User.Type != entity.TypeRider {
		t.Fatalf("expected user type rider, got %s", result.User.Type)
	}
}

func TestVerifyOTP_SamePhone_BothApps_OneUnifiedAccount(t *testing.T) {
	otpRepo := newFakeOTPRepo()
	userRepo := newFakeUserRepo()
	roleRepo := newFakeRoleRepo()
	findOrCreate := app.NewFindOrCreateUserUseCase(userRepo, roleRepo)
	uc := app.NewVerifyOTPUseCase(otpRepo, findOrCreate)

	// First login from the Rider app creates the account.
	c1, _ := entity.NewOTPChallenge("otp-1", "+84901111111", "123456", "login", time.Now())
	_ = otpRepo.Save(context.Background(), c1)
	riderResult, err := uc.Execute(context.Background(), app.VerifyOTPInput{
		PhoneNumber: "+84901111111",
		Code:        "123456",
		UserType:    entity.TypeRider,
	})
	if err != nil {
		t.Fatalf("rider login: unexpected error: %v", err)
	}
	if riderResult.User.Type != entity.TypeRider {
		t.Fatalf("expected Type=rider, got %s", riderResult.User.Type)
	}
	if riderResult.User.DriverEnabled {
		t.Fatal("expected DriverEnabled=false right after a rider-app-only login")
	}

	// Same phone logging into the Driver app must NOT be rejected — it's
	// the same account, now also driver-enabled, still Type=rider.
	c2, _ := entity.NewOTPChallenge("otp-2", "+84901111111", "654321", "login", time.Now())
	_ = otpRepo.Save(context.Background(), c2)
	driverResult, err := uc.Execute(context.Background(), app.VerifyOTPInput{
		PhoneNumber: "+84901111111",
		Code:        "654321",
		UserType:    entity.TypeDriver,
	})
	if err != nil {
		t.Fatalf("driver login: unexpected error: %v", err)
	}
	if driverResult.IsNewUser {
		t.Fatal("expected the SAME account to be reused, not a new one, for the driver-app login")
	}
	if driverResult.User.ID != riderResult.User.ID {
		t.Fatalf("expected same user ID across apps, got %s vs %s", driverResult.User.ID, riderResult.User.ID)
	}
	if driverResult.User.Type != entity.TypeRider {
		t.Fatalf("expected Type to remain rider after driver-app login, got %s", driverResult.User.Type)
	}
	if !driverResult.User.DriverEnabled {
		t.Fatal("expected DriverEnabled=true after logging into the Driver app")
	}
}

func TestRequestOTP_CooldownBlocksImmediateResend(t *testing.T) {
	otpRepo := newFakeOTPRepo()
	provider := &countingProvider{}
	uc := app.NewRequestOTPUseCase(otpRepo, provider)

	if _, err := uc.Execute(context.Background(), "+84902222222"); err != nil {
		t.Fatalf("unexpected error on first request: %v", err)
	}
	if provider.calls != 1 {
		t.Fatalf("expected provider to be called once, got %d", provider.calls)
	}
	_, err := uc.Execute(context.Background(), "+84902222222")
	if err == nil {
		t.Fatal("expected cooldown error on immediate resend")
	}
	if errors.GetCode(err) != errors.CodeResourceExhausted {
		t.Fatalf("expected CodeResourceExhausted, got %v", errors.GetCode(err))
	}
}

type countingProvider struct{ calls int }

func (p *countingProvider) Send(_ context.Context, _, _ string) error {
	p.calls++
	return nil
}

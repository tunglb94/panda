package app

import (
	"context"
	"strings"
	"time"

	"github.com/fairride/identity/domain/entity"
	"github.com/fairride/identity/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// VerifyOTPInput carries the code the client submitted and which app is
// logging in (the Rider app always sends "rider", the Driver app always
// sends "driver"). UserType here means "app context", not account Type —
// every account is TypeRider regardless; see FindOrCreateUserUseCase.
type VerifyOTPInput struct {
	PhoneNumber string
	Code        string
	UserType    entity.UserType
}

// VerifyOTPResult reports the resolved user and whether this was their
// first-ever login (auto-provisioned just now).
type VerifyOTPResult struct {
	User      *entity.User
	IsNewUser bool
}

// VerifyOTPUseCase checks a submitted code against the latest challenge for
// that phone number, then finds-or-creates the account — this is the entire
// "no office visit" signup path: a correct OTP on a new phone number IS the
// signup.
type VerifyOTPUseCase struct {
	otpRepo      repository.OTPRepository
	findOrCreate *FindOrCreateUserUseCase
}

func NewVerifyOTPUseCase(otpRepo repository.OTPRepository, findOrCreate *FindOrCreateUserUseCase) *VerifyOTPUseCase {
	return &VerifyOTPUseCase{otpRepo: otpRepo, findOrCreate: findOrCreate}
}

func (uc *VerifyOTPUseCase) Execute(ctx context.Context, in VerifyOTPInput) (*VerifyOTPResult, error) {
	phoneNumber := strings.TrimSpace(in.PhoneNumber)
	code := strings.TrimSpace(in.Code)
	if phoneNumber == "" || code == "" {
		return nil, domainerrors.InvalidArgument("phone and code are required")
	}

	challenge, err := uc.otpRepo.FindLatestByPhone(ctx, phoneNumber)
	if err != nil {
		if domainerrors.GetCode(err) == domainerrors.CodeNotFound {
			return nil, domainerrors.Unauthenticated("no otp was requested for this phone")
		}
		return nil, err
	}

	verifyErr := challenge.Verify(code, time.Now())
	// Verify mutates challenge.Attempts/Consumed in place even on failure
	// (an incorrect guess still counts toward the attempt limit) — persist
	// either way before returning.
	if saveErr := uc.otpRepo.Save(ctx, challenge); saveErr != nil {
		return nil, saveErr
	}
	if verifyErr != nil {
		return nil, verifyErr
	}

	user, isNew, err := uc.findOrCreate.ByPhone(ctx, ByPhoneInput{
		PhoneNumber: phoneNumber,
		AppContext:  in.UserType,
	})
	if err != nil {
		return nil, err
	}
	return &VerifyOTPResult{User: user, IsNewUser: isNew}, nil
}

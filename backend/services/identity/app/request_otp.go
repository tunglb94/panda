package app

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/fairride/identity/domain/entity"
	"github.com/fairride/identity/domain/repository"
	"github.com/fairride/identity/infrastructure/otp"
	domainerrors "github.com/fairride/shared/errors"
)

// RequestOTPResult is what the handler needs to build its response —
// including the plaintext Code, which the handler alone decides whether to
// echo back (only ever when APP_ENV=development; see gateway auth_handler.go).
type RequestOTPResult struct {
	Code      string
	ExpiresIn time.Duration
}

// RequestOTPUseCase generates and delivers a fresh OTP code for a phone
// number, enforcing a resend cooldown so one caller can't hammer the
// downstream SMS/Zalo provider.
type RequestOTPUseCase struct {
	otpRepo  repository.OTPRepository
	provider otp.Provider
}

func NewRequestOTPUseCase(otpRepo repository.OTPRepository, provider otp.Provider) *RequestOTPUseCase {
	return &RequestOTPUseCase{otpRepo: otpRepo, provider: provider}
}

func (uc *RequestOTPUseCase) Execute(ctx context.Context, phoneNumber string) (*RequestOTPResult, error) {
	phoneNumber = strings.TrimSpace(phoneNumber)
	if phoneNumber == "" {
		return nil, domainerrors.InvalidArgument("phone is required")
	}

	now := time.Now()
	if latest, err := uc.otpRepo.FindLatestByPhone(ctx, phoneNumber); err == nil {
		if remaining := latest.CooldownRemaining(now); remaining > 0 {
			return nil, domainerrors.ResourceExhausted(
				fmt.Sprintf("please wait %d seconds before requesting another code", int(remaining.Seconds())+1))
		}
	} else if domainerrors.GetCode(err) != domainerrors.CodeNotFound {
		return nil, err
	}

	code, err := generateOTPCode()
	if err != nil {
		return nil, domainerrors.Internal("generate otp code").WithMeta("error", err.Error())
	}
	id, err := newID()
	if err != nil {
		return nil, domainerrors.Internal("generate otp challenge id").WithMeta("error", err.Error())
	}
	challenge, err := entity.NewOTPChallenge(id, phoneNumber, code, "login", now)
	if err != nil {
		return nil, err
	}
	if err := uc.otpRepo.Save(ctx, challenge); err != nil {
		return nil, err
	}
	if err := uc.provider.Send(ctx, phoneNumber, code); err != nil {
		return nil, err
	}

	return &RequestOTPResult{Code: code, ExpiresIn: entity.OTPTTL}, nil
}

// generateOTPCode returns a cryptographically random 6-digit numeric string
// (zero-padded — "000042" is a valid code, not "42").
func generateOTPCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

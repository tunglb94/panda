package app

import (
	"context"
	"strings"

	"github.com/fairride/identity/domain/entity"
	"github.com/fairride/identity/infrastructure/googleauth"
	domainerrors "github.com/fairride/shared/errors"
)

// GoogleLoginInput carries the ID token the client obtained from
// google_sign_in, and which app is logging in (app context — see
// VerifyOTPInput's doc comment; every account is TypeRider regardless).
type GoogleLoginInput struct {
	IDToken  string
	UserType entity.UserType
}

// GoogleLoginResult mirrors VerifyOTPResult.
type GoogleLoginResult struct {
	User      *entity.User
	IsNewUser bool
}

// GoogleLoginUseCase verifies a Google ID token server-side, then
// finds-or-creates the account keyed by google_sub (falling back to email —
// see FindOrCreateUserUseCase.ByGoogle). Google accounts have no phone
// number — PhoneNumber stays "" until a future phone-link flow.
type GoogleLoginUseCase struct {
	verifier     googleauth.Verifier
	findOrCreate *FindOrCreateUserUseCase
}

func NewGoogleLoginUseCase(verifier googleauth.Verifier, findOrCreate *FindOrCreateUserUseCase) *GoogleLoginUseCase {
	return &GoogleLoginUseCase{verifier: verifier, findOrCreate: findOrCreate}
}

func (uc *GoogleLoginUseCase) Execute(ctx context.Context, in GoogleLoginInput) (*GoogleLoginResult, error) {
	idToken := strings.TrimSpace(in.IDToken)
	if idToken == "" {
		return nil, domainerrors.InvalidArgument("id_token is required")
	}

	identity, err := uc.verifier.Verify(ctx, idToken)
	if err != nil {
		return nil, err
	}
	if identity.Email == "" || !identity.EmailVerified {
		return nil, domainerrors.PermissionDenied("google account email is not verified")
	}

	user, isNew, err := uc.findOrCreate.ByGoogle(ctx, ByGoogleInput{
		GoogleSub:  identity.Sub,
		Email:      identity.Email,
		Name:       identity.Name,
		AppContext: in.UserType,
	})
	if err != nil {
		return nil, err
	}
	return &GoogleLoginResult{User: user, IsNewUser: isNew}, nil
}

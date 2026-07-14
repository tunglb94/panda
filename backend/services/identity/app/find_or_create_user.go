package app

import (
	"context"
	"time"

	"github.com/fairride/identity/domain/entity"
	"github.com/fairride/identity/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// FindOrCreateUserUseCase is the shared "no office visit" account
// resolution used by both OTP login and Google login: look the account up
// by its natural key (phone for OTP, google_sub then email for Google), and
// if none exists, create one on the spot. Every account created this way is
// immediately Activated — the OTP/Google verification the caller already
// did IS the phone/email verification entity.User's lifecycle comment
// describes ("PendingVerification → Active (phone OTP verified)").
//
// Every new account is TypeRider/RoleRider regardless of which app first
// created it — a login from the Driver app additionally flips
// DriverEnabled (see ensureAppContext) instead of changing Type/RoleID, so
// one account can be both a Rider and a Driver (see entity.User's doc
// comment for the full rationale).
type FindOrCreateUserUseCase struct {
	users repository.UserRepository
	roles repository.RoleRepository
}

func NewFindOrCreateUserUseCase(users repository.UserRepository, roles repository.RoleRepository) *FindOrCreateUserUseCase {
	return &FindOrCreateUserUseCase{users: users, roles: roles}
}

// ByPhoneInput resolves/creates an account by phone number (OTP login path).
type ByPhoneInput struct {
	PhoneNumber string
	Name        string          // used only when creating a new account
	AppContext  entity.UserType // which app is logging in — "driver" flips DriverEnabled; never changes Type/RoleID
}

// ByPhone finds the user with PhoneNumber, or creates a new Rider account if
// none exists.
func (uc *FindOrCreateUserUseCase) ByPhone(ctx context.Context, in ByPhoneInput) (*entity.User, bool, error) {
	existing, err := uc.users.FindByPhone(ctx, in.PhoneNumber)
	if err == nil {
		user, ensureErr := uc.ensureAppContext(ctx, existing, in.AppContext)
		return user, false, ensureErr
	}
	if domainerrors.GetCode(err) != domainerrors.CodeNotFound {
		return nil, false, err
	}

	roleID, err := uc.riderRoleID(ctx)
	if err != nil {
		return nil, false, err
	}
	name := in.Name
	if name == "" {
		name = in.PhoneNumber
	}
	id, err := newID()
	if err != nil {
		return nil, false, domainerrors.Internal("generate user id").WithMeta("error", err.Error())
	}
	now := time.Now()
	user, err := entity.NewUser(id, in.PhoneNumber, name, "", entity.TypeRider, roleID, now)
	if err != nil {
		return nil, false, err
	}
	if err := user.Activate(now); err != nil {
		return nil, false, err
	}
	if in.AppContext == entity.TypeDriver {
		user.EnableDriverCapability(now)
	}
	if err := uc.users.Save(ctx, user); err != nil {
		return nil, false, err
	}
	return user, true, nil
}

// ByGoogleInput resolves/creates an account by Google identity.
type ByGoogleInput struct {
	GoogleSub  string
	Email      string
	Name       string
	AppContext entity.UserType
}

// ByGoogle tries GoogleSub first (the stable identifier), then Email (a
// phone/OTP-created account whose email happens to match gets GoogleSub
// auto-linked — this is a same-email auto-link, not a full account merge;
// merging two genuinely separate existing accounts is deferred, see the
// plan's Known Gaps). Creates a new Rider account if neither matches.
func (uc *FindOrCreateUserUseCase) ByGoogle(ctx context.Context, in ByGoogleInput) (*entity.User, bool, error) {
	if existing, err := uc.users.FindByGoogleSub(ctx, in.GoogleSub); err == nil {
		user, ensureErr := uc.ensureAppContext(ctx, existing, in.AppContext)
		return user, false, ensureErr
	} else if domainerrors.GetCode(err) != domainerrors.CodeNotFound {
		return nil, false, err
	}

	if in.Email != "" {
		if existing, err := uc.users.FindByEmail(ctx, in.Email); err == nil {
			now := time.Now()
			if err := existing.LinkGoogleSub(in.GoogleSub, now); err != nil {
				return nil, false, err
			}
			if in.AppContext == entity.TypeDriver {
				existing.EnableDriverCapability(now)
			}
			if err := uc.users.Save(ctx, existing); err != nil {
				return nil, false, err
			}
			return existing, false, checkStatus(existing)
		} else if domainerrors.GetCode(err) != domainerrors.CodeNotFound {
			return nil, false, err
		}
	}

	roleID, err := uc.riderRoleID(ctx)
	if err != nil {
		return nil, false, err
	}
	name := in.Name
	if name == "" {
		name = in.Email
	}
	id, err := newID()
	if err != nil {
		return nil, false, domainerrors.Internal("generate user id").WithMeta("error", err.Error())
	}
	now := time.Now()
	user, err := entity.NewGoogleUser(id, in.Email, in.GoogleSub, name, roleID, now)
	if err != nil {
		return nil, false, err
	}
	if err := user.Activate(now); err != nil {
		return nil, false, err
	}
	if in.AppContext == entity.TypeDriver {
		user.EnableDriverCapability(now)
	}
	if err := uc.users.Save(ctx, user); err != nil {
		return nil, false, err
	}
	return user, true, nil
}

// ensureAppContext checks the resolved account's status and, when logging
// in from the Driver app, flips DriverEnabled on (idempotent) — the "no
// office visit" auto-provisioning that used to require a matching Type now
// applies to the capability flag instead.
func (uc *FindOrCreateUserUseCase) ensureAppContext(ctx context.Context, user *entity.User, appContext entity.UserType) (*entity.User, error) {
	if err := checkStatus(user); err != nil {
		return user, err
	}
	if appContext == entity.TypeDriver && !user.DriverEnabled {
		user.EnableDriverCapability(time.Now())
		if err := uc.users.Save(ctx, user); err != nil {
			return user, err
		}
	}
	return user, nil
}

func (uc *FindOrCreateUserUseCase) riderRoleID(ctx context.Context) (string, error) {
	role, err := uc.roles.FindByName(ctx, entity.RoleRider)
	if err != nil {
		return "", err
	}
	return role.ID, nil
}

func checkStatus(user *entity.User) error {
	if user.Status == entity.StatusDeactivated {
		return domainerrors.PermissionDenied("this account has been deactivated")
	}
	if user.Status == entity.StatusSuspended {
		return domainerrors.PermissionDenied("this account is suspended")
	}
	return nil
}

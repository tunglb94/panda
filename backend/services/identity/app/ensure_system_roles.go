package app

import (
	"context"
	"time"

	"github.com/fairride/identity/domain/entity"
	"github.com/fairride/identity/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// systemRoleIDs gives every system role a deterministic ID (rather than a
// random UUID) so EnsureSystemRolesUseCase is idempotent across restarts —
// running it twice must not create duplicate roles.
var systemRoleIDs = map[string]string{
	entity.RoleRider:  "role-rider",
	entity.RoleDriver: "role-driver",
}

// EnsureSystemRolesUseCase seeds the "rider"/"driver" system roles if they
// don't already exist. entity.Role's own doc comment says system roles "are
// seeded at service startup" — nothing implemented that until now, and
// FindOrCreateUserUseCase needs a RoleID to assign to every new account, so
// this runs once at gateway startup before the first login can happen.
type EnsureSystemRolesUseCase struct {
	roles repository.RoleRepository
}

func NewEnsureSystemRolesUseCase(roles repository.RoleRepository) *EnsureSystemRolesUseCase {
	return &EnsureSystemRolesUseCase{roles: roles}
}

func (uc *EnsureSystemRolesUseCase) Execute(ctx context.Context) error {
	now := time.Now()
	for name, id := range systemRoleIDs {
		_, err := uc.roles.FindByID(ctx, id)
		if err == nil {
			continue
		}
		if domainerrors.GetCode(err) != domainerrors.CodeNotFound {
			return err
		}
		role, err := entity.NewRole(id, name, "system role — seeded at startup", true, now)
		if err != nil {
			return err
		}
		if err := uc.roles.Save(ctx, role); err != nil {
			return err
		}
	}
	return nil
}

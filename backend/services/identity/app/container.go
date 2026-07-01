// Package app is the application layer of the Identity service.
// It owns the dependency graph: use cases are wired to their repository
// implementations here, and the Container is the single composition root.
package app

import (
	"github.com/fairride/identity/domain/repository"
)

// Container is the composition root for the Identity service.
// All repository implementations are injected at startup via New.
// Use case types declared in future phases will be added as fields here.
type Container struct {
	Roles       repository.RoleRepository
	Permissions repository.PermissionRepository
}

// New constructs a Container. Both repositories are required.
// Panics on nil input — a missing repository is a programmer error, not a runtime error.
func New(roles repository.RoleRepository, permissions repository.PermissionRepository) *Container {
	if roles == nil {
		panic("identity: RoleRepository must not be nil")
	}
	if permissions == nil {
		panic("identity: PermissionRepository must not be nil")
	}
	return &Container{
		Roles:       roles,
		Permissions: permissions,
	}
}

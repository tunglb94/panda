package entity

import (
	"time"

	"github.com/fairride/shared/errors"
)

// System role names. These roles exist on every FAIRRIDE deployment and cannot be deleted.
// They are seeded at service startup and map directly to the user types in DOC-0002 §6.12.
const (
	RoleRider         = "rider"
	RoleDriver        = "driver"
	RoleFleetOperator = "fleet_operator"
	RoleCityManager   = "city_manager"
	RoleSupportAgent  = "support_agent"
	RoleSuperAdmin    = "super_admin"
)

// Role is a named collection of Permissions. Roles are assigned to users at
// the application layer; the domain models what a Role is and what it permits.
type Role struct {
	ID          string
	Name        string
	Description string
	// IsSystem marks roles that were seeded at startup and cannot be deleted.
	IsSystem    bool
	permissions map[string]Permission // keyed by Permission.ID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewRole creates a new Role with no permissions.
// id and name must be non-empty.
func NewRole(id, name, description string, isSystem bool, now time.Time) (*Role, error) {
	if id == "" {
		return nil, errors.InvalidArgument("role id must not be empty")
	}
	if name == "" {
		return nil, errors.InvalidArgument("role name must not be empty")
	}
	return &Role{
		ID:          id,
		Name:        name,
		Description: description,
		IsSystem:    isSystem,
		permissions: make(map[string]Permission),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// ReconstituteRole rebuilds a Role from a persistence record.
// No validation is applied — data is assumed already valid.
func ReconstituteRole(
	id, name, description string,
	isSystem bool,
	perms []Permission,
	createdAt, updatedAt time.Time,
) *Role {
	r := &Role{
		ID:          id,
		Name:        name,
		Description: description,
		IsSystem:    isSystem,
		permissions: make(map[string]Permission, len(perms)),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
	for _, p := range perms {
		r.permissions[p.ID] = p
	}
	return r
}

// AddPermission idempotently grants a Permission to this Role.
func (r *Role) AddPermission(p Permission) {
	r.permissions[p.ID] = p
}

// RemovePermission revokes the permission with the given ID. No-op if not present.
func (r *Role) RemovePermission(permissionID string) {
	delete(r.permissions, permissionID)
}

// HasPermission reports whether this Role grants the permission identified by name
// (e.g. "trips:read"). Returns false if no permission with that name is present.
func (r *Role) HasPermission(name string) bool {
	for _, p := range r.permissions {
		if p.Name == name {
			return true
		}
	}
	return false
}

// Permissions returns a snapshot of all permissions currently granted to this Role.
// Mutations to the returned slice do not affect the Role.
func (r *Role) Permissions() []Permission {
	out := make([]Permission, 0, len(r.permissions))
	for _, p := range r.permissions {
		out = append(out, p)
	}
	return out
}

// PermissionCount returns how many permissions are currently granted.
func (r *Role) PermissionCount() int {
	return len(r.permissions)
}

// CanDelete returns an error if this Role must not be deleted.
// System roles cannot be deleted — they are platform invariants.
func (r *Role) CanDelete() error {
	if r.IsSystem {
		return errors.PreconditionFailed("system role '" + r.Name + "' cannot be deleted")
	}
	return nil
}

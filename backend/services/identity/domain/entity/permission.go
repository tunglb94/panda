package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// Permission is a fine-grained capability within the FAIRRIDE platform.
// Names follow the "resource:action" convention, e.g. "trips:read", "drivers:manage".
type Permission struct {
	ID          string
	Name        string
	Resource    string
	Action      string
	Description string
	CreatedAt   time.Time
}

// Well-known permission names. Each follows the "resource:action" pattern.
// These constants are the canonical source of truth for permission names across all services.
const (
	PermTripsRead    = "trips:read"
	PermTripsWrite   = "trips:write"
	PermTripsManage  = "trips:manage"

	PermDriversRead   = "drivers:read"
	PermDriversWrite  = "drivers:write"
	PermDriversManage = "drivers:manage"

	PermRidersRead   = "riders:read"
	PermRidersWrite  = "riders:write"
	PermRidersManage = "riders:manage"

	PermWalletRead  = "wallet:read"
	PermWalletWrite = "wallet:write"

	PermPaymentsRead  = "payments:read"
	PermPaymentsWrite = "payments:write"

	PermDispatchRead  = "dispatch:read"
	PermDispatchWrite = "dispatch:write"

	PermReviewsRead  = "reviews:read"
	PermReviewsWrite = "reviews:write"

	PermReportsRead = "reports:read"

	PermSupportRead  = "support:read"
	PermSupportWrite = "support:write"

	PermAdminRead   = "admin:read"
	PermAdminWrite  = "admin:write"
	PermAdminManage = "admin:manage"
)

// NewPermission constructs a Permission with a validated name.
// id must be non-empty. name must follow the "resource:action" pattern.
func NewPermission(id, name, description string, createdAt time.Time) (*Permission, error) {
	if id == "" {
		return nil, errors.InvalidArgument("permission id must not be empty")
	}
	resource, action, err := parsePermissionName(name)
	if err != nil {
		return nil, err
	}
	return &Permission{
		ID:          id,
		Name:        name,
		Resource:    resource,
		Action:      action,
		Description: description,
		CreatedAt:   createdAt,
	}, nil
}

// ReconstitutePermission rebuilds a Permission from a persistence record.
// No validation is applied — data is assumed already valid.
func ReconstitutePermission(id, name, resource, action, description string, createdAt time.Time) *Permission {
	return &Permission{
		ID:          id,
		Name:        name,
		Resource:    resource,
		Action:      action,
		Description: description,
		CreatedAt:   createdAt,
	}
}

// parsePermissionName splits "resource:action" and returns an error if malformed.
func parsePermissionName(name string) (resource, action string, err error) {
	parts := strings.SplitN(name, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.InvalidArgument(
			"permission name must follow 'resource:action' format, got: " + name,
		)
	}
	return parts[0], parts[1], nil
}

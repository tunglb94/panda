package entity

import "github.com/fairride/shared/errors"

// LicenseClass is a Vietnamese driving-license class ("hạng bằng lái").
// Only the classes relevant to Panda's ride tiers are modeled — Panda
// doesn't offer any service tier that would need e.g. class C/D/E (trucks).
//
// Which ServiceType each class legally permits is deliberately NOT encoded
// here (Phần 1 of the Driver KYC Hardening spec — "Không hardcode luật
// GPLX... Rule Engine quyết định"): that mapping lives in the
// license_capabilities table and is read through
// repository.LicenseCapabilityRepository, so a change in Vietnamese law
// only ever requires a data UPDATE, never a migration or redeploy.
type LicenseClass string

const (
	LicenseClassA1 LicenseClass = "A1"
	LicenseClassA2 LicenseClass = "A2"
	LicenseClassB1 LicenseClass = "B1"
	LicenseClassB2 LicenseClass = "B2"
)

func validateLicenseClass(l LicenseClass) error {
	switch l {
	case LicenseClassA1, LicenseClassA2, LicenseClassB1, LicenseClassB2:
		return nil
	default:
		return errors.InvalidArgument("unknown license class: " + string(l))
	}
}

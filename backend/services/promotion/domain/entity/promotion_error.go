package entity

import (
	stderrors "errors"

	"github.com/fairride/shared/errors"
)

// ReasonCode is a machine-readable classifier for why a voucher failed
// validation, carried in the DomainError's Meta["reason"]. These are exactly
// the 8 checks the sprint brief requires the engine to compute, plus the two
// additional checks required to honor the MinOrder and NotFound fields that
// were also requested but not named among the 8 examples.
type ReasonCode string

const (
	ReasonNotFound         ReasonCode = "VOUCHER_NOT_FOUND"
	ReasonInvalidStatus    ReasonCode = "VOUCHER_INVALID_STATUS"     // "voucher hợp lệ" — not active/draft/paused
	ReasonExpired          ReasonCode = "VOUCHER_EXPIRED"            // "voucher hết hạn"
	ReasonBudgetExhausted  ReasonCode = "VOUCHER_BUDGET_EXHAUSTED"   // "voucher hết ngân sách"
	ReasonUsageExhausted   ReasonCode = "VOUCHER_USAGE_EXHAUSTED"    // "voucher hết lượt"
	ReasonWrongCity        ReasonCode = "VOUCHER_WRONG_CITY"         // "voucher sai thành phố"
	ReasonWrongVehicle     ReasonCode = "VOUCHER_WRONG_VEHICLE"      // "voucher sai loại xe"
	ReasonWrongServiceType ReasonCode = "VOUCHER_WRONG_SERVICE_TYPE" // "voucher sai hạng dịch vụ" (bike/bike_plus/car/car_xl)
	ReasonWrongTripType    ReasonCode = "VOUCHER_WRONG_TRIP_TYPE"    // "voucher sai loại chuyến" (ride/delivery)
	ReasonWrongMembership  ReasonCode = "VOUCHER_WRONG_MEMBERSHIP"   // "voucher sai membership"
	ReasonWrongTiming      ReasonCode = "VOUCHER_WRONG_TIMING"       // "voucher sai thời gian" (before start_time or after end_time)
	ReasonMinOrderNotMet   ReasonCode = "VOUCHER_MIN_ORDER_NOT_MET"
	ReasonNotEligible      ReasonCode = "VOUCHER_NOT_ELIGIBLE"     // PromotionRule-level rejection (e.g. not a first ride, not the rider's birthday)
	ReasonRuleNotDefined   ReasonCode = "VOUCHER_RULE_NOT_DEFINED" // promotion type has no BRB-approved rule yet (TODO)
)

// PromotionError builds a *errors.DomainError for a voucher-validation
// failure, tagging it with the machine-readable ReasonCode so callers (and
// tests) can assert on the specific check that failed without string-matching
// the message.
func PromotionError(reason ReasonCode, message string) *errors.DomainError {
	code := errors.CodeUnprocessable
	if reason == ReasonNotFound {
		code = errors.CodeNotFound
	}
	return errors.New(code, message).WithMeta("reason", string(reason))
}

// ReasonOf extracts the ReasonCode from err (or any wrapped error), or ""
// if err carries none.
func ReasonOf(err error) ReasonCode {
	var de *errors.DomainError
	if stderrors.As(err, &de) {
		if r, ok := de.Meta["reason"].(string); ok {
			return ReasonCode(r)
		}
	}
	return ""
}

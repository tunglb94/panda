module github.com/fairride/ai_simulation

go 1.22

require (
	github.com/fairride/dispatch v0.0.0
	github.com/fairride/pricing v0.0.0
	github.com/fairride/promotion v0.0.0
	github.com/fairride/trip v0.0.0
)

replace github.com/fairride/dispatch => ../../services/dispatch
replace github.com/fairride/pricing => ../../services/pricing
replace github.com/fairride/promotion => ../../services/promotion
replace github.com/fairride/trip => ../../services/trip
replace github.com/fairride/shared => ../../shared

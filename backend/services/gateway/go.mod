module github.com/fairride/gateway

go 1.22

require (
	github.com/fairride/booking v0.0.0
	github.com/fairride/identity v0.0.0
	github.com/fairride/shared v0.0.0
	github.com/rs/zerolog v1.33.0
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.2
)

replace (
	github.com/fairride/booking  => ../../services/booking
	github.com/fairride/dispatch => ../../services/dispatch
	github.com/fairride/identity => ../../services/identity
	github.com/fairride/pricing  => ../../services/pricing
	github.com/fairride/shared   => ../../shared
	github.com/fairride/trip     => ../../services/trip
)

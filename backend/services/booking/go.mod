module github.com/fairride/booking

go 1.22

require (
	github.com/fairride/dispatch v0.0.0
	github.com/fairride/pricing v0.0.0
	github.com/fairride/shared v0.0.0
	github.com/fairride/trip v0.0.0
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
)

replace (
	github.com/fairride/dispatch => ../../services/dispatch
	github.com/fairride/pricing => ../../services/pricing
	github.com/fairride/shared => ../../shared
	github.com/fairride/trip => ../../services/trip
)

module github.com/fairride/gateway

go 1.22

require (
	github.com/fairride/booking v0.0.0
	github.com/fairride/dispatch v0.0.0
	github.com/fairride/driver v0.0.0
	github.com/fairride/identity v0.0.0
	github.com/fairride/review v0.0.0
	github.com/fairride/shared v0.0.0
	github.com/rs/zerolog v1.33.0
	google.golang.org/grpc v1.64.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace (
	github.com/fairride/booking => ../../services/booking
	github.com/fairride/dispatch => ../../services/dispatch
	github.com/fairride/driver => ../../services/driver
	github.com/fairride/identity => ../../services/identity
	github.com/fairride/pricing => ../../services/pricing
	github.com/fairride/review => ../../services/review
	github.com/fairride/shared => ../../shared
	github.com/fairride/trip => ../../services/trip
)

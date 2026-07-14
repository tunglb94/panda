package otp

import (
	"context"

	"github.com/rs/zerolog"
)

// MockOTPProvider is the only Provider implemented today. It never sends a
// real SMS/ZNS message — it just logs the code so a developer can read it
// off the server console. RequestOTPUseCase also returns the plaintext code
// to the gateway handler, which additionally echoes it back to the client
// as debug_otp_code when APP_ENV=development (see auth_handler.go) —
// MockOTPProvider.Send is the "would have sent it over the wire" half of
// that story, kept separate so a real provider swap doesn't touch the
// handler's debug-echo logic at all.
type MockOTPProvider struct {
	log zerolog.Logger
}

// NewMockOTPProvider constructs a MockOTPProvider that logs via log.
func NewMockOTPProvider(log zerolog.Logger) *MockOTPProvider {
	return &MockOTPProvider{log: log}
}

func (p *MockOTPProvider) Send(_ context.Context, phoneNumber, code string) error {
	p.log.Info().
		Str("phone", phoneNumber).
		Str("otp_code", code).
		Msg("otp: mock provider — no real SMS/ZNS sent")
	return nil
}

// Package otp defines the delivery abstraction for one-time-password codes.
// The use case layer (identity/app) only ever depends on the Provider
// interface — swapping MockOTPProvider for a real Zalo ZNS or SMS gateway
// later means adding one new file here and changing a single line in
// gateway/cmd/server/main.go. No use case or handler changes.
package otp

import "context"

// Provider delivers an already-generated OTP code to phoneNumber. It never
// generates or validates codes itself — that stays in identity/app, so every
// Provider implementation is a pure delivery mechanism.
type Provider interface {
	Send(ctx context.Context, phoneNumber, code string) error
}

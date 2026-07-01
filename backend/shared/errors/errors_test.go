package errors_test

import (
	"fmt"
	"testing"

	"github.com/fairride/shared/errors"
)

func TestNew(t *testing.T) {
	err := errors.New(errors.CodeNotFound, "user not found")

	if err.Code != errors.CodeNotFound {
		t.Errorf("Code: got %q, want %q", err.Code, errors.CodeNotFound)
	}
	if err.Message != "user not found" {
		t.Errorf("Message: got %q, want %q", err.Message, "user not found")
	}
	if err.Cause != nil {
		t.Error("Cause: expected nil for errors.New")
	}
}

func TestWrap(t *testing.T) {
	cause := fmt.Errorf("database connection refused")
	err := errors.Wrap(errors.CodeInternalError, "failed to load user", cause)

	if err.Code != errors.CodeInternalError {
		t.Errorf("Code: got %q, want %q", err.Code, errors.CodeInternalError)
	}
	if err.Cause != cause {
		t.Error("Cause: expected the wrapped error to be the original cause")
	}
}

func TestIsCode(t *testing.T) {
	err := errors.NotFound("trip not found")

	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Error("IsCode: expected true for CodeNotFound")
	}
	if errors.IsCode(err, errors.CodeInternalError) {
		t.Error("IsCode: expected false for CodeInternalError on a NOT_FOUND error")
	}
}

func TestIsCode_Wrapped(t *testing.T) {
	inner := errors.NotFound("driver not found")
	outer := fmt.Errorf("service call failed: %w", inner)

	if !errors.IsCode(outer, errors.CodeNotFound) {
		t.Error("IsCode: expected true when DomainError is wrapped by standard error")
	}
}

func TestGetCode(t *testing.T) {
	err := errors.InvalidArgument("missing phone number")
	if errors.GetCode(err) != errors.CodeInvalidArgument {
		t.Errorf("GetCode: got %q, want %q", errors.GetCode(err), errors.CodeInvalidArgument)
	}
}

func TestGetCode_NonDomainError(t *testing.T) {
	err := fmt.Errorf("plain error")
	if errors.GetCode(err) != "" {
		t.Errorf("GetCode: expected empty string for non-DomainError, got %q", errors.GetCode(err))
	}
}

func TestWithMeta(t *testing.T) {
	err := errors.NotFound("user not found").WithMeta("user_id", "u_123")

	if err.Meta["user_id"] != "u_123" {
		t.Errorf("WithMeta: expected meta[user_id]='u_123', got %v", err.Meta["user_id"])
	}
	if err.Code != errors.CodeNotFound {
		t.Errorf("WithMeta: code must not change, got %q", err.Code)
	}
}

func TestError_Message(t *testing.T) {
	err := errors.New(errors.CodeConflict, "phone already registered")
	want := "[CONFLICT] phone already registered"
	if err.Error() != want {
		t.Errorf("Error(): got %q, want %q", err.Error(), want)
	}
}

func TestSentinels(t *testing.T) {
	cases := []struct {
		err  *errors.DomainError
		code errors.Code
	}{
		{errors.NotFound("x"), errors.CodeNotFound},
		{errors.AlreadyExists("x"), errors.CodeAlreadyExists},
		{errors.InvalidArgument("x"), errors.CodeInvalidArgument},
		{errors.PermissionDenied("x"), errors.CodePermissionDenied},
		{errors.Unauthenticated("x"), errors.CodeUnauthenticated},
		{errors.Internal("x"), errors.CodeInternalError},
		{errors.Unavailable("x"), errors.CodeUnavailable},
		{errors.ResourceExhausted("x"), errors.CodeResourceExhausted},
		{errors.Conflict("x"), errors.CodeConflict},
		{errors.Unprocessable("x"), errors.CodeUnprocessable},
		{errors.PreconditionFailed("x"), errors.CodePreconditionFailed},
	}

	for _, tc := range cases {
		if tc.err.Code != tc.code {
			t.Errorf("sentinel %q: got code %q, want %q", tc.code, tc.err.Code, tc.code)
		}
	}
}

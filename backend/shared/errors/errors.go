// Package errors defines domain error types for FAIRRIDE services.
// All service boundaries return *DomainError; infrastructure adapters
// translate these to gRPC status codes or HTTP status codes at the edge.
package errors

import (
	"errors"
	"fmt"
)

// Code is a machine-readable error classifier, mapped to gRPC/HTTP status codes at the port layer.
type Code string

const (
	CodeNotFound          Code = "NOT_FOUND"
	CodeAlreadyExists     Code = "ALREADY_EXISTS"
	CodeInvalidArgument   Code = "INVALID_ARGUMENT"
	CodePermissionDenied  Code = "PERMISSION_DENIED"
	CodeUnauthenticated   Code = "UNAUTHENTICATED"
	CodeInternalError     Code = "INTERNAL_ERROR"
	CodeUnavailable       Code = "UNAVAILABLE"
	CodeResourceExhausted Code = "RESOURCE_EXHAUSTED"
	CodeConflict          Code = "CONFLICT"
	CodeUnprocessable     Code = "UNPROCESSABLE"
	CodePreconditionFailed Code = "PRECONDITION_FAILED"
)

// DomainError is the canonical error type for all FAIRRIDE domain and application layers.
type DomainError struct {
	Code    Code
	Message string
	Cause   error
	Meta    map[string]any
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

// WithMeta returns a copy of the error with additional metadata.
func (e *DomainError) WithMeta(key string, value any) *DomainError {
	meta := make(map[string]any, len(e.Meta)+1)
	for k, v := range e.Meta {
		meta[k] = v
	}
	meta[key] = value
	return &DomainError{Code: e.Code, Message: e.Message, Cause: e.Cause, Meta: meta}
}

// New creates a new DomainError with the given code and message.
func New(code Code, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

// Wrap creates a DomainError that wraps an underlying cause.
func Wrap(code Code, message string, cause error) *DomainError {
	return &DomainError{Code: code, Message: message, Cause: cause}
}

// IsCode reports whether err (or any wrapped error) carries the given code.
func IsCode(err error, code Code) bool {
	var de *DomainError
	if errors.As(err, &de) {
		return de.Code == code
	}
	return false
}

// GetCode extracts the Code from err. Returns "" if err is not a DomainError.
func GetCode(err error) Code {
	var de *DomainError
	if errors.As(err, &de) {
		return de.Code
	}
	return ""
}

// ─── Sentinel constructors ────────────────────────────────────────────────────

func NotFound(msg string) *DomainError          { return New(CodeNotFound, msg) }
func AlreadyExists(msg string) *DomainError     { return New(CodeAlreadyExists, msg) }
func InvalidArgument(msg string) *DomainError   { return New(CodeInvalidArgument, msg) }
func PermissionDenied(msg string) *DomainError  { return New(CodePermissionDenied, msg) }
func Unauthenticated(msg string) *DomainError   { return New(CodeUnauthenticated, msg) }
func Internal(msg string) *DomainError          { return New(CodeInternalError, msg) }
func Unavailable(msg string) *DomainError       { return New(CodeUnavailable, msg) }
func ResourceExhausted(msg string) *DomainError { return New(CodeResourceExhausted, msg) }
func Conflict(msg string) *DomainError          { return New(CodeConflict, msg) }
func Unprocessable(msg string) *DomainError     { return New(CodeUnprocessable, msg) }
func PreconditionFailed(msg string) *DomainError { return New(CodePreconditionFailed, msg) }

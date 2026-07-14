package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	domainerrors "github.com/fairride/shared/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func writeGRPCError(w http.ResponseWriter, err error) {
	writeJSON(w, grpcToHTTP(err), errorResponse{Error: grpcMessage(err)})
}

func writeBadRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, errorResponse{Error: msg})
}

func grpcToHTTP(err error) int {
	st, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError
	}
	switch st.Code() {
	case codes.NotFound:
		return http.StatusNotFound
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.FailedPrecondition:
		return http.StatusUnprocessableEntity
	case codes.AlreadyExists:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func grpcMessage(err error) string {
	if st, ok := status.FromError(err); ok && st.Message() != "" {
		return st.Message()
	}
	return "internal server error"
}

func writeDomainError(w http.ResponseWriter, err error) {
	var de *domainerrors.DomainError
	if errors.As(err, &de) {
		writeJSON(w, domainCodeToHTTP(de.Code), errorResponse{Error: de.Message})
		return
	}
	writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
}

func domainCodeToHTTP(code domainerrors.Code) int {
	switch code {
	case domainerrors.CodeNotFound:
		return http.StatusNotFound
	case domainerrors.CodeAlreadyExists:
		return http.StatusConflict
	case domainerrors.CodeInvalidArgument:
		return http.StatusBadRequest
	case domainerrors.CodeUnauthenticated:
		return http.StatusUnauthorized
	case domainerrors.CodePermissionDenied:
		return http.StatusForbidden
	case domainerrors.CodePreconditionFailed:
		return http.StatusUnprocessableEntity
	case domainerrors.CodeResourceExhausted:
		return http.StatusTooManyRequests
	case domainerrors.CodeUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

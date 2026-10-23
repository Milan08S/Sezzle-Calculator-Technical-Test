package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Milan08S/Sezzle-Calculator-Technical-Test/backend/internal/calculator"
)

// Request-level errors; domain errors come from the calculator package.
var (
	errInvalidJSON     = errors.New("invalid JSON")
	errMissingField    = errors.New("missing field")
	errPayloadTooLarge = errors.New("request body too large")
)

var errorMappings = []struct {
	target error
	status int
	code   string
}{
	{errInvalidJSON, http.StatusBadRequest, "INVALID_JSON"},
	{errMissingField, http.StatusBadRequest, "MISSING_FIELD"},
	{calculator.ErrUnknownOperation, http.StatusBadRequest, "UNKNOWN_OPERATION"},
	{calculator.ErrInvalidArity, http.StatusBadRequest, "INVALID_ARITY"},
	{errPayloadTooLarge, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE"},
	{calculator.ErrDivisionByZero, http.StatusUnprocessableEntity, "DIVISION_BY_ZERO"},
	{calculator.ErrNegativeSqrt, http.StatusUnprocessableEntity, "NEGATIVE_SQRT"},
	{calculator.ErrNonFiniteResult, http.StatusUnprocessableEntity, "NON_FINITE_RESULT"},
}

// writeError is the only place that writes error responses.
// Unmapped errors become a generic 500 so internal details never reach the client.
func writeError(w http.ResponseWriter, err error) {
	for _, m := range errorMappings {
		if errors.Is(err, m.target) {
			writeJSON(w, m.status, errorResponse{errorBody{Code: m.code, Message: err.Error()}})
			return
		}
	}
	slog.Error("internal error", "err", err)
	writeJSON(w, http.StatusInternalServerError, errorResponse{errorBody{Code: "INTERNAL_ERROR", Message: "internal server error"}})
}

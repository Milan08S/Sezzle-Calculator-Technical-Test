// Package httpapi exposes the calculator over HTTP/JSON.
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/Milan08S/Sezzle-Calculator-Technical-Test/backend/internal/calculator"
)

const maxBodyBytes = 1 << 20 // 1 MB

// Calculator is what the HTTP layer needs from the domain; calculator.Registry satisfies it.
type Calculator interface {
	Execute(op string, operands []float64) (float64, error)
	Operations() []calculator.OperationInfo
}

// NewHandler returns the routes of the API.
func NewHandler(calc Calculator) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/calculate", calculate(calc))
	mux.HandleFunc("GET /api/v1/operations", operations(calc))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
	})
	return mux
}

func calculate(calc Calculator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req calculateRequest
		if err := decodeJSON(http.MaxBytesReader(w, r.Body, maxBodyBytes), &req); err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				writeError(w, fmt.Errorf("%w: limit is %d bytes", errPayloadTooLarge, tooLarge.Limit))
			} else {
				writeError(w, fmt.Errorf("%w: %v", errInvalidJSON, err))
			}
			return
		}

		op, operands, err := req.validate()
		if err != nil {
			writeError(w, err)
			return
		}

		result, err := calc.Execute(op, operands)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, calculateResponse{Operation: op, Operands: operands, Result: result})
	}
}

func operations(calc Calculator) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		infos := calc.Operations()
		resp := operationsResponse{Operations: make([]operationDTO, len(infos))}
		for i, info := range infos {
			resp.Operations[i] = operationDTO{Name: info.Name, Arity: info.Arity}
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

// validate checks required fields and unwraps the pointers.
func (req calculateRequest) validate() (string, []float64, error) {
	if req.Operation == nil {
		return "", nil, fmt.Errorf("%w: operation", errMissingField)
	}
	if req.Operands == nil {
		return "", nil, fmt.Errorf("%w: operands", errMissingField)
	}
	operands := make([]float64, len(req.Operands))
	for i, p := range req.Operands {
		if p == nil {
			return "", nil, fmt.Errorf("%w: operands[%d] is null", errMissingField, i)
		}
		operands[i] = *p
	}
	return *req.Operation, operands, nil
}

// decodeJSON decodes exactly one JSON value into v, rejecting unknown fields and trailing data.
func decodeJSON(r io.Reader, v any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("unexpected data after JSON object")
		}
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write response", "err", err)
	}
}

package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/Milan08S/Sezzle-Calculator-Technical-Test/backend/internal/calculator"
)

// fakeCalculator records its input and returns canned values.
type fakeCalculator struct {
	result      float64
	err         error
	gotOp       string
	gotOperands []float64
}

func (f *fakeCalculator) Execute(op string, operands []float64) (float64, error) {
	f.gotOp, f.gotOperands = op, operands
	return f.result, f.err
}

func (f *fakeCalculator) Operations() []calculator.OperationInfo {
	return []calculator.OperationInfo{{Name: "add", Arity: 2}, {Name: "sqrt", Arity: 1}}
}

func serve(h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, target, strings.NewReader(body)))
	return rec
}

func decodeBody[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	var v T
	if err := json.NewDecoder(rec.Body).Decode(&v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return v
}

func TestCalculateSuccess(t *testing.T) {
	fake := &fakeCalculator{result: 5}
	rec := serve(NewHandler(fake), http.MethodPost, "/api/v1/calculate", `{"operation": "divide", "operands": [10, 2]}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %s", rec.Code, rec.Body)
	}
	got := decodeBody[calculateResponse](t, rec)
	want := calculateResponse{Operation: "divide", Operands: []float64{10, 2}, Result: 5}
	if got.Operation != want.Operation || !slices.Equal(got.Operands, want.Operands) || got.Result != want.Result {
		t.Errorf("response = %+v, want %+v", got, want)
	}
	if fake.gotOp != "divide" || !slices.Equal(fake.gotOperands, []float64{10, 2}) {
		t.Errorf("calculator called with (%q, %v)", fake.gotOp, fake.gotOperands)
	}
}

func TestCalculateErrors(t *testing.T) {
	const valid = `{"operation": "add", "operands": [1, 2]}`
	tests := []struct {
		name       string
		body       string
		calcErr    error
		wantStatus int
		wantCode   string
	}{
		{"malformed JSON", `{"operation": `, nil, 400, "INVALID_JSON"},
		{"empty body", ``, nil, 400, "INVALID_JSON"},
		{"unknown field", `{"operation": "add", "operands": [1, 2], "extra": 1}`, nil, 400, "INVALID_JSON"},
		{"string operand", `{"operation": "add", "operands": ["5", 2]}`, nil, 400, "INVALID_JSON"},
		{"operation not a string", `{"operation": 1, "operands": [1, 2]}`, nil, 400, "INVALID_JSON"},
		{"trailing object", valid + `{}`, nil, 400, "INVALID_JSON"},
		{"trailing garbage", valid + ` x`, nil, 400, "INVALID_JSON"},
		{"missing operation", `{"operands": [1, 2]}`, nil, 400, "MISSING_FIELD"},
		{"null operation", `{"operation": null, "operands": [1, 2]}`, nil, 400, "MISSING_FIELD"},
		{"missing operands", `{"operation": "add"}`, nil, 400, "MISSING_FIELD"},
		{"null operands", `{"operation": "add", "operands": null}`, nil, 400, "MISSING_FIELD"},
		{"null operand", `{"operation": "add", "operands": [1, null]}`, nil, 400, "MISSING_FIELD"},
		{"unknown operation", valid, fmt.Errorf("x: %w", calculator.ErrUnknownOperation), 400, "UNKNOWN_OPERATION"},
		{"invalid arity", valid, fmt.Errorf("x: %w", calculator.ErrInvalidArity), 400, "INVALID_ARITY"},
		{"payload too large", `{"operation": "` + strings.Repeat("a", maxBodyBytes) + `"}`, nil, 413, "PAYLOAD_TOO_LARGE"},
		{"division by zero", valid, fmt.Errorf("x: %w", calculator.ErrDivisionByZero), 422, "DIVISION_BY_ZERO"},
		{"negative sqrt", valid, fmt.Errorf("x: %w", calculator.ErrNegativeSqrt), 422, "NEGATIVE_SQRT"},
		{"non-finite result", valid, fmt.Errorf("x: %w", calculator.ErrNonFiniteResult), 422, "NON_FINITE_RESULT"},
		{"unexpected error", valid, errors.New("secret db password"), 500, "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(NewHandler(&fakeCalculator{err: tt.calcErr}), http.MethodPost, "/api/v1/calculate", tt.body)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			got := decodeBody[errorResponse](t, rec)
			if got.Error.Code != tt.wantCode {
				t.Errorf("code = %q, want %q (message %q)", got.Error.Code, tt.wantCode, got.Error.Message)
			}
			if got.Error.Message == "" {
				t.Error("message is empty")
			}
			if strings.Contains(got.Error.Message, "secret") {
				t.Errorf("internal detail leaked: %q", got.Error.Message)
			}
		})
	}
}

func TestOperationsEndpoint(t *testing.T) {
	rec := serve(NewHandler(&fakeCalculator{}), http.MethodGet, "/api/v1/operations", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	got := decodeBody[operationsResponse](t, rec)
	want := []operationDTO{{Name: "add", Arity: 2}, {Name: "sqrt", Arity: 1}}
	if !slices.Equal(got.Operations, want) {
		t.Errorf("operations = %v, want %v", got.Operations, want)
	}
}

func TestHealthz(t *testing.T) {
	rec := serve(NewHandler(&fakeCalculator{}), http.MethodGet, "/healthz", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := decodeBody[healthResponse](t, rec); got.Status != "ok" {
		t.Errorf("status field = %q, want ok", got.Status)
	}
}

func TestWrongMethod(t *testing.T) {
	tests := []struct{ method, target string }{
		{http.MethodGet, "/api/v1/calculate"},
		{http.MethodPost, "/api/v1/operations"},
		{http.MethodDelete, "/healthz"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.target, func(t *testing.T) {
			rec := serve(NewHandler(&fakeCalculator{}), tt.method, tt.target, "")
			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("status = %d, want 405", rec.Code)
			}
		})
	}
}

package httpapi

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// captureLogs redirects the default slog logger to a buffer for the test's duration.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func TestRecover(t *testing.T) {
	logs := captureLogs(t)
	h := Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("kaboom") }))

	rec := serve(h, http.MethodGet, "/", "")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	got := decodeBody[errorResponse](t, rec)
	if got.Error.Code != "INTERNAL_ERROR" || strings.Contains(got.Error.Message, "kaboom") {
		t.Errorf("error = %+v, want generic INTERNAL_ERROR", got.Error)
	}
	if !strings.Contains(logs.String(), "kaboom") || !strings.Contains(logs.String(), "goroutine") {
		t.Errorf("log should contain panic value and stack, got %q", logs)
	}
}

func TestRecoverRepanicsAbortHandler(t *testing.T) {
	h := Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) }))
	defer func() {
		if rec := recover(); rec != http.ErrAbortHandler {
			t.Errorf("recovered %v, want http.ErrAbortHandler", rec)
		}
	}()
	serve(h, http.MethodGet, "/", "")
}

func TestLogging(t *testing.T) {
	logs := captureLogs(t)
	h := Logging(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) }))

	serve(h, http.MethodPost, "/api/v1/calculate", "")

	for _, want := range []string{"method=POST", "path=/api/v1/calculate", "status=418", "duration="} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("log %q missing %q", logs, want)
		}
	}
}

func TestCORS(t *testing.T) {
	const allowed = "http://localhost:5173"
	h := CORS([]string{allowed})(NewHandler(&fakeCalculator{}))

	tests := []struct {
		name        string
		method      string
		origin      string
		preflight   bool
		wantStatus  int
		wantAllowed string
	}{
		{"allowed origin", http.MethodGet, allowed, false, 200, allowed},
		{"disallowed origin", http.MethodGet, "http://evil.example", false, 200, ""},
		{"no origin", http.MethodGet, "", false, 200, ""},
		{"preflight allowed", http.MethodOptions, allowed, true, 204, allowed},
		{"preflight disallowed", http.MethodOptions, "http://evil.example", true, 405, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/operations", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.preflight {
				req.Header.Set("Access-Control-Request-Method", http.MethodPost)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantAllowed {
				t.Errorf("Allow-Origin = %q, want %q", got, tt.wantAllowed)
			}
			if got := rec.Header().Get("Vary"); got != "Origin" {
				t.Errorf("Vary = %q, want Origin", got)
			}
			if tt.preflight && tt.wantAllowed != "" && !strings.Contains(rec.Header().Get("Access-Control-Allow-Methods"), "POST") {
				t.Error("preflight response missing POST in Allow-Methods")
			}
		})
	}
}

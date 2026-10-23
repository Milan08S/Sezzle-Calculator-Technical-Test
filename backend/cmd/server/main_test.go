package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestIntegration exercises the full stack (middleware + API + real registry) over real HTTP.
func TestIntegration(t *testing.T) {
	srv := httptest.NewServer(newHandler(config{allowedOrigins: []string{"http://localhost:5173"}}))
	defer srv.Close()

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string // substring
	}{
		{"divide", "POST", "/api/v1/calculate", `{"operation":"divide","operands":[10,2]}`, 200, `"result":5`},
		{"sqrt", "POST", "/api/v1/calculate", `{"operation":"sqrt","operands":[2.25]}`, 200, `"result":1.5`},
		{"percent", "POST", "/api/v1/calculate", `{"operation":"percent","operands":[200,15]}`, 200, `"result":30`},
		{"division by zero", "POST", "/api/v1/calculate", `{"operation":"divide","operands":[1,0]}`, 422, `"code":"DIVISION_BY_ZERO"`},
		{"overflow", "POST", "/api/v1/calculate", `{"operation":"power","operands":[10,400]}`, 422, `"code":"NON_FINITE_RESULT"`},
		{"unknown operation", "POST", "/api/v1/calculate", `{"operation":"mod","operands":[1,2]}`, 400, `"code":"UNKNOWN_OPERATION"`},
		{"wrong arity", "POST", "/api/v1/calculate", `{"operation":"sqrt","operands":[1,2]}`, 400, `"code":"INVALID_ARITY"`},
		{"operations", "GET", "/api/v1/operations", "", 200, `{"name":"add","arity":2}`},
		{"healthz", "GET", "/healthz", "", 200, `{"status":"ok"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, srv.URL+tt.path, strings.NewReader(tt.body))
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d; body %s", resp.StatusCode, tt.wantStatus, body)
			}
			if !strings.Contains(string(body), tt.wantBody) {
				t.Errorf("body %s does not contain %s", body, tt.wantBody)
			}
		})
	}

	t.Run("operations are sorted and complete", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/api/v1/operations")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var got struct{ Operations []struct{ Name string } }
		if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, op := range got.Operations {
			names = append(names, op.Name)
		}
		want := []string{"add", "divide", "multiply", "percent", "power", "sqrt", "subtract"}
		if !slices.Equal(names, want) {
			t.Errorf("names = %v, want %v", names, want)
		}
	})
}

func TestStaticDir(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.html"), "<h1>app</h1>")
	mustWrite(t, filepath.Join(dir, "app.js"), "console.log(1)")
	srv := httptest.NewServer(newHandler(config{staticDir: dir}))
	defer srv.Close()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"root serves index", "GET", "/", 200, "<h1>app</h1>"},
		{"existing file", "GET", "/app.js", 200, "console.log(1)"},
		{"unknown path falls back to index", "GET", "/some/client/route", 200, "<h1>app</h1>"},
		{"traversal stays inside dir", "GET", "/../../etc/passwd", 200, "<h1>app</h1>"},
		{"api still served", "GET", "/healthz", 200, `"ok"`},
		{"unknown api path is not index", "GET", "/api/v1/nope", 404, "not found"},
		{"api wrong method keeps 405", "GET", "/api/v1/calculate", 405, "Method Not Allowed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, srv.URL+tt.path, nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != tt.wantStatus || !strings.Contains(string(body), tt.wantBody) {
				t.Errorf("got %d %q, want %d containing %q", resp.StatusCode, body, tt.wantStatus, tt.wantBody)
			}
		})
	}
}

func TestCheckHealth(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer healthy.Close()
	sick := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer sick.Close()

	if err := checkHealth(healthy.URL); err != nil {
		t.Errorf("healthy server: %v", err)
	}
	if err := checkHealth(sick.URL); err == nil {
		t.Error("503 should be unhealthy")
	}
	if err := checkHealth("http://127.0.0.1:1"); err == nil {
		t.Error("unreachable server should be unhealthy")
	}
}

func TestLoadConfig(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("STATIC_DIR", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	if got := loadConfig(); got.port != "8080" || !slices.Equal(got.allowedOrigins, []string{"http://localhost:5173"}) || got.staticDir != "" {
		t.Errorf("defaults = %+v", got)
	}

	t.Setenv("PORT", "9000")
	t.Setenv("STATIC_DIR", "/srv/www")
	t.Setenv("CORS_ALLOWED_ORIGINS", " http://a.test , ,http://b.test")
	got := loadConfig()
	if got.port != "9000" || got.staticDir != "/srv/www" || !slices.Equal(got.allowedOrigins, []string{"http://a.test", "http://b.test"}) {
		t.Errorf("from env = %+v", got)
	}
}

func TestRunShutsDownOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- run(ctx, config{port: "0"}) }()
	cancel()
	if err := <-done; err != nil {
		t.Errorf("run() = %v, want nil after graceful shutdown", err)
	}
}

func TestRunReturnsListenError(t *testing.T) {
	if err := run(context.Background(), config{port: "-1"}); err == nil {
		t.Error("run() with invalid port should fail")
	}
}

func mustWrite(t *testing.T, name, content string) {
	t.Helper()
	if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

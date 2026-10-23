// Command server runs the calculator HTTP API.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Milan08S/Sezzle-Calculator-Technical-Test/backend/internal/calculator"
	"github.com/Milan08S/Sezzle-Calculator-Technical-Test/backend/internal/httpapi"
)

type config struct {
	port           string
	allowedOrigins []string
	staticDir      string
}

func loadConfig() config {
	var origins []string
	for _, o := range strings.Split(envOr("CORS_ALLOWED_ORIGINS", "http://localhost:5173"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}
	return config{
		port:           envOr("PORT", "8080"),
		allowedOrigins: origins,
		staticDir:      os.Getenv("STATIC_DIR"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	healthcheck := flag.Bool("healthcheck", false, "GET /healthz on the local server and exit 0 if healthy, 1 otherwise")
	flag.Parse()
	cfg := loadConfig()

	if *healthcheck {
		if err := checkHealth("http://localhost:" + cfg.port + "/healthz"); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

// run serves until ctx is cancelled, then shuts down gracefully.
func run(ctx context.Context, cfg config) error {
	srv := &http.Server{
		Addr:              ":" + cfg.port,
		Handler:           newHandler(cfg),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	slog.Info("listening", "addr", srv.Addr, "static_dir", cfg.staticDir)

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// newHandler wires registry → API → middleware (Recover outermost).
func newHandler(cfg config) http.Handler {
	var h http.Handler = httpapi.NewHandler(calculator.NewRegistry())
	if cfg.staticDir != "" {
		// API paths stay on the API mux so they keep their 404/405 behaviour.
		mux := http.NewServeMux()
		mux.Handle("/api/", h)
		mux.Handle("/healthz", h)
		mux.Handle("/", spaHandler(cfg.staticDir))
		h = mux
	}
	return httpapi.Recover(httpapi.Logging(httpapi.CORS(cfg.allowedOrigins)(h)))
}

// spaHandler serves files from dir and falls back to index.html for unknown paths,
// so client-side routes survive a page reload.
func spaHandler(dir string) http.Handler {
	root := http.Dir(dir) // http.Dir.Open rejects paths escaping dir
	files := http.FileServer(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := root.Open(path.Clean(r.URL.Path))
		if err != nil {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		f.Close()
		files.ServeHTTP(w, r)
	})
}

func checkHealth(url string) error {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("unhealthy: " + resp.Status)
	}
	return nil
}

# Prompts

<!-- Paste prompts below. Claude appends a one-line outcome per phase. -->

# Task: Full-stack calculator (Sezzle take-home assignment)

Build a full-stack calculator: a Go REST API backend and a React + TypeScript frontend that consumes it. The evaluation prioritizes correctness, clarity, maintainability, idiomatic code, and test quality over extra features. Target effort: 2–4 hours of focused work. Do not over-engineer.

I am learning Go. When you write Go code, briefly explain any non-obvious idiom (error wrapping, consumer-side interfaces, table-driven tests, etc.) in your messages to me, not in code comments.

## Working rules

- Work in the phases listed at the end. After each phase: run all tests, run the linters, give to me one or more small commits commands with Conventional Commit messages I will do it by myself, then STOP and summarize what you did and any decision you made. Wait for my approval before the next phase.
- Before Phase 1, show me the full directory tree and the API contract and wait for approval.
- If you want to deviate from this spec, ask first. Do not add dependencies beyond the ones listed without asking.
- Maintain `PROMPTS.md` at the repo root: I will paste my prompts there; you append a one-line summary of each phase's outcome.
- We will work in branches based on development with a prefix (feature/, hotfix/, etc).
- The only way we can Push to development or main is with a PR.
- We have to create a CI workflow for automatic tests

## Repository layout

```
/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── calculator/   # domain: pure, MUST NOT import net/http
│   │   │   ├── operation.go
│   │   │   ├── operations.go
│   │   │   ├── registry.go
│   │   │   └── errors.go
│   │   └── httpapi/      # HTTP transport: depends on calculator, never the reverse
│   │       ├── handler.go
│   │       ├── dto.go
│   │       ├── errors.go
│   │       └── middleware.go
│   └── go.mod
├── frontend/             # Vite + React + TypeScript
├── Dockerfile
├── Makefile
├── README.md
└── PROMPTS.md
```

## Backend (Go 1.23+)

Use only the standard library (`net/http` with Go 1.22+ routing patterns, `encoding/json`, `log/slog`, `math`, `errors`). No web frameworks, no DI frameworks, no `pkg/` directory, no `utils`/`common`/`models` packages, no service/use-case layer between handler and registry. Avoid package-name stutter (`calculator.Registry`, not `calculator.CalculatorRegistry`).

### Domain: `internal/calculator`

- `Operation` is a struct: `Name string`, `Arity int`, `Apply func(operands []float64) (float64, error)`.
- Individual operation functions are unexported. The only public entry point for executing operations is `Registry.Execute(name string, operands []float64) (float64, error)`.
- `Registry.Execute` centralizes the shared rules in this order: unknown operation → `ErrUnknownOperation`; `len(operands) != Arity` → `ErrInvalidArity`; call `Apply`; result is NaN or ±Inf → `ErrNonFiniteResult`. Wrap errors with context using `fmt.Errorf("...: %w", err)` so `errors.Is` works.
- `Registry.Operations()` returns name and arity of every operation, sorted by name.
- Operations:
  - `add`, `subtract`, `multiply`, `divide`, `power`: arity 2.
  - `sqrt`: arity 1. Negative operand → `ErrNegativeSqrt`.
  - `percent`: arity 2, defined as `a * b / 100` ("b percent of a"). Document this assumption.
  - `divide`: divisor equal to 0 → `ErrDivisionByZero`.
- Sentinel errors live in `errors.go`.
- Arithmetic uses `float64` (IEEE 754). Do not use a decimal library; document the precision trade-off in the README.

### Transport: `internal/httpapi`

- Declare a consumer-side interface in `httpapi`:
  `type Calculator interface { Execute(op string, operands []float64) (float64, error); Operations() []calculator.OperationInfo }`
  (adjust the info type name as appropriate). `calculator.Registry` satisfies it implicitly. Handler tests use a fake.
- Endpoints:
  - `POST /api/v1/calculate` — request `{"operation": "divide", "operands": [10, 2]}`, response `200 {"operation": "divide", "operands": [10, 2], "result": 5}`.
  - `GET /api/v1/operations` — response `200 {"operations": [{"name": "add", "arity": 2}, ...]}`.
  - `GET /healthz` — `200 {"status": "ok"}`.
- Request validation:
  - Decode with `json.Decoder` + `DisallowUnknownFields()`; reject trailing data after the JSON object.
  - Limit body size with `http.MaxBytesReader` (1 MB).
  - DTO uses `Operation *string` and `Operands []*float64` so that a missing `operation`, a missing `operands` field, and `null` elements inside `operands` are all detected and rejected (a plain `float64` would silently decode them as 0).
  - Numbers sent as strings (`"5"`) must be rejected.
- Error response body (single format for every error, produced only by an unexported `writeError` function):
  `{"error": {"code": "DIVISION_BY_ZERO", "message": "division by zero"}}`
- Error mapping (in `httpapi/errors.go`, using `errors.Is`; the mapping table is unexported):

  | Code | Status | Cause |
  |---|---|---|
  | `INVALID_JSON` | 400 | Malformed JSON, unknown fields, wrong types, trailing data |
  | `MISSING_FIELD` | 400 | `operation` or `operands` missing, or a `null` operand |
  | `UNKNOWN_OPERATION` | 400 | `ErrUnknownOperation` |
  | `INVALID_ARITY` | 400 | `ErrInvalidArity` |
  | `PAYLOAD_TOO_LARGE` | 413 | Body exceeds limit |
  | `DIVISION_BY_ZERO` | 422 | `ErrDivisionByZero` |
  | `NEGATIVE_SQRT` | 422 | `ErrNegativeSqrt` |
  | `NON_FINITE_RESULT` | 422 | `ErrNonFiniteResult` |
  | `INTERNAL_ERROR` | 500 | Anything else; log details, never leak them to the client |

- Responses set `Content-Type: application/json`. Wrong HTTP method returns 405 (handled by the 1.22 mux patterns).
- Middleware in `httpapi/middleware.go` (same package, not a separate one): `Recover` (panic → 500 via `writeError`, logs stack), `Logging` (slog: method, path, status, duration), `CORS` (allowed origins from config).

### `cmd/server/main.go` (composition root)

- Config from env vars with defaults: `PORT` (8080), `CORS_ALLOWED_ORIGINS` (`http://localhost:5173`), `STATIC_DIR` (empty; if set, serve the built frontend from that directory with `http.FileServer`, falling back to `index.html` for unknown non-API paths).
- Build registry → handler → middleware chain (`Recover` outermost).
- `http.Server` with `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`.
- Graceful shutdown with `signal.NotifyContext` (SIGINT, SIGTERM) and `server.Shutdown` with a timeout.

### Backend tests

- `calculator`: table-driven tests for every operation (normal cases, negative numbers, decimals, zero, large values), every registry rule, and `Operations()` ordering. Target ~100% coverage for this package.
- `httpapi`: `httptest.NewRecorder` with a fake `Calculator`. Cover: success, every row of the error table, unknown fields, `null` operand, string operand, trailing data, oversized body, wrong method, CORS headers, panic recovery.
- One integration test using `httptest.NewServer` with the real registry.
- Run with `-race`. Produce `coverage.out` and `coverage.html`.
- Lint with `go vet` and `gofmt -l` (must be clean).

## Frontend (Vite + React + TypeScript, strict mode)

- Dependencies: React, Vite, Vitest, React Testing Library, `@testing-library/user-event`, jsdom. No UI or state libraries. Plain CSS (CSS modules are fine).
- Structure:
  - `src/api/calculatorClient.ts` — the only module that calls `fetch`. Typed request/response/error types mirroring the backend contract. Base URL from `import.meta.env.VITE_API_BASE_URL`, defaulting to the same origin. Maps network failures and non-2xx responses to a typed error.
  - `src/state/calculatorReducer.ts` — pure `useReducer` state machine with explicit states (`idle`, `inputtingFirst`, `operatorPending`, `inputtingSecond`, `loading`, `result`, `error`). Handles digit input, single decimal point, sign toggle, backspace, clear, binary operators, unary `sqrt`, chaining (pressing an operator after a result uses the result as the first operand).
  - `src/components/` — `Calculator`, `Display`, `Keypad`.
- Client-side validation before calling the API (no empty operand, no repeated decimal point, max input length). The backend remains the source of truth; display backend error messages by `code`.
- Display formatting: avoid showing floating-point noise (e.g. format with up to 12 significant digits, strip trailing zeros). Document this.
- Keyboard support: digits, `.`, `+ - * /`, `^`, `%`, `Enter`/`=`, `Backspace`, `Escape` (clear).
- Accessibility: buttons have accessible names; the display uses `aria-live="polite"`; visible focus states.
- Responsive: usable from 320px width up; grid-based keypad.
- Vite dev server proxies `/api` to `http://localhost:8080`.
- Tests: reducer unit tests (all transitions), client tests with mocked `fetch` (success, error code, network failure), component tests for key user flows (`5 + 3 =` shows `8`; division by zero shows the error; keyboard input works; loading state disables input). Generate a coverage report with `vitest --coverage`.
- Lint with ESLint (TypeScript + react-hooks rules) and `tsc --noEmit` (must be clean).

## Docker

- `backend/Dockerfile`: multi-stage. Go stage builds a static binary (`CGO_ENABLED=0`, `-trimpath`, `-ldflags="-s -w"`); final stage `gcr.io/distroless/static-debian12:nonroot`, `EXPOSE 8080`.
- The backend binary supports a `-healthcheck` flag: it performs `GET http://localhost:$PORT/healthz` and exits 0 on 200, 1 otherwise. This is required because distroless has no curl/wget.
- `frontend/Dockerfile`: multi-stage. Node stage runs `npm ci` and `npm run build`; final stage `nginx:1.27-alpine` with the build output and a custom `frontend/nginx.conf`.
- `frontend/nginx.conf`: serves static files with `try_files $uri $uri/ /index.html;`, proxies `/api/` and `/healthz` to `http://backend:8080`, sets long cache headers for hashed assets and `no-cache` for `index.html`, enables gzip.
- `compose.yaml` at the repo root:
  - `backend` service: builds `./backend`, no published ports, healthcheck `["CMD", "/server", "-healthcheck"]`.
  - `frontend` service: builds `./frontend`, publishes `8080:80`, `depends_on: backend: condition: service_healthy`.
- `docker compose up --build` serves the whole app at `http://localhost:8080` (no CORS needed: same origin through nginx).
- Add `.dockerignore` files for both services.

## Makefile

Targets: `backend-run`, `backend-test`, `backend-coverage`, `frontend-install`, `frontend-dev`, `frontend-test`, `frontend-coverage`, `test` (both), `lint` (both), `docker-build`, `docker-run`.

## README.md

Sections:
1. Overview and screenshot placeholder.
2. Prerequisites (Go, Node, Docker versions).
3. Running locally (backend, frontend, both via Makefile) and with Docker.
4. API reference: every endpoint, request/response schemas, the error code table, and `curl` examples for success and for each error class.
5. Testing and coverage (commands and where reports are written); include the current coverage percentages.
6. Design decisions, each with the rationale and the rejected alternative:
   - Binary/unary operations instead of expression parsing.
   - Single `/calculate` endpoint + operation registry (Strategy + Registry, Open/Closed) instead of one endpoint per operation.
   - Package-by-responsibility layout; domain independent of HTTP; consumer-side interface.
   - Standard library only.
   - `float64` and display-side formatting instead of a decimal library.
   - Pointer DTO fields to distinguish missing/null from zero.
   - NaN/Inf rejection (JSON cannot encode them).
   - Frontend state machine with `useReducer`.
   - Single container serving API + static assets.
7. Assumptions (percent semantics, precision, limits).
8. Possible future work (expression parsing, history, rate limiting, OpenAPI spec).
9. AI usage: link to `PROMPTS.md`.

## Phases

1. Backend domain (`calculator`) + tests.
2. Backend transport (`httpapi`), `main.go` + tests, integration test, coverage.
3. Frontend API client + reducer + tests.
4. Frontend components, styling, keyboard support + tests, coverage.
5. Dockerfile, Makefile, `.dockerignore`.
6. README and final review: run every Makefile target, verify `docker run` end to end, check lint and formatting, list anything incomplete.

## Phase outcomes

- **Phase 0 (plan):** Approved directory tree and API contract; chose two Docker images + compose and a CI workflow grown per phase.
- **Phase 1 (backend domain):** `internal/calculator` (Operation, Registry with Execute/Operations, 7 operations, sentinel errors) with table-driven tests at 100% coverage; CI backend job (gofmt, vet, `go test -race`).
- **Phase 2 (backend transport):** `internal/httpapi` (consumer-side `Calculator` interface, pointer DTO validation, single `writeError` with errors.Is mapping table, Recover/Logging/CORS middleware) and `cmd/server` (env config, STATIC_DIR SPA fallback, timeouts, graceful shutdown, `-healthcheck`); handler/middleware/integration tests, total coverage 91%.

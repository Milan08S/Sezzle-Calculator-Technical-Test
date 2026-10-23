.PHONY: backend-run backend-test backend-coverage frontend-install frontend-dev frontend-test frontend-coverage test lint docker-build docker-run

backend-run:
	cd backend && go run ./cmd/server

backend-test:
	cd backend && go test -race ./...

backend-coverage:
	cd backend && go test -race -coverprofile=coverage.out ./... \
		&& go tool cover -html=coverage.out -o coverage.html \
		&& go tool cover -func=coverage.out | tail -1

frontend-install:
	cd frontend && npm ci

frontend-dev:
	cd frontend && npm run dev

frontend-test:
	cd frontend && npm test

frontend-coverage:
	cd frontend && npm run coverage

test: backend-test frontend-test

lint:
	@cd backend && out="$$(gofmt -l .)" && if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
	cd backend && go vet ./...
	cd frontend && npm run lint && npm run typecheck

docker-build:
	docker compose build

docker-run:
	docker compose up --build

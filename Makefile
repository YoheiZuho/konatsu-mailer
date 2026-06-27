.PHONY: build run dev migrate-up migrate-down test test-integration cover cover-html lint fmt

build:
	docker compose build

run:
	docker compose up --build

dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

migrate-up:
	docker compose run --rm backend migrate -path /app/migrations -database "$$DATABASE_URL" up

migrate-down:
	docker compose run --rm backend migrate -path /app/migrations -database "$$DATABASE_URL" down 1

test:
	go test -race -count=1 ./...

# Integration tests against a throwaway PostgreSQL (requires Docker).
test-integration:
	@docker rm -f konatsu-it-pg >/dev/null 2>&1 || true
	docker run -d --name konatsu-it-pg -e POSTGRES_USER=it -e POSTGRES_PASSWORD=it \
		-e POSTGRES_DB=it -p 55432:5432 postgres:17-alpine >/dev/null
	@echo "waiting for postgres..."; \
	until docker exec konatsu-it-pg pg_isready -U it -d it >/dev/null 2>&1; do sleep 1; done
	-TEST_DATABASE_URL="postgres://it:it@localhost:55432/it?sslmode=disable" \
		go test -tags=integration -count=1 ./internal/store/... ./internal/api/...
	@docker rm -f konatsu-it-pg >/dev/null 2>&1 || true

# C0 (statement) coverage across the logic packages.
cover:
	go test -cover -count=1 ./...

cover-html:
	go test -coverprofile=coverage.out -count=1 ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

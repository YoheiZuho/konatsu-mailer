.PHONY: build run dev migrate-up migrate-down test lint fmt

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

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

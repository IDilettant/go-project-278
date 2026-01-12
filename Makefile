ENV_FILE := $(CURDIR)/.env

define load_env
	set -a; [ -f $(ENV_FILE) ] && . $(ENV_FILE); set +a;
endef

test:
	go test -v ./... -race

test-integration:
	$(load_env) \
	go test -v ./... -race -tags=integration

lint:
	golangci-lint run ./...

build:
	go build -o bin/shortener ./cmd/api

cover:
	go test -v ./... -race -count=1 -tags=integration \
		-covermode=atomic -coverpkg=./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

dev:
	$(load_env) air

db-up:
	docker compose up -d

db-down:
	docker compose down

migrate-up:
	$(load_env) \
	goose -dir ./db/migrations postgres "$$DATABASE_URL" up

migrate-down:
	$(load_env) \
	goose -dir ./db/migrations postgres "$$DATABASE_URL" down

sqlc:
	sqlc generate

swagger:
	go generate ./...

.PHONY: test lint build dev sqlc swagger migrate-up migrate-down db-up db-down

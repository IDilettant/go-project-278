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

docs-open-up:
	$(load_env) \
	docker compose -f docker-compose.docs.yml up -d --remove-orphans
	$(load_env) \
	sh -c 'URL="$${DOCS_URL}:$${DOCS_PORT}"; \
	echo "API docs: $${URL}"; \
	(command -v xdg-open >/dev/null && xdg-open "$${URL}" >/dev/null 2>&1 || true); \
	(command -v open >/dev/null && open "$${URL}" >/dev/null 2>&1 || true)'

docs-down:
	docker compose -f docker-compose.docs.yml down


.PHONY: test lint build dev sqlc docs migrate-up migrate-down db-up db-down

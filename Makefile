test:
	go test -v ./... -race

lint:
	golangci-lint run ./...

build:
	go build -o bin/shortener ./cmd/api

cover:
	go test -v ./... -race -count=1 -covermode=atomic -coverprofile=coverage.out
	go tool cover -func=coverage.out

dev:
	set -a; . ./.env; set +a; air

.PHONY: test lint build dev

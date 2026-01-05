test:
	go test -v ./... -race

lint:
	golangci-lint run ./...

build:
	go build -o bin/shortener ./cmd/api

dev:
	air

.PHONY: test lint build dev

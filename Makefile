BINARY_NAME := goobservabilitystack
BUILD_DIR   := bin

.PHONY: deps build test lint up down

## deps: download and tidy Go module dependencies.
deps:
	go mod download
	go mod tidy

## build: compile the server binary into $(BUILD_DIR)/$(BINARY_NAME).
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

## test: run all tests with the race detector enabled.
test:
	go test -race ./...

## lint: run golangci-lint across all packages.
lint:
	golangci-lint run ./...

## up: build and start the full stack with Docker Compose.
up:
	docker compose up --build

## down: stop and remove the full demo stack.
down:
	docker compose down

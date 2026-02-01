.PHONY: build run dev test test-coverage lint fmt vet clean docker-build migrate-up help

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=ugc
BINARY_DIR=bin
DOCKER_IMAGE=ugc:latest

build: ## Build the application
	go build -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/ugc

run: ## Run the application
	go run ./cmd/ugc

dev: ## Run with hot reload (requires air)
	air

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage report
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter (golangci-lint)
	golangci-lint run

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

clean: ## Clean build artifacts
	rm -rf $(BINARY_DIR)/ coverage.out coverage.html

docker-build: ## Build docker image
	docker build -t $(DOCKER_IMAGE) .

migrate-up: ## Run database migrations
	go run ./cmd/ugc migrate up

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Makefile for osquery extension

# Variables
NAME = browser_extend_extension
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Build the extension
build:
	go build -o $(NAME) cmd/$(NAME)/main.go

# Build for all supported platforms
build-all: build-linux build-darwin build-windows

# Build for Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(NAME)-linux-amd64 cmd/$(NAME)/main.go
	GOOS=linux GOARCH=arm64 go build -o $(NAME)-linux-arm64 cmd/$(NAME)/main.go

# Build for macOS
build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o $(NAME)-darwin-amd64 cmd/$(NAME)/main.go
	GOOS=darwin GOARCH=arm64 go build -o $(NAME)-darwin-arm64 cmd/$(NAME)/main.go

# Build for Windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(NAME)-windows-amd64.exe cmd/$(NAME)/main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(NAME) $(NAME)-*

# Install dependencies
deps:
	go mod tidy

# Run linter
lint:
	golangci-lint run

# Run linter with auto-fix
lint-fix:
	golangci-lint run --fix

# Run all checks
check: lint test

.PHONY: build build-all build-linux build-darwin build-windows test test-coverage clean deps lint lint-fix check
# Makefile for osquery extension

# Variables
BUILD_PATH = cmd/browser_extend_extension
NAME = browser-extend-extension
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Build the extension
build:
	go build -o $(NAME) $(BUILD_PATH)/main.go

# Build for all supported platforms
build-all: build-linux build-darwin build-windows

# Build for current supported platforms
build-current: build-linux build-darwin-arm64 build-windows-amd64

# Build for Linux
build-linux: build-linux-amd64 build-linux-arm64

# Build for Linux (AMD64)
build-linux-amd64: clean cache-clean
	GOOS=linux GOARCH=amd64 go build -o $(NAME)-linux-amd64 $(BUILD_PATH)/main.go

# Build for Linux (ARM64)
build-linux-arm64: clean cache-clean
	GOOS=linux GOARCH=arm64 go build -o $(NAME)-linux-arm64 $(BUILD_PATH)/main.go

# Build for macOS
build-darwin: build-darwin-amd64 build-darwin-arm64

# Build for macOS (AMD64)
build-darwin-amd64:clean cache-clean
	GOOS=darwin GOARCH=amd64 go build -o $(NAME)-darwin-amd64 $(BUILD_PATH)/main.go

# Build for macOS (ARM64)
build-darwin-arm64: clean cache-clean
	GOOS=darwin GOARCH=arm64 go build -a -o $(NAME)-darwin-arm64 $(BUILD_PATH)/main.go

# Build for Windows (AMD64 only)
build-windows: build-windows-amd64

# Build for Windows (AMD64)
build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -o $(NAME)-windows-amd64.exe $(BUILD_PATH)/main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# cache clean
cache-clean:
	go clean -cache -modcache -i -r

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

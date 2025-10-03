# Suggested Commands (updated)

- Build: go build -o osquery-browser-history cmd/browser_extend_extension/main.go
- Make targets: make build, make build-all, make test, make lint, make check
- Tests: go test -v ./..., go test -run TestName ./...
- Format: go fmt ./...
- Lint: golangci-lint run
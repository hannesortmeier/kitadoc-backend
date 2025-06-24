.PHONY: all build test test-e2e clean

# Default target
all: build

# Build the application
build:
	go build -o bin/kitadoc-backend ./main.go

build-amd64:
	env GOOS=linux GOARCH=amd64 go build -o bin/kitadoc-backend-linux-amd64 ./main.go

build-arm64:
	env GOOS=linux GOARCH=arm64 go build -o bin/kitadoc-backend-linux-arm64 ./main.go

# Run tests
test:
	go test -v ./... -coverprofile=coverage.txt -covermode=atomic

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Install dependencies
deps:
	go mod tidy
	go mod download

.PHONY: all build test test-e2e clean test-db run-dev

# Default target
all: build

# Build the application
build:
	go build -tags=viper_bind_struct -o bin/kitadoc-backend ./main.go

# Run test application
run-dev:
	KINDERGARTEN_LOG_LEVEL=debug KINDERGARTEN_LOG_FORMAT=text KINDERGARTEN_SERVER_JWT_SECRET=dsjfhaksdfhasfh KINDERGARTEN_ADMIN_USERNAME=admin KINDERGARTEN_ADMIN_PASSWORD=admin KINDERGARTEN_NORMAL_USERNAME=teacher KINDERGARTEN_NORMAL_PASSWORD=teacher bin/kitadoc-backend

build-amd64:
	env GOOS=linux CGO_ENABLED=1 GOARCH=amd64 go build -o bin/kitadoc-backend-linux-amd64 ./main.go

build-arm64:
	env GOOS=linux CGO_ENABLED=1 GOARCH=arm64 go build -o bin/kitadoc-backend-linux-arm64 ./main.go

# Run tests
test:
	go test -v ./... -coverprofile=coverage.txt -covermode=atomic

pre-commit:
	pre-commit run --all-files

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Install dependencies
deps:
	go mod tidy
	go mod download

# Create test sqlite database
test-db:
	rm -rf test.db
	sqlite3 test.db < database/data_model.sql
	sqlite3 test.db < database/sample_data.sql

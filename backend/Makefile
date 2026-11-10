.PONY: help build run test clean migrate setup install seed

help:
	@echo "Available commands:"
	@echo "  make setup     - Complete initial setup (install deps, create DB, migrate, build)"
	@echo "  make build     - Build the application"
	@echo "  make run       - Run the application"
	@echo "  make test      - Run tests"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make migrate   - Run database migrations"
	@echo "  make seed      - Seed database with test data"
	@echo "  make deps      - Download dependencies"
	@echo "  make install    - Install all prerequisites (Go, MySQL)"

setup:
	@echo "Running complete setup..."
	@./setup.sh

build:
	@echo "Building application..."
	@go build -o bin/server ./cmd/server

run:
	@echo "Running application..."
	@go run ./cmd/server/main.go

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean

migrate:
	@echo "Running database migrations..."
	@cd migrations && go run migrate.go

seed:
	@echo "Seeding database with test data..."
	@cd tools/seed && go run main.go

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

install:
	@echo "Installing prerequisites..."
	@echo "Note: This script only checks and guides installation. Manual installation may be required."
	@echo ""
	@echo "Checking prerequisites..."
	@which go || (echo "Go not found. Please install from https://golang.org/dl/"; exit 1)
	@echo "✓ Go installed: $$(go version)"
	@which mysql || (echo "MySQL not found. Please install with: sudo apt-get install mysql-server"; exit 1)
	@echo "✓ MySQL installed: $$(mysql --version)"

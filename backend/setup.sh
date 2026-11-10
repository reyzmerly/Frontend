#!/bin/bash

# UNSTOPPABLE API Setup Script
# This script performs all initial setup for the University of Kigali Invoices & Payments API

set -e

# Add XAMPP to PATH if it exists
if [ -d /opt/lampp ]; then
    export PATH="$PATH:/opt/lampp/bin"
fi

echo "=========================================="
echo "UNSTOPPABLE API Setup"
echo "=========================================="
echo ""

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${NC}→ $1"
}

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    print_warning "Running as root is not recommended. Please run as a regular user."
fi

# Check prerequisites
print_info "Checking prerequisites..."

# Check for Go
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.27 or higher from https://golang.org/dl/"
    echo "For Ubuntu/Debian:"
    echo "  wget https://go.dev/dl/go1.27.3.linux-amd64.tar.gz"
    echo "  sudo tar -C /usr/local -xzf go1.27.3.linux-amd64.tar.gz"
    echo "  export PATH=$PATH:/usr/local/go/bin"
    echo "  echo 'export PATH=\$PATH:/usr/local/go/bin' >> ~/.bashrc"
    exit 1
fi
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
print_success "Go installed: $GO_VERSION"

# Check for MySQL (XAMPP uses mysqld directly)
if ! command -v mysql &> /dev/null; then
    print_warning "MySQL client not found in PATH. Checking for XAMPP MySQL..."
    if [ -f /opt/lampp/bin/mysql ]; then
        export PATH="$PATH:/opt/lampp/bin"
        print_success "XAMPP MySQL found at /opt/lampp/bin/mysql"
    elif [ -f /usr/bin/mysql ]; then
        export PATH="$PATH:/usr/bin"
        print_success "MySQL found at /usr/bin/mysql"
    else
        print_warning "MySQL client not found. Will attempt to connect using Go driver."
    fi
fi

# Check for .env file
if [ ! -f .env ]; then
    print_warning ".env file not found. Creating from .env.example..."
    if [ -f .env.example ]; then
        cp .env.example .env
        print_success ".env file created"
        print_warning "Please edit .env with your actual database credentials before running the application."
    else
        print_error ".env.example not found. Cannot create .env file."
        exit 1
    fi
else
    print_success ".env file exists"
fi

# Install Go dependencies
print_info "Installing Go dependencies..."
go mod download
go mod tidy
print_success "Dependencies installed"

# Check if MySQL is running
print_info "Checking MySQL connection..."
if command -v mysql &> /dev/null; then
    # Read database credentials from .env
    DB_HOST=$(grep DB_HOST .env | cut -d '=' -f2)
    DB_PORT=$(grep DB_PORT .env | cut -d '=' -f2)
    DB_USER=$(grep DB_USER .env | cut -d '=' -f2)
    DB_PASSWORD=$(grep DB_PASSWORD .env | cut -d '=' -f2)
    DB_NAME=$(grep DB_NAME .env | cut -d '=' -f2)

    if [ -n "$DB_PASSWORD" ]; then
        if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1;" &> /dev/null; then
            print_success "MySQL connection successful"
        else
            print_warning "Cannot connect to MySQL with provided credentials. Will attempt to create database without connection check."
        fi
    else
        print_warning "DB_PASSWORD is empty. Skipping connection check."
    fi
else
    print_warning "MySQL client not in PATH. Skipping connection check - will use Go driver."
fi

# Create database if it doesn't exist
print_info "Creating database if not exists..."
DB_NAME=$(grep DB_NAME .env | cut -d '=' -f2)
DB_USER=$(grep DB_USER .env | cut -d '=' -f2)
DB_PASSWORD=$(grep DB_PASSWORD .env | cut -d '=' -f2)
DB_HOST=$(grep DB_HOST .env | cut -d '=' -f2)
DB_PORT=$(grep DB_PORT .env | cut -d '=' -f2)

if command -v mysql &> /dev/null; then
    if [ -n "$DB_PASSWORD" ]; then
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "CREATE DATABASE IF NOT EXISTS $DB_NAME;" 2>/dev/null
    else
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -e "CREATE DATABASE IF NOT EXISTS $DB_NAME;" 2>/dev/null
    fi
fi
print_success "Database '$DB_NAME' ready (will be created by migrations if needed)"

# Run migrations
print_info "Running database migrations..."
cd migrations
go run migrate.go
cd ..
print_success "Migrations completed"

# Seed database with test data
print_info "Seeding database with test data..."
cd tools/seed
go run main.go
cd ../..
print_success "Seed data inserted"

# Build the application
print_info "Building application..."
go build -o bin/server ./cmd/server
print_success "Application built successfully"

# Create logs directory
mkdir -p logs
print_success "Logs directory created"

echo ""
echo "=========================================="
print_success "Setup completed successfully!"
echo "=========================================="
echo ""
echo "To start the application:"
echo "  make run"
echo "  or"
echo "  ./bin/server"
echo ""
echo "The server will start on http://localhost:9107"
echo ""
echo "API Documentation:"
echo "  Health Check: GET http://localhost:9107/health"
echo "  Base Path: /api/v1"
echo ""
print_warning "Make sure to:"
echo "  1. Update .env with your actual database credentials"
echo "  2. Change JWT_SECRET to a secure random value in production"
echo "  3. Ensure MoPay credentials are correct for real payments"
echo ""

# University of Kigali Invoices & Payments System - Backend API

A comprehensive backend API for managing student invoices, payments, and financial records for the University of Kigali.

## Project Overview

This system provides:
- RESTful API for students to view invoices, track payments, and check outstanding balances
- API endpoints for authorized university staff to manage invoices and record payments
- Finance reporting capabilities including account reports and PDF/CSV exports
- Role-based access control for different user types
- JWT authentication for secure API access

## Project Structure

```
backend/
├── cmd/           # Application entry point
├── internal/      # Internal packages (config, middleware, routes)
├── services/      # Service-oriented architecture
│   ├── user/      # User service (controllers, models)
│   ├── student/   # Student service (controllers, models)
│   ├── invoice/   # Invoice service (controllers, models)
│   ├── payment/   # Payment service (controllers, models)
│   ├── finance/   # Finance service (reports, exports)
│   ├── fee/       # Fee service (controllers, models)
│   ├── admin/     # Admin service (controllers, models)
│   └── webhook/   # Payment webhook handlers
├── shared/        # Shared components (repositories)
├── migrations/    # Database migrations
├── tools/         # Database utility scripts
├── Makefile       # Build and run commands
└── setup.sh       # Initial setup script
```

## Technology Stack

- **Language**: Go (Golang)
- **Database**: MySQL 8.0+
- **Architecture**: Service-Oriented Architecture (SOA)
- **Driver**: go-sql-driver/mysql
- **Router**: Chi (github.com/go-chi/chi/v5)
- **Authentication**: JWT (github.com/golang-jwt/jwt/v5)
- **Validation**: validator (github.com/go-playground/validator/v10)
- **PDF Generation**: github.com/jung-kurt/gofpdf

## Getting Started

### Prerequisites
- Go 1.21+
- MySQL 8.0+ (XAMPP MySQL recommended for development)
- Git

### Setup Instructions

1. Clone the repository:
   ```bash
   git clone https://github.com/kwizerisezerano/TEAM3.git
   cd TEAM3/backend
   ```

2. Run the setup script:
   ```bash
   ./setup.sh
   ```

   The setup script will:
   - Create a `.env` file with default configuration
   - Create the `university_invoices` database
   - Run database migrations
   - Install Go dependencies

3. Run the server:
   ```bash
   make run
   ```

The API will be available at `http://localhost:9107`

### Manual Setup

If you prefer manual setup:

1. Navigate to backend directory:
   ```bash
   cd backend
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Configure environment:
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

4. Create database and run migrations:
   ```bash
   mysql -u root -p -e "CREATE DATABASE university_invoices;"
   mysql -u root -p university_invoices < migrations/schema.sql
   ```

5. Run the server:
   ```bash
   make run
   ```

## API Documentation

The API provides RESTful endpoints with JWT authentication at `/api/v1/`.

### Authentication
- `POST /api/v1/auth/login` - User login and JWT token generation

### Users
- `GET /api/v1/users` - List users (admin only)
- `GET /api/v1/users/{id}` - Get user by ID
- `PUT /api/v1/users/{id}` - Update user

### Students
- `GET /api/v1/students` - List students
- `GET /api/v1/students/{id}` - Get student by ID
- `POST /api/v1/students` - Create student
- `PUT /api/v1/students/{id}` - Update student

### Fee Types
- `GET /api/v1/fee-types` - List fee types
- `POST /api/v1/fee-types` - Create fee type
- `PUT /api/v1/fee-types/{id}` - Update fee type
- `DELETE /api/v1/fee-types/{id}` - Delete fee type

### Invoices
- `GET /api/v1/invoices` - List invoices (supports search by student number, name, invoice number)
- `GET /api/v1/invoices/{invoiceId}` - Get invoice by ID
- `POST /api/v1/invoices` - Create invoice (draft or issue)
- `PUT /api/v1/invoices/{invoiceId}` - Update invoice (draft only)
- `POST /api/v1/invoices/{invoiceId}/issue` - Issue draft invoice

### Payments
- `GET /api/v1/payments` - List payments
- `POST /api/v1/payments` - Record payment
- `POST /api/v1/payments/{paymentId}/cancel` - Cancel pending payment

### Finance Reports
- `GET /api/v1/finance/reports/accounts` - Generate accounts report (JSON)
- `GET /api/v1/finance/reports/export` - Export accounts report (PDF/CSV)

### Webhooks
- `POST /api/v1/webhooks/hdev` - Payment webhook from payment gateway

## Features

### Core Features
- ✅ JWT authentication with role-based access control
- ✅ User and student profile management
- ✅ Configurable fee structure
- ✅ Invoice generation and management (draft, issue, update)
- ✅ Invoice search by student number, name, and invoice number
- ✅ Payment processing (cash, bank transfer, mobile money, card, check)
- ✅ Payment cancellation (pending payments only)
- ✅ Automatic receipt generation for paid payments
- ✅ Payment webhook integration (HDEV/MoPay)
- ✅ Finance accounts report with aggregation by student
- ✅ Finance report export (PDF and CSV)
- ✅ Audit trail for all operations
- ✅ Outstanding balance tracking
- ✅ Service-oriented architecture with isolated controllers and models

## Database Schema

The system uses MySQL with the following main tables:
- `users` - User accounts
- `students` - Student information
- `fee_types` - Fee definitions
- `invoices` - Invoice records
- `invoice_items` - Invoice line items
- `payments` - Payment records
- `receipts` - Receipt records
- `audit_logs` - Audit trail

## User Roles

- **student** - View own invoices and payments
- **finance** - Manage invoices, process payments, generate reports
- **admin** - Full system access including user management

## Development

### Prerequisites
- Go 1.21+
- MySQL 8.0+
- Git

### Available Commands

```bash
make help          # Show available commands
make build         # Build the application (creates bin/server)
make run           # Run the application (uses go run)
make clean         # Clean build artifacts
make deps          # Download Go dependencies
```

### Database Setup

```bash
# Create database manually
mysql -u root -p -e "CREATE DATABASE university_invoices;"

# Run migrations
mysql -u root -p university_invoices < migrations/schema.sql

# Or use the setup script
./setup.sh
```

### Server Configuration

Edit `.env` file to configure:
- Database connection
- JWT secret
- Server port (default: 9107)

### Example Environment Variables

```
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=university_invoices
JWT_SECRET=your-secret-key
SERVER_PORT=9107
```

## Architecture

### Service-Oriented Design

The backend follows a service-oriented architecture where each service is self-contained:

- **User Service**: Handles user authentication and management
- **Student Service**: Manages student records
- **Invoice Service**: Handles invoice creation, search, and management
- **Payment Service**: Processes payments, cancellations, and receipt generation
- **Finance Service**: Generates financial reports and exports
- **Fee Service**: Manages fee structure
- **Admin Service**: Administrative operations
- **Webhook Service**: Handles payment gateway notifications

Each service has:
- Controllers for HTTP handling and business logic
- Models for data structures and validation

### Shared Components

- **Repositories**: Database operations shared across services (invoice, payment, student, user, fee_type, receipt, audit)
- **Middleware**: CORS and JWT authentication middleware
- **Routes**: Centralized route registration
- **Config**: Configuration management

## Security Considerations

- Passwords are hashed using bcrypt
- SQL injection prevention through parameterized queries
- JWT-based authentication for API access
- Role-based access control (student, finance, admin)
- Audit trail for all financial operations
- CORS enabled for cross-origin requests

## Database Schema

The system uses MySQL with the following main tables:
- `users` - User accounts and authentication
- `students` - Student information linked to users
- `fee_types` - Fee definitions and structures
- `invoices` - Invoice records with status tracking
- `invoice_items` - Invoice line items
- `payments` - Payment records with status tracking
- `receipts` - Receipt records for completed payments
- `audit_logs` - Audit trail for all operations

## Invoice Status Flow

Invoices follow this status progression:
1. **Draft** - Initial state, can be modified
2. **Issued** - Sent to student, no longer editable
3. **Partially Paid** - Partial payment received
4. **Paid** - Fully paid
5. **Overdue** - Past due date without full payment

## Payment Status Flow

Payments follow this status progression:
1. **Pending** - Payment initiated but not confirmed
2. **Paid** - Payment successfully completed
3. **Failed** - Payment processing failed
4. **Cancelled** - Payment cancelled (only if pending)

## Testing Credentials

For development testing:
- Finance user: `finance@uk.ac.rw` / `password123`
- Student user: `student@uk.ac.rw` / `password123`
- Student number: `UOK/2024/0001`

## License

This project is for University of Kigali internal use.

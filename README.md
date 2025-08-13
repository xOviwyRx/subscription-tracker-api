# Subscription Aggregation Service

A REST API service for aggregating data about user online subscriptions, built with Go and PostgreSQL.

## 📋 Project Overview

This service provides functionality to manage and aggregate user subscription data, allowing users to track their online subscriptions with cost analysis and filtering capabilities.

## 🚀 Features

- **CRUD Operations**: Complete subscription management (Create, Read, Update, Delete)
- **Cost Aggregation**: Calculate total subscription costs for selected periods with filtering
- **User Management**: Support for multiple users with UUID identification
- **RESTful API**: Clean REST endpoints with proper HTTP methods
- **Swagger Documentation**: Interactive API documentation
- **Docker Support**: Easy deployment with Docker Compose
- **Database Migrations**: Automated database schema management
- **Comprehensive Testing**: Unit tests for handlers and service layers with mocking

## 🏗️ Architecture

This project follows **Clean Architecture** principles with dependency inversion for maintainability and testability:

```
🌐 Presentation Layer (HTTP Handlers, Validation, Middleware)
           ↓ depends on interfaces
🧠 Business Layer (Services, Business Logic, Transaction Management)
           ↓ depends on interfaces  
🗄️ Data Layer (Repository Pattern, CRUD Operations)
           ↓ depends on concrete implementations
🔌 External Layer (PostgreSQL, Logger, Configuration)
```

**Key Benefits:**
- **🧪 Testable**: Easy to mock dependencies for unit testing
- **🔧 Maintainable**: Clear separation of concerns across layers
- **📦 Modular**: Each layer has single responsibility
- **🔄 Flexible**: Can swap implementations without code changes

**Flow Example**: `HTTP Request → Handler → Service (business logic) → Repository → Database`

> 📋 **[Detailed Architecture Documentation](https://xoviwyrx.github.io/subscription-tracker-api/docs/architecture.html)** - Interactive visual guide

## 🧪 Testing

The project includes comprehensive unit tests for critical layers:

### Handler Tests
- **HTTP Request/Response Testing**: Validates API endpoints behavior
- **Input Validation**: Tests for malformed JSON, invalid parameters
- **Error Handling**: Proper HTTP status codes and error messages
- **Mock Service Integration**: Isolated testing using service mocks

**Test Coverage:**
- `CreateSubscription` - Success and validation scenarios
- `GetSubscription` - Success, not found, and invalid ID cases
- `DeleteSubscription` - Successful deletion
- Error status code mapping

### Service Tests
- **Business Logic Validation**: Input validation and business rules
- **Transaction Execution**: Database transaction execution
- **Cost Calculation**: Subscription cost aggregation logic
- **Date Validation**: MM-YYYY format validation and date range checks
- **Mock Repository Integration**: Isolated testing with repository mocks

**Test Coverage:**
- `CreateSubscription` - Success, validation errors, duplicate checks
- `UpdateSubscription` - Success and not found scenarios
- `DeleteSubscription` - Success scenario
- `CalculateTotalCost` - Cost aggregation and validation
- Date utility functions (`isValidDate`, `calculateMonthsBetween`)

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/handlers
go test ./internal/service
```

**Testing Stack:**
- **Framework**: Go's built-in `testing` package
- **Assertions**: `github.com/stretchr/testify/assert`
- **Mocking**: `github.com/stretchr/testify/mock`
- **HTTP Testing**: `net/http/httptest`
- **Test Database**: SQLite in-memory for service tests

## 📊 Data Model

Each subscription record contains:
- **Service Name**: Name of the subscription service
- **Cost**: Monthly cost (integer)
- **User ID**: UUID format user identifier
- **Start Date**: Subscription start date (month and year)
- **End Date**: Optional subscription end date

### Example Subscription Record
```json
{
  "service_name": "Netflix",
  "price": 990,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025"
}
```

## 🛠 Technology Stack

- **Backend**: Go 1.24+
- **Database**: PostgreSQL
- **ORM**: GORM with repository pattern
- **Router**: Gin HTTP framework
- **Documentation**: Swagger/OpenAPI
- **Logging**: Logrus structured logging
- **Containerization**: Docker & Docker Compose
- **Configuration**: YAML configuration files
- **Testing**: Go testing, Testify, SQLite (in-memory)

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose

### Running with Docker Compose

1. **Clone the repository**
```bash
git clone <repository-url>
cd subscription_tracker_api
```

2. **Start the services**
```bash
docker-compose up -d
```

3. **Access the application**
- API: http://localhost:8080
- Swagger UI: http://localhost:8080/swagger/index.html

## 📚 API Endpoints

### Subscriptions

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/subscriptions` | Get all subscriptions with optional filtering |
| `POST` | `/api/v1/subscriptions` | Create a new subscription |
| `GET` | `/api/v1/subscriptions/{id}` | Get subscription by ID |
| `PUT` | `/api/v1/subscriptions/{id}` | Update subscription |
| `DELETE` | `/api/v1/subscriptions/{id}` | Delete subscription |

### Aggregation

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/subscriptions/calculate-cost` | Get total cost for period with filtering |

### Query Parameters for Filtering

- `user_id`: Filter by user UUID
- `start_date`: Filter by start date (MM-YYYY format)
- `end_date`: Filter by end date (MM-YYYY format)
- `service_name`: Filter by service name

### Example API Calls

**Create Subscription:**
```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Netflix",
    "price": 990,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

**Get Total Cost:**
```bash
curl "http://localhost:8080/api/v1/subscriptions/calculate-cost?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&start_date=01-2025&end_date=12-2025"
```
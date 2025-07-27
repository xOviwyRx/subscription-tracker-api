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
- **Documentation**: Swagger/OpenAPI
- **Containerization**: Docker & Docker Compose
- **Configuration**: YAML configuration files

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
- Swagger UI: http://localhost:8080/swagger/

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
| `GET` | `/api/v1/subscriptions/total-cost` | Get total cost for period with filtering |

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
curl "http://localhost:8080/api/v1/subscriptions/total-cost?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&start_date=01-2025&end_date=12-2025"
```

## 🗄 Database Schema

```sql
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL,
    user_id UUID NOT NULL,
    start_date VARCHAR(7) NOT NULL, -- MM-YYYY format
    end_date VARCHAR(7),            -- MM-YYYY format, optional
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_dates ON subscriptions(start_date, end_date);
```

## ⚙️ Configuration

The application uses YAML configuration file:

```yaml
# config.yaml
server:
  host: "localhost"
  port: "8080"

database:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "password"
  dbname: "subscription_tracker"
  sslmode: "disable"
```

## 📖 API Documentation

Interactive Swagger documentation is available at `/swagger/` when the service is running.

To regenerate Swagger docs:
```bash
swag init -g cmd/server/main.go -o ./cmd/server/docs
```
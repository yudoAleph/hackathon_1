# Contact Management API

A RESTful API for contact management built with Go, Gin, and MySQL.

## Features

- Health check endpoint
- User registration and authentication with JWT
- Contact CRUD operations
- Favorite contacts
- Search and pagination
- MySQL database with proper indexing
- Docker support
- Gin framework for high-performance routing

## Prerequisites

- Go 1.24+
- MySQL 8.0+
- Docker (optional)

## Server Configuration

The server runs on **port 9001** by default.

## Project Structure

```
hackathon_1/
├── cmd/
│   └── server/
│       └── main.go          # Main entry point
├── internal/
│   └── app/
│       ├── handlers/        # HTTP handlers
│       ├── models/          # Data models
│       ├── repository/      # Database layer
│       └── routes/          # Route definitions
├── pkg/
│   └── config/             # Configuration
├── configs/                # Config files
└── docs/                   # Documentation
```

## Database Setup

### Using Migrations (Recommended)

The project now uses a proper migration system for database schema management:

```bash
# Apply all migrations
make migrate-up

# Check migration status
make migrate-status

# Rollback if needed
make migrate-down
```

See [`MIGRATIONS.md`](MIGRATIONS.md) for detailed migration documentation.

### Legacy Setup (setup_db.sh - No Longer Used)

The `setup_db.sh` script is **deprecated** and should not be used. It has been replaced by the migration system above.

## Environment Configuration

Copy `.env.example` to `.env` and update the values:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=your_mysql_user
DB_PASSWORD=your_mysql_password
DB_NAME=getcontact

# JWT Configuration
JWT_SECRET=your_jwt_secret_key

# Server Configuration
PORT=8080
ENVIRONMENT=development
```

## Installation & Running

1. Install dependencies:

   ```bash
   go mod tidy
   ```

2. Run the application:

   ```bash
   make run
   ```

   **Note:** The `make run` command now automatically:
   - Checks if the database and tables exist
   - Runs migrations if needed
   - Starts the application

   You can also run manually:
   ```bash
   go run ./cmd/server/main.go
   ```

3. The API will be available at `http://localhost:9001`

### Health Check

You can verify the server is running by accessing the health check endpoint:

```bash
curl http://localhost:9001/health
```

Response:
```json
{
  "status": "healthy",
  "service": "contact-management-api",
  "version": "1.0.0"
}
```

## API Endpoints

### Health & System

- `GET /health` - Health check endpoint
- `GET /api/v1/ping` - Ping endpoint

### Authentication

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login

### Contacts (Protected routes)

- `GET /api/v1/contacts?q=&page=1&limit=20` - List contacts with search/pagination
- `POST /api/v1/contacts` - Create new contact
- `GET /api/v1/contacts/{id}` - Get contact details
- `PUT /api/v1/contacts/{id}` - Update contact
- `DELETE /api/v1/contacts/{id}` - Delete contact

### User Profile

- `GET /api/v1/me` - Get user profile
- `PUT /api/v1/me` - Update user profile

## Database Schema

### Users Table

- `id` (Primary Key, Auto Increment)
- `full_name` (Indexed)
- `email` (Unique, Indexed)
- `phone` (Indexed)
- `password` (Hashed)
- `avatar_url`
- `created_at` (Indexed)
- `updated_at`

### Contacts Table

- `id` (Primary Key, Auto Increment)
- `user_id` (Foreign Key to users.id, Indexed)
- `full_name` (Indexed)
- `phone` (Indexed)
- `email` (Indexed)
- `favorite` (Indexed)
- `created_at` (Indexed)
- `updated_at`

### Indexes

- Single column indexes on frequently queried fields
- Composite indexes for user-specific queries
- Foreign key constraints with CASCADE delete

## Docker Support

Build and run with Docker:

```bash
docker-compose up --build
```

## API Documentation

Complete API documentation with cURL examples is available in [`api_examples.md`](api_examples.md).

### Quick API Test

Run the included test script to verify all endpoints:

```bash
./test_api.sh
# or test against a remote server
./test_api.sh http://your-server-ip
```

This script will:

- Test the health endpoint
- Register a test user
- Login and get JWT token
- Test all protected endpoints (profile, contacts CRUD)
- Validate response codes and data

**Note:** Make sure the server is running and migrations are applied before running the test script.

## VPS Deployment

### Quick Setup

Deploy to production VPS with nginx reverse proxy:

```bash
# 1. Setup nginx reverse proxy (HTTP + HTTPS)
sudo ./setup_nginx.sh yourdomain.com

# 2. Run application
make run
```

Your API will be accessible at:
- **HTTP:** http://yourdomain.com
- **HTTPS:** https://yourdomain.com (after SSL setup)

### Troubleshooting

If you encounter issues on VPS:

```bash
sudo ./troubleshoot_vps.sh
```

This diagnostic script checks:
- System resources and connectivity
- Port status (9001, 3306, 80, 443)
- Firewall configuration
- Application and database status
- Nginx configuration
- SSL certificates
- Recent errors in logs

See [VPS Deployment Guide](docs/VPS_DEPLOYMENT.md) for detailed instructions.

## Project Structure

```
├── cmd/server/          # Application entry point
├── configs/             # Configuration management
├── internal/
│   ├── app/
│   │   ├── handlers/    # HTTP handlers
│   │   ├── models/      # Data models
│   │   ├── repository/  # Database layer
│   │   ├── routes/      # Route definitions
│   │   └── service/     # Business logic
│   ├── logger/          # Logging middleware
│   └── middleware/      # HTTP middleware
├── pkg/
│   ├── db/             # Database utilities
│   └── redis/          # Redis client
├── database_schema.sql # MySQL schema
├── setup_db.sh        # Database setup script
└── docker-compose.yml # Docker configuration
```

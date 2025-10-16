# ===============================
# Application Commands
# ===============================

# Check and migrate database if needed
# This command automatically checks if the database and tables exist
# If not, it will run migrations before starting the application
check-db:
	@echo "🔍 Checking database and tables..."
	@mysql -h localhost -u yudo -pyudo123 -e "USE hackathon_getcontact; SHOW TABLES;" > /dev/null 2>&1 || \
	(echo "⚠️  Database atau tabel belum ada, menjalankan migrasi..." && $(MAKE) migrate-up)
	@echo "✅ Database check complete"

# Run app locally (with automatic database check and migration)
# This will automatically check database and run migrations if needed
run: check-db
	@echo "🚀 Starting application..."
	go run ./cmd/server/main.go

# Build the application
build:
	go build -o bin/server ./cmd/server/main.go

# Run the built binary
start:
	./bin/server

# Kill process on port 9001
kill-9001:
	lsof -ti:9001 | xargs kill -9 || true

# Kill process on port 8080 (legacy)
kill-8080:
	lsof -ti:8080 | xargs kill -9 || true

# Database migration commands
migrate-up:
	go run ./cmd/migrate/main.go -command=up

migrate-down:
	go run ./cmd/migrate/main.go -command=down

migrate-status:
	go run ./cmd/migrate/main.go -command=status

# Run tests with coverage
test:
	go test ./... -coverprofile=tmp/coverage.out && go tool cover -func=tmp/coverage.out

# Run tests with detailed coverage for internal/app packages
test-coverage:
	go test -coverpkg=./internal/app/... ./internal/app -coverprofile=tmp/coverage.out && go tool cover -func=tmp/coverage.out

# Generate HTML coverage report
test-coverage-html:
	go test -coverpkg=./internal/app/... ./internal/app -coverprofile=tmp/coverage.out && go tool cover -html=tmp/coverage.out -o tmp/coverage.html

# Generate Swagger docs
swag:
	swag init -g cmd/server/main.go -o docs

# ===============================
# Docker Commands
# ===============================

# Build the Docker image
docker-build:
	docker build -t user-service .

# Run containers with build (foreground mode)
docker-up:
	docker-compose up --build

# Start containers in background
docker-start:
	docker-compose up -d

# Reload containers (rebuild and restart)
docker-reload:
	docker-compose down && docker-compose up --build -d

# Stop running containers (keep data)
docker-stop:
	docker-compose stop

# Stop and remove containers (cleanup)
docker-down:
	docker-compose down

# ===============================
# Database Commands
# ===============================

# Clean SQLite DB file (if persisted to host)
clean-db:
	rm -f test.db

# Format code
fmt:
	go fmt ./...

# Tidy modules
tidy:
	go mod tidy

# Run static analysis (gosec and golangci-lint)
lint:
	golangci-lint run ./...

gosec:
	gosec ./...

# Clean build artifacts and temporary files
clean:
	rm -rf bin/* tmp/coverage.* tmp/*.out tmp/*.db

# Run full CI checks: linting, security, swagger, formatting
ci: tidy fmt swag lint gosec test
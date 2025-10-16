# Automatic Database Migration

## Overview

The project now includes automatic database checking and migration when using `make run`. This ensures that the database and all required tables are properly set up before starting the application.

## How It Works

When you run `make run`, the following steps occur automatically:

1. **Database Check**: The system checks if the database `hackathon_getcontact` exists and has tables
2. **Auto Migration**: If the database or tables are missing, migrations are automatically executed
3. **Application Start**: Once the database is ready, the application starts normally

## Commands

### Run Application (with Auto-Check)

```bash
make run
```

This command will:
- ✅ Check database existence
- ✅ Run migrations if needed
- ✅ Start the application

### Manual Database Check

```bash
make check-db
```

This will check the database and run migrations if needed, without starting the application.

### Manual Migration Commands

```bash
# Apply all migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check migration status
make migrate-status
```

## Example Output

### First Run (No Database)

```bash
$ make run
🔍 Checking database and tables...
⚠️  Database atau tabel belum ada, menjalankan migrasi...
📦 Running database migrations...
Applying migration: 001_create_users_table
Successfully applied migration: 001_create_users_table
Applying migration: 002_create_contacts_table
Successfully applied migration: 002_create_contacts_table
Applying migration: 003_fix_schema_migrations_table
Successfully applied migration: 003_fix_schema_migrations_table
✅ All migrations completed successfully!
✅ Database check complete
🚀 Starting application...
{"timestamp":"2025-10-16T14:47:45.263189+07:00","level":"INFO","msg":"Starting Contact Management API"}
Server starting on port 9001...
```

### Subsequent Runs (Database Exists)

```bash
$ make run
🔍 Checking database and tables...
✅ Database check complete
🚀 Starting application...
{"timestamp":"2025-10-16T14:47:45.263189+07:00","level":"INFO","msg":"Starting Contact Management API"}
Server starting on port 9001...
```

## Benefits

1. **Zero Manual Setup**: No need to run migrations manually before first start
2. **Idempotent**: Safe to run multiple times - migrations that are already applied will be skipped
3. **Error Prevention**: Prevents runtime errors due to missing database schema
4. **Developer Friendly**: New developers can just run `make run` without database setup knowledge

## Technical Details

### Migration File

Location: `cmd/migrate/main.go`

This file:
- Loads environment variables from `configs/.env`
- Connects to MySQL database
- Executes migrations defined in `internal/app/migrations/migrations.go`
- Tracks applied migrations in `schema_migrations` table

### Check Logic

The `check-db` target in Makefile uses MySQL client to verify database existence:

```makefile
check-db:
	@mysql -h localhost -u yudo -pyudo123 -e "USE hackathon_getcontact; SHOW TABLES;" > /dev/null 2>&1 || \
	(echo "⚠️  Database atau tabel belum ada, menjalankan migrasi..." && $(MAKE) migrate-up)
```

If the command fails (database doesn't exist or has no tables), it automatically runs `migrate-up`.

## Troubleshooting

### Migration Fails

If migration fails:

1. Check database credentials in `configs/.env`:
   ```env
   DB_HOST=localhost
   DB_PORT=3306
   DB_USER=yudo
   DB_PASSWORD=yudo123
   DB_NAME=hackathon_getcontact
   ```

2. Ensure MySQL is running:
   ```bash
   mysql -h localhost -u yudo -pyudo123 -e "SELECT 1"
   ```

3. Check migration files in `internal/app/migrations/migrations.go` for syntax errors

### Database Already Exists but Check Fails

If you get false positives on the check:

1. Run manual check:
   ```bash
   mysql -h localhost -u yudo -pyudo123 -e "USE hackathon_getcontact; SHOW TABLES;"
   ```

2. If tables exist, you can skip check and run directly:
   ```bash
   go run ./cmd/server/main.go
   ```

### Force Re-run Migrations

If you need to reset the database:

```bash
# Drop database
mysql -h localhost -u yudo -pyudo123 -e "DROP DATABASE IF EXISTS hackathon_getcontact; CREATE DATABASE hackathon_getcontact;"

# Re-run migrations
make migrate-up
```

## Best Practices

1. **Always use `make run`** for development to ensure database is up-to-date
2. **Review migrations** before applying in production
3. **Backup database** before running migrations in production
4. **Version control migrations** - all migration files should be committed to git
5. **Test migrations** on a separate database before production deployment

## See Also

- [Migration Documentation](MIGRATIONS.md) - Detailed migration system documentation
- [Database Schema](../README.md#database-schema) - Database schema documentation
- [Environment Configuration](../README.md#environment-configuration) - .env setup guide

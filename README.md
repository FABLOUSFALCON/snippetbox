# Snippetbox

A web application for creating and sharing text snippets, built with Go and PostgreSQL.

## Tech Stack

- **Language**: Go 1.25+
- **Database**: PostgreSQL (using pgx driver for performance)
- **Session Store**: PostgreSQL-backed sessions
- **Frontend**: HTML templates with partials
- **Security**: HTTPS/TLS support, CSRF protection, bcrypt password hashing

## Project Structure

```
snippetbox/
├── cmd/web/              # Application entry point
│   ├── main.go          # Server setup, DB connection, routing
│   ├── handlers.go      # HTTP handlers (home, create snippet, user auth)
│   ├── middleware.go    # Security, logging, session middleware
│   ├── routes.go        # Route definitions
│   └── templates.go     # Template rendering
├── internal/
│   ├── models/          # Database models
│   │   ├── snippets.go  # Snippet CRUD operations
│   │   ├── users.go     # User authentication
│   │   └── testdata/    # Test SQL files
│   └── validator/       # Form validation
├── ui/                  # Frontend assets
│   ├── html/            # Templates (base, pages, partials)
│   └── static/          # CSS, JS, images
├── schema.sql           # Production database schema
└── setup_db.sh          # One-command database setup
```

## Database Setup

### What Each SQL File Does:

1. **`schema.sql`** - Production database schema
   - Creates `snippets` and `users` tables
   - Creates `sessions` table (for login sessions)
   - Adds indexes for performance
   - Inserts sample snippet data
   - **When to run**: Once after setting up your database
   - **How it runs**: Manually via `psql` or through `setup_db.sh`

2. **`setup_db.sh`** - Automated setup script
   - Creates PostgreSQL databases (`snippetbox` and `test_snippetbox`)
   - Creates users (`web` and `test_web`)
   - Sets up permissions
   - Runs `schema.sql` automatically
   - **When to run**: First time setup or reset

3. **`internal/models/testdata/setup.sql`** - Test database schema
   - Used by Go tests automatically
   - Creates tables for each test run
   - Inserts test data

4. **`internal/models/testdata/teardown.sql`** - Test cleanup
   - Drops tables after tests
   - Runs automatically via `t.Cleanup()`

### Quick Setup (First Time):

```bash
# 1. Make sure PostgreSQL is running (Docker or local)
docker run -d --name postgres18 -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:18

# 2. Run the setup script (creates everything)
./setup_db.sh

# 3. Done! Database is ready with tables and sample data
```

### Manual Setup (if you want to understand):

```bash
# 1. Create database and user
psql -U postgres -c "CREATE DATABASE snippetbox;"
psql -U postgres -c "CREATE USER web WITH PASSWORD 'pass';"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE snippetbox TO web;"

# 2. Run schema to create tables
psql -U web -d snippetbox -f schema.sql

# 3. Verify
psql -U web -d snippetbox -c "\dt"  # Should show snippets, users, sessions tables
```

## Running the Application

### Local Development:

```bash
# Build
go build -o web ./cmd/web

# Run (uses default DSN, no flags needed!)
./web

# With debug mode
./web -debug

# With TLS (local HTTPS)
./web -tls
```

### Configuration Options:

The app uses **both flags and environment variables** (this is good practice!):

**Priority order**: 
1. Command-line flags (highest priority)
2. Environment variables
3. Default values (lowest priority)

**Available flags:**
```bash
./web -help

  -addr string
        HTTP network address (default ":4001")
  -dsn string
        PostgreSQL data source name (default "postgres://web:pass@localhost:5432/snippetbox?sslmode=disable")
  -debug
        Enable debug mode (default false)
  -tls
        Enable TLS (default false - Render/cloud handles this)
  -cert string
        TLS certificate file path (default "./tls/localhost+1.pem")
  -key string
        TLS key file path (default "./tls/localhost+1-key.pem")
```

**Environment variables:**
```bash
# Override DSN via environment
export DATABASE_URL="postgres://user:pass@host:5432/dbname"
./web

# Override port (useful for cloud platforms like Render)
export PORT=8080
./web -addr=:$PORT
```

**Why flags + env is good:**
- ✅ **Local dev**: Use defaults, no typing needed
- ✅ **Cloud deploy**: Set `DATABASE_URL` env var, app picks it up
- ✅ **Testing**: Override specific values via flags
- ✅ **Security**: Secrets in env vars, not in code
- ✅ **Flexibility**: Can override anything when needed

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test -v ./internal/models/

# Tests automatically:
# - Create test database
# - Run setup.sql
# - Execute tests
# - Run teardown.sql
# - Clean up
```

## Deployment (Render/Cloud)

```bash
# Build command:
go build -o bin/web ./cmd/web

# Start command:
./bin/web -addr=:$PORT

# Environment variables (set in platform dashboard):
DATABASE_URL=your_postgres_connection_string
```

The app automatically:
- Uses `DATABASE_URL` if set
- Runs without TLS (cloud platforms provide HTTPS)
- Handles sessions in PostgreSQL

## Key Features

### Security:
- Bcrypt password hashing
- CSRF protection on all POST requests
- Secure session cookies
- SQL injection protection (parameterized queries)
- TLS support for local development

### Performance:
- pgx driver (fastest PostgreSQL driver for Go)
- Connection pooling via pgxpool
- Database indexes on frequently queried columns
- Template caching

### User Features:
- User signup/login/logout
- Password change functionality
- Create and view snippets
- Snippets auto-expire after set duration
- Session-based authentication

## Development Workflow

```bash
# 1. Make changes to code
vim cmd/web/handlers.go

# 2. Run tests
go test ./...

# 3. Run locally
./web -debug

# 4. Open browser
open http://localhost:4001

# 5. Check logs in terminal
```

## Database Schema Overview

**snippets table:**
- Stores text snippets with title, content, expiry
- Auto-expires after set duration
- Indexed on creation date

**users table:**
- Stores user accounts with bcrypt hashed passwords
- Unique email constraint
- Tracks account creation date

**sessions table:**
- Stores login sessions
- Auto-cleanup of expired sessions
- PostgreSQL-backed (persists across restarts)

## Common Tasks

**View database:**
```bash
psql -U web -d snippetbox
\dt              # List tables
SELECT * FROM snippets;
SELECT * FROM users;
```

**Reset database:**
```bash
./setup_db.sh    # Drops and recreates everything
```

**Add sample data:**
```bash
psql -U web -d snippetbox -f schema.sql
```

**Check what's running:**
```bash
lsof -i :4001    # Check if app is running
docker ps        # Check if PostgreSQL is running
```

## Learn More

This project follows patterns from "Let's Go" by Alex Edwards. Key concepts:

- **main.go**: Application bootstrap, dependency injection
- **handlers.go**: HTTP request/response handling
- **middleware.go**: Cross-cutting concerns (logging, auth, security)
- **models/**: Data layer abstraction
- **templates**: Server-side rendering with Go templates

Each file is commented to explain what it does and why!
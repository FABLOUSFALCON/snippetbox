#!/bin/bash
# Quick PostgreSQL Setup for Snippetbox

set -e

echo "🚀 Setting up PostgreSQL database for Snippetbox..."

# Database connection details
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"

echo "📊 Creating database and users..."

# Create databases and users
PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER <<-EOSQL
    -- Create production database
    CREATE DATABASE snippetbox;
    
    -- Create production user
    CREATE USER web WITH PASSWORD 'pass';
    
    -- Grant privileges
    GRANT CONNECT ON DATABASE snippetbox TO web;
    
    -- Create test database
    CREATE DATABASE test_snippetbox;
    
    -- Create test user
    CREATE USER test_web WITH PASSWORD 'pass';
    
    -- Grant privileges on test database
    GRANT CONNECT ON DATABASE test_snippetbox TO test_web;
EOSQL

echo "📝 Setting up snippetbox schema..."

# Setup production database schema
PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d snippetbox <<-EOSQL
    -- Grant privileges on schema
    GRANT ALL ON SCHEMA public TO web;
    GRANT ALL ON ALL TABLES IN SCHEMA public TO web;
    GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO web;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO web;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO web;
EOSQL

# Run schema as 'web' user
PGPASSWORD=pass psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U web -d snippetbox -f schema.sql

echo "🧪 Setting up test database..."

# Setup test database
PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d test_snippetbox <<-EOSQL
    -- Grant privileges on schema
    GRANT ALL ON SCHEMA public TO test_web;
    GRANT ALL ON ALL TABLES IN SCHEMA public TO test_web;
    GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO test_web;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO test_web;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO test_web;
EOSQL

echo "✅ Setup complete!"
echo ""
echo "Connection strings:"
echo "  Production: postgres://web:pass@localhost:5432/snippetbox?sslmode=disable"
echo "  Test: postgres://test_web:pass@localhost:5432/test_snippetbox?sslmode=disable"
echo ""
echo "Run the app:"
echo "  ./web -dsn='postgres://web:pass@localhost:5432/snippetbox?sslmode=disable'"

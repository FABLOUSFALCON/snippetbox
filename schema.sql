-- PostgreSQL Schema for Snippetbox Application
-- Converted from MySQL for maximum performance with pgx driver

-- Create snippets table (matches original MySQL schema)
CREATE TABLE IF NOT EXISTS snippets (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    created TIMESTAMP NOT NULL,
    expires TIMESTAMP NOT NULL
);

-- Add index on created column for better query performance
CREATE INDEX IF NOT EXISTS idx_snippets_created ON snippets(created);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    hashed_password CHAR(60) NOT NULL,
    created TIMESTAMP NOT NULL
);

-- Add unique constraint on email
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_uc_email') THEN
        ALTER TABLE users ADD CONSTRAINT users_uc_email UNIQUE (email);
    END IF;
END $$;

-- Create sessions table for scs/postgresstore
CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMP(6) NOT NULL
);

-- Create index on expiry for session cleanup
CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions(expiry);

-- Add some dummy records (matching the original MySQL data)
INSERT INTO snippets (title, content, created, expires) VALUES (
    'An old silent pond',
    'An old silent pond...' || E'\n' || 'A frog jumps into the pond,' || E'\n' || 'splash! Silence again.' || E'\n\n' || '– Matsuo Bashō',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP + INTERVAL '365 days'
)
ON CONFLICT DO NOTHING;

INSERT INTO snippets (title, content, created, expires) VALUES (
    'Over the wintry forest',
    'Over the wintry' || E'\n' || 'forest, winds howl in rage' || E'\n' || 'with no leaves to blow.' || E'\n\n' || '– Natsume Soseki',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP + INTERVAL '365 days'
)
ON CONFLICT DO NOTHING;

INSERT INTO snippets (title, content, created, expires) VALUES (
    'First autumn morning',
    'First autumn morning' || E'\n' || 'the mirror I stare into' || E'\n' || 'shows my father''s face.' || E'\n\n' || '– Murakami Kijo',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP + INTERVAL '7 days'
)
ON CONFLICT DO NOTHING;

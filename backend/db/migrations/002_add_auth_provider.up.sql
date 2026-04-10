-- Add auth_provider column to users table to support OAuth login
ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(20) NOT NULL DEFAULT 'email';

-- Allow OAuth users to have an empty password_hash
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

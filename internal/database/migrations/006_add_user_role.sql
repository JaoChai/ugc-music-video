-- Migration: 006_add_user_role
-- Description: Add role column to users table for admin access control

-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'user' NOT NULL;

-- Create index for role queries
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Add check constraint for valid roles
ALTER TABLE users ADD CONSTRAINT chk_users_role CHECK (role IN ('user', 'admin'));

-- +goose Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;
DROP INDEX IF EXISTS idx_users_role;
ALTER TABLE users DROP COLUMN IF EXISTS role;

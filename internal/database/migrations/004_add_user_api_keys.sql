-- +migrate Up
-- Add encrypted API key columns to users table
ALTER TABLE users
ADD COLUMN openrouter_api_key TEXT,
ADD COLUMN kie_api_key TEXT;

-- +migrate Down
ALTER TABLE users
DROP COLUMN IF EXISTS openrouter_api_key,
DROP COLUMN IF EXISTS kie_api_key;

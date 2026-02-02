-- Migration: 004_add_user_api_keys
-- Description: Add encrypted API key columns to users table

ALTER TABLE users
ADD COLUMN IF NOT EXISTS openrouter_api_key TEXT,
ADD COLUMN IF NOT EXISTS kie_api_key TEXT;

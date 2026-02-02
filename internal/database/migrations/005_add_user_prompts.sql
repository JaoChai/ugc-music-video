-- Migration: 005_add_user_prompts
-- Description: Add custom AI agent prompt columns to users table

ALTER TABLE users
ADD COLUMN IF NOT EXISTS song_concept_prompt TEXT,
ADD COLUMN IF NOT EXISTS song_selector_prompt TEXT,
ADD COLUMN IF NOT EXISTS image_concept_prompt TEXT;

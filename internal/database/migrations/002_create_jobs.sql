-- Migration: 002_create_jobs
-- Description: Create jobs table for tracking UGC generation workflow

CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'pending',
    concept TEXT NOT NULL,
    llm_model VARCHAR(100),
    song_prompt JSONB,
    suno_task_id VARCHAR(100),
    generated_songs JSONB,
    selected_song_id VARCHAR(100),
    image_prompt JSONB,
    nano_task_id VARCHAR(100),
    audio_url TEXT,
    image_url TEXT,
    video_url TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_jobs_user_id ON jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at DESC);

-- Trigger to auto-update updated_at
CREATE TRIGGER update_jobs_updated_at
    BEFORE UPDATE ON jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Migration: 003_create_assets
-- Description: Create assets table for storing generated media files

CREATE TABLE IF NOT EXISTS assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID REFERENCES jobs(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    asset_type VARCHAR(50) NOT NULL, -- 'audio', 'image', 'video'
    storage_provider VARCHAR(50) DEFAULT 'r2', -- 'r2', 's3', 'local'
    storage_key TEXT NOT NULL, -- path/key in storage
    original_url TEXT, -- original URL from provider (Suno, Nano, etc.)
    public_url TEXT, -- CDN/public accessible URL
    file_size BIGINT,
    mime_type VARCHAR(100),
    metadata JSONB, -- additional metadata (duration, dimensions, etc.)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_assets_job_id ON assets(job_id);
CREATE INDEX IF NOT EXISTS idx_assets_user_id ON assets(user_id);
CREATE INDEX IF NOT EXISTS idx_assets_asset_type ON assets(asset_type);
CREATE INDEX IF NOT EXISTS idx_assets_created_at ON assets(created_at DESC);

-- Trigger to auto-update updated_at
CREATE TRIGGER update_assets_updated_at
    BEFORE UPDATE ON assets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

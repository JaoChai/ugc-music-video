-- Add YouTube integration fields
ALTER TABLE users ADD COLUMN IF NOT EXISTS youtube_refresh_token TEXT;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS youtube_url TEXT;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS youtube_video_id VARCHAR(100);
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS youtube_error TEXT;

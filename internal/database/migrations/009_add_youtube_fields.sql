-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS youtube_refresh_token TEXT;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS youtube_url TEXT;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS youtube_video_id VARCHAR(100);
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS youtube_error TEXT;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS youtube_refresh_token;
ALTER TABLE jobs DROP COLUMN IF EXISTS youtube_url;
ALTER TABLE jobs DROP COLUMN IF EXISTS youtube_video_id;
ALTER TABLE jobs DROP COLUMN IF EXISTS youtube_error;

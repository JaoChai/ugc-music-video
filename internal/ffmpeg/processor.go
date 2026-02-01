// Package ffmpeg provides video processing capabilities using FFmpeg.
package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// Processor handles video processing operations using FFmpeg.
type Processor struct {
	logger *zap.Logger
}

// NewProcessor creates a new FFmpeg processor.
func NewProcessor(logger *zap.Logger) *Processor {
	return &Processor{
		logger: logger,
	}
}

// CreateMusicVideoInput contains the input parameters for creating a music video.
type CreateMusicVideoInput struct {
	AudioURL   string // URL of the audio file
	ImageURL   string // URL of the background image
	OutputPath string // Path where the output video will be saved
}

// CreateMusicVideoOutput contains the result of creating a music video.
type CreateMusicVideoOutput struct {
	OutputPath string        // Path to the generated video
	Duration   time.Duration // Duration of the video
	FileSize   int64         // Size of the video file in bytes
}

// CreateMusicVideo creates a music video by combining an audio file with a static image.
// It downloads the audio and image from URLs, then uses FFmpeg to create the video.
func (p *Processor) CreateMusicVideo(ctx context.Context, input CreateMusicVideoInput) (*CreateMusicVideoOutput, error) {
	p.logger.Info("starting music video creation",
		zap.String("audio_url", input.AudioURL),
		zap.String("image_url", input.ImageURL),
		zap.String("output_path", input.OutputPath),
	)

	// Create temp directory for intermediate files
	tempDir, err := os.MkdirTemp("", "ugc-video-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download audio file
	audioPath := filepath.Join(tempDir, "audio.mp3")
	if err := downloadFile(ctx, input.AudioURL, audioPath); err != nil {
		return nil, fmt.Errorf("failed to download audio: %w", err)
	}
	p.logger.Debug("downloaded audio file", zap.String("path", audioPath))

	// Download image file
	imagePath := filepath.Join(tempDir, "image.png")
	if err := downloadFile(ctx, input.ImageURL, imagePath); err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	p.logger.Debug("downloaded image file", zap.String("path", imagePath))

	// Ensure output directory exists
	outputDir := filepath.Dir(input.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create video using FFmpeg
	// Command: ffmpeg -loop 1 -i image.png -i audio.mp3 -c:v libx264 -tune stillimage -c:a aac -b:a 192k -pix_fmt yuv420p -shortest output.mp4
	args := []string{
		"-loop", "1",
		"-i", imagePath,
		"-i", audioPath,
		"-c:v", "libx264",
		"-tune", "stillimage",
		"-c:a", "aac",
		"-b:a", "192k",
		"-pix_fmt", "yuv420p",
		"-shortest",
		"-y", // Overwrite output file if exists
		input.OutputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	p.logger.Debug("executing ffmpeg command",
		zap.Strings("args", args),
	)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg command failed: %w", err)
	}

	// Get output file info
	fileInfo, err := os.Stat(input.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat output file: %w", err)
	}

	// Get video duration using ffprobe
	duration, err := p.getVideoDuration(ctx, input.OutputPath)
	if err != nil {
		p.logger.Warn("failed to get video duration, using 0", zap.Error(err))
		duration = 0
	}

	p.logger.Info("music video created successfully",
		zap.String("output_path", input.OutputPath),
		zap.Int64("file_size", fileInfo.Size()),
		zap.Duration("duration", duration),
	)

	return &CreateMusicVideoOutput{
		OutputPath: input.OutputPath,
		Duration:   duration,
		FileSize:   fileInfo.Size(),
	}, nil
}

// getVideoDuration uses ffprobe to get the duration of a video file.
func (p *Processor) getVideoDuration(ctx context.Context, videoPath string) (time.Duration, error) {
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	}

	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe command failed: %w", err)
	}

	var seconds float64
	if _, err := fmt.Sscanf(string(output), "%f", &seconds); err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

// downloadFile downloads a file from a URL to a local path.
func downloadFile(ctx context.Context, url, destPath string) error {
	// Use curl for downloading as it handles various edge cases well
	cmd := exec.CommandContext(ctx, "curl", "-L", "-o", destPath, "-s", "-f", url)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("curl download failed: %w", err)
	}

	// Verify file exists and has content
	fileInfo, err := os.Stat(destPath)
	if err != nil {
		return fmt.Errorf("downloaded file not found: %w", err)
	}
	if fileInfo.Size() == 0 {
		return fmt.Errorf("downloaded file is empty")
	}

	return nil
}

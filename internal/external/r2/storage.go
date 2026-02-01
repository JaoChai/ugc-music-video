// Package r2 provides a client for Cloudflare R2 storage using AWS SDK v2.
// R2 is S3-compatible, so we use the standard AWS S3 SDK.
package r2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Config holds the configuration for R2 storage client.
type Config struct {
	// AccountID is the Cloudflare account ID
	AccountID string

	// AccessKeyID is the R2 access key ID
	AccessKeyID string

	// SecretAccessKey is the R2 secret access key
	SecretAccessKey string

	// BucketName is the name of the R2 bucket
	BucketName string

	// PublicURL is the optional public URL for the bucket (e.g., custom domain or r2.dev URL)
	// If set, GetPublicURL will return URLs using this base URL
	PublicURL string
}

// Client is a Cloudflare R2 storage client.
type Client struct {
	s3Client   *s3.Client
	presigner  *s3.PresignClient
	bucketName string
	publicURL  string
}

// NewClient creates a new R2 storage client.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.AccountID == "" {
		return nil, errors.New("r2: AccountID is required")
	}
	if cfg.AccessKeyID == "" {
		return nil, errors.New("r2: AccessKeyID is required")
	}
	if cfg.SecretAccessKey == "" {
		return nil, errors.New("r2: SecretAccessKey is required")
	}
	if cfg.BucketName == "" {
		return nil, errors.New("r2: BucketName is required")
	}

	// R2 endpoint format: https://{account_id}.r2.cloudflarestorage.com
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	// Create credentials provider
	creds := credentials.NewStaticCredentialsProvider(
		cfg.AccessKeyID,
		cfg.SecretAccessKey,
		"", // session token is not used for R2
	)

	// Create S3 client with R2 configuration
	s3Client := s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Credentials:  creds,
		Region:       "auto", // R2 uses "auto" as region
	})

	// Normalize public URL (remove trailing slash if present)
	publicURL := strings.TrimSuffix(cfg.PublicURL, "/")

	return &Client{
		s3Client:   s3Client,
		presigner:  s3.NewPresignClient(s3Client),
		bucketName: cfg.BucketName,
		publicURL:  publicURL,
	}, nil
}

// Upload uploads a file to R2 storage.
func (c *Client) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(c.bucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	}

	_, err := c.s3Client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("r2: failed to upload object %q: %w", key, err)
	}

	return nil
}

// UploadFromURL downloads content from a URL and uploads it to R2 storage.
func (c *Client) UploadFromURL(ctx context.Context, key string, sourceURL string) error {
	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return fmt.Errorf("r2: failed to create request for %q: %w", sourceURL, err)
	}

	// Download the file
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("r2: failed to download from %q: %w", sourceURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("r2: unexpected status code %d when downloading from %q", resp.StatusCode, sourceURL)
	}

	// Determine content type from response or default to binary
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload to R2
	return c.Upload(ctx, key, resp.Body, contentType)
}

// GetPresignedURL generates a presigned URL for private access to an object.
func (c *Client) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}

	presignedReq, err := c.presigner.PresignGetObject(ctx, input, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("r2: failed to generate presigned URL for %q: %w", key, err)
	}

	return presignedReq.URL, nil
}

// GetPublicURL returns the public URL for an object.
// Returns an empty string if publicURL is not configured.
func (c *Client) GetPublicURL(key string) string {
	if c.publicURL == "" {
		return ""
	}

	// Ensure key doesn't have leading slash for proper URL construction
	key = strings.TrimPrefix(key, "/")

	return fmt.Sprintf("%s/%s", c.publicURL, key)
}

// Delete removes an object from R2 storage.
func (c *Client) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}

	_, err := c.s3Client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("r2: failed to delete object %q: %w", key, err)
	}

	return nil
}

// Exists checks if an object exists in R2 storage.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}

	_, err := c.s3Client.HeadObject(ctx, input)
	if err != nil {
		// Check for NotFound error using AWS SDK v2 error types
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}

		// Also check for NoSuchKey error
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return false, nil
		}

		// Fallback: check error message for common not found patterns
		// This handles cases where R2 might return errors differently than S3
		if isNotFoundError(err) {
			return false, nil
		}

		return false, fmt.Errorf("r2: failed to check if object %q exists: %w", key, err)
	}

	return true, nil
}

// isNotFoundError checks if the error indicates the object was not found.
// This is a fallback for error patterns not covered by AWS SDK error types.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "NotFound") ||
		strings.Contains(errStr, "NoSuchKey") ||
		strings.Contains(errStr, "404") ||
		strings.Contains(errStr, "not found")
}

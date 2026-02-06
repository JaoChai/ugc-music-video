package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/jaochai/ugc/internal/database"
	"github.com/jaochai/ugc/internal/models"
)

// ErrUserNotFound is returned when a user is not found in the database.
var ErrUserNotFound = errors.New("user not found")

// UserRepository defines the interface for user data access operations.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateAPIKeys(ctx context.Context, userID uuid.UUID, openRouterKey, kieKey *string) error
	GetAPIKeys(ctx context.Context, userID uuid.UUID) (openRouterKey, kieKey *string, err error)
	DeleteAPIKeys(ctx context.Context, userID uuid.UUID) error
	UpdateYouTubeToken(ctx context.Context, userID uuid.UUID, encryptedToken *string) error
	GetYouTubeToken(ctx context.Context, userID uuid.UUID) (*string, error)
}

// userRepository implements UserRepository using pgx.
type userRepository struct {
	db *database.DB
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(db *database.DB) UserRepository {
	return &userRepository{db: db}
}

// Create inserts a new user into the database.
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	// Set default role if not specified
	if user.Role == "" {
		user.Role = "user"
	}

	query := `
		INSERT INTO users (id, email, password_hash, name, openrouter_model, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`

	err := r.db.Pool().QueryRow(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.OpenRouterModel,
		user.Role,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by their ID.
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, name, role, openrouter_model, openrouter_api_key, kie_api_key, youtube_refresh_token, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	err := r.db.Pool().QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.Role,
		&user.OpenRouterModel,
		&user.OpenRouterAPIKey,
		&user.KIEAPIKey,
		&user.YouTubeRefreshToken,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by their email address.
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, name, role, openrouter_model, openrouter_api_key, kie_api_key, youtube_refresh_token, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	err := r.db.Pool().QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.Role,
		&user.OpenRouterModel,
		&user.OpenRouterAPIKey,
		&user.KIEAPIKey,
		&user.YouTubeRefreshToken,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// Update updates an existing user in the database.
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET email = $2, password_hash = $3, name = $4, openrouter_model = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	result, err := r.db.Pool().Exec(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.OpenRouterModel,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	// Fetch the updated_at timestamp
	err = r.db.Pool().QueryRow(
		ctx,
		`SELECT updated_at FROM users WHERE id = $1`,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to fetch updated_at: %w", err)
	}

	return nil
}

// Delete removes a user from the database by their ID.
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Pool().Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateAPIKeys updates the encrypted API keys for a user.
func (r *userRepository) UpdateAPIKeys(ctx context.Context, userID uuid.UUID, openRouterKey, kieKey *string) error {
	query := `
		UPDATE users
		SET openrouter_api_key = $2, kie_api_key = $3, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Pool().Exec(ctx, query, userID, openRouterKey, kieKey)
	if err != nil {
		return fmt.Errorf("failed to update API keys: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetAPIKeys retrieves the encrypted API keys for a user.
func (r *userRepository) GetAPIKeys(ctx context.Context, userID uuid.UUID) (openRouterKey, kieKey *string, err error) {
	query := `
		SELECT openrouter_api_key, kie_api_key
		FROM users
		WHERE id = $1
	`

	err = r.db.Pool().QueryRow(ctx, query, userID).Scan(&openRouterKey, &kieKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrUserNotFound
		}
		return nil, nil, fmt.Errorf("failed to get API keys: %w", err)
	}

	return openRouterKey, kieKey, nil
}

// DeleteAPIKeys removes the API keys for a user.
func (r *userRepository) DeleteAPIKeys(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET openrouter_api_key = NULL, kie_api_key = NULL, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Pool().Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete API keys: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateYouTubeToken updates the encrypted YouTube refresh token for a user.
// Pass nil to clear the token (disconnect).
func (r *userRepository) UpdateYouTubeToken(ctx context.Context, userID uuid.UUID, encryptedToken *string) error {
	query := `
		UPDATE users
		SET youtube_refresh_token = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Pool().Exec(ctx, query, userID, encryptedToken)
	if err != nil {
		return fmt.Errorf("failed to update YouTube token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetYouTubeToken retrieves the encrypted YouTube refresh token for a user.
func (r *userRepository) GetYouTubeToken(ctx context.Context, userID uuid.UUID) (*string, error) {
	query := `SELECT youtube_refresh_token FROM users WHERE id = $1`

	var token *string
	err := r.db.Pool().QueryRow(ctx, query, userID).Scan(&token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get YouTube token: %w", err)
	}

	return token, nil
}


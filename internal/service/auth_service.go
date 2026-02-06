package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/jaochai/ugc/internal/models"
	"github.com/jaochai/ugc/internal/repository"
)

// Auth service errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrUserNotFound       = errors.New("user not found")
)

// Claims represents the JWT claims
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, input models.CreateUserInput) (*models.User, error)
	Login(ctx context.Context, input models.LoginInput) (string, *models.User, error)
	ValidateToken(token string) (*Claims, error)
	RefreshToken(token string) (string, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GenerateShortToken(userID uuid.UUID, expiry time.Duration) (string, error)
	ValidateShortToken(tokenString string) (uuid.UUID, error)
}

// authService implements AuthService
type authService struct {
	userRepo  repository.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
	logger    *zap.Logger
}

// NewAuthService creates a new AuthService instance
func NewAuthService(
	userRepo repository.UserRepository,
	jwtSecret string,
	jwtExpiry time.Duration,
	logger *zap.Logger,
) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
		logger:    logger,
	}
}

// Register creates a new user account
func (s *authService) Register(ctx context.Context, input models.CreateUserInput) (*models.User, error) {
	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		s.logger.Error("failed to check existing email", zap.Error(err))
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		Name:         input.Name,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("user registered successfully", zap.String("email", user.Email), zap.String("user_id", user.ID.String()))

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *authService) Login(ctx context.Context, input models.LoginInput) (string, *models.User, error) {
	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		s.logger.Error("failed to get user by email", zap.Error(err))
		return "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		s.logger.Error("failed to generate token", zap.Error(err))
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	s.logger.Info("user logged in successfully", zap.String("email", user.Email), zap.String("user_id", user.ID.String()))

	return token, user, nil
}

// ValidateToken parses and validates a JWT token
func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken validates an existing token and generates a new one with fresh expiry
func (s *authService) RefreshToken(tokenString string) (string, error) {
	// Validate existing token
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		// Allow refresh for expired tokens (within reasonable time)
		if !errors.Is(err, ErrTokenExpired) {
			return "", err
		}
		// Re-parse to get claims even if expired
		claims = &Claims{}
		_, _ = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.jwtSecret), nil
		})
	}

	// Create new token with fresh expiry
	now := time.Now()
	newClaims := &Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   claims.UserID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	tokenString, err = token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	s.logger.Info("token refreshed successfully", zap.String("user_id", claims.UserID.String()))

	return tokenString, nil
}

// GetUserByID retrieves a user by their ID
func (s *authService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		s.logger.Error("failed to get user by id", zap.Error(err), zap.String("user_id", id.String()))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GenerateShortToken creates a short-lived JWT for OAuth state parameter (CSRF protection).
func (s *authService) GenerateShortToken(userID uuid.UUID, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ValidateShortToken validates a short-lived JWT and returns the user ID.
func (s *authService) ValidateShortToken(tokenString string) (uuid.UUID, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	if !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	return claims.UserID, nil
}

// generateToken creates a new JWT token for the given user
func (s *authService) generateToken(user *models.User) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

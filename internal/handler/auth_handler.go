// Package handler provides HTTP handlers for the API.
package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/jaochai/ugc/internal/middleware"
	"github.com/jaochai/ugc/internal/models"
	"github.com/jaochai/ugc/internal/repository"
	"github.com/jaochai/ugc/internal/service"
	"github.com/jaochai/ugc/pkg/response"
)

// LoginResponse represents the response for successful login
type LoginResponse struct {
	Token string              `json:"token"`
	User  models.UserResponse `json:"user"`
}

// RefreshResponse represents the response for token refresh
type RefreshResponse struct {
	Token string `json:"token"`
}

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService   service.AuthService
	userRepo      repository.UserRepository
	cryptoService service.CryptoService
	logger        *zap.Logger
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(
	authService service.AuthService,
	userRepo repository.UserRepository,
	cryptoService service.CryptoService,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		userRepo:      userRepo,
		cryptoService: cryptoService,
		logger:        logger,
	}
}

// RegisterRoutes registers all auth routes to the given router group
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)

		// Protected routes
		protected := auth.Group("")
		protected.Use(middleware.AuthMiddleware(h.authService, h.logger))
		{
			protected.GET("/me", h.Me)
			protected.GET("/api-keys", h.GetAPIKeysStatus)
			protected.PUT("/api-keys", h.UpdateAPIKeys)
			protected.DELETE("/api-keys", h.DeleteAPIKeys)
		}
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param input body models.CreateUserInput true "User registration data"
// @Success 201 {object} response.Response{data=models.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input models.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Debug("failed to bind registration input", zap.Error(err))
		response.BadRequest(c, "invalid request body")
		return
	}

	// Validate input
	if err := h.validateCreateUserInput(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Call service to register user
	user, err := h.authService.Register(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			response.BadRequest(c, "email already exists")
			return
		}
		h.logger.Error("failed to register user", zap.Error(err))
		response.Error(c, err)
		return
	}

	h.logger.Info("user registered successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
	)

	response.Created(c, user.ToResponse())
}

// Login handles user authentication
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param input body models.LoginInput true "User login credentials"
// @Success 200 {object} response.Response{data=LoginResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Debug("failed to bind login input", zap.Error(err))
		response.BadRequest(c, "invalid request body")
		return
	}

	// Validate input
	if err := h.validateLoginInput(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Call service to authenticate user
	token, user, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(c, "invalid email or password")
			return
		}
		h.logger.Error("failed to login user", zap.Error(err))
		response.Error(c, err)
		return
	}

	h.logger.Info("user logged in successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
	)

	response.Success(c, LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	})
}

// Refresh handles JWT token refresh
// @Summary Refresh JWT token
// @Description Refresh an existing JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} response.Response{data=RefreshResponse}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Unauthorized(c, "authorization header required")
		return
	}

	// Check Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		response.Unauthorized(c, "invalid authorization header format")
		return
	}

	tokenString := parts[1]

	// Call service to refresh token
	newToken, err := h.authService.RefreshToken(tokenString)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			response.Unauthorized(c, "invalid token")
			return
		}
		h.logger.Error("failed to refresh token", zap.Error(err))
		response.Error(c, err)
		return
	}

	h.logger.Debug("token refreshed successfully")

	response.Success(c, RefreshResponse{
		Token: newToken,
	})
}

// Me handles getting the current user's profile
// @Summary Get current user
// @Description Get the authenticated user's profile
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=models.UserResponse}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// Call service to get user
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		h.logger.Error("failed to get user", zap.Error(err), zap.String("user_id", userID.String()))
		response.Error(c, err)
		return
	}

	response.Success(c, user.ToResponse())
}

// validateCreateUserInput validates the user registration input
func (h *AuthHandler) validateCreateUserInput(input *models.CreateUserInput) error {
	if input.Email == "" {
		return errors.New("email is required")
	}

	// Basic email format validation
	if !strings.Contains(input.Email, "@") || !strings.Contains(input.Email, ".") {
		return errors.New("invalid email format")
	}

	if input.Password == "" {
		return errors.New("password is required")
	}

	if len(input.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	return nil
}

// validateLoginInput validates the login input
func (h *AuthHandler) validateLoginInput(input *models.LoginInput) error {
	if input.Email == "" {
		return errors.New("email is required")
	}

	if input.Password == "" {
		return errors.New("password is required")
	}

	return nil
}

// GetAPIKeysStatus returns the status of user's API keys (has/doesn't have)
// @Summary Get API keys status
// @Description Returns whether the user has configured API keys (not the actual keys)
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=models.APIKeysStatusResponse}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/api-keys [get]
func (h *AuthHandler) GetAPIKeysStatus(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	openRouterKey, kieKey, err := h.userRepo.GetAPIKeys(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get API keys status", zap.Error(err), zap.String("user_id", userID.String()))
		response.Error(c, err)
		return
	}

	// Decrypt and check if keys exist (only check non-empty after decryption)
	hasOpenRouterKey := false
	hasKIEKey := false

	if openRouterKey != nil && *openRouterKey != "" {
		decrypted, err := h.cryptoService.Decrypt(*openRouterKey)
		if err != nil {
			h.logger.Warn("failed to decrypt OpenRouter API key", zap.Error(err), zap.String("user_id", userID.String()))
		} else if decrypted != "" {
			hasOpenRouterKey = true
		}
	}

	if kieKey != nil && *kieKey != "" {
		decrypted, err := h.cryptoService.Decrypt(*kieKey)
		if err != nil {
			h.logger.Warn("failed to decrypt KIE API key", zap.Error(err), zap.String("user_id", userID.String()))
		} else if decrypted != "" {
			hasKIEKey = true
		}
	}

	response.Success(c, models.APIKeysStatusResponse{
		HasOpenRouterKey: hasOpenRouterKey,
		HasKIEKey:        hasKIEKey,
	})
}

// UpdateAPIKeys updates the user's API keys
// @Summary Update API keys
// @Description Updates the user's API keys (encrypted at rest)
// @Tags auth
// @Accept json
// @Produce json
// @Param input body models.UpdateAPIKeysInput true "API keys to update"
// @Security BearerAuth
// @Success 200 {object} response.Response{data=models.APIKeysStatusResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/api-keys [put]
func (h *AuthHandler) UpdateAPIKeys(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	var input models.UpdateAPIKeysInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	// Get current keys
	currentOpenRouterKey, currentKIEKey, err := h.userRepo.GetAPIKeys(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get current API keys", zap.Error(err))
		response.Error(c, err)
		return
	}

	// Encrypt new keys if provided, otherwise keep existing
	var encryptedOpenRouterKey, encryptedKIEKey *string

	if input.OpenRouterAPIKey != nil && *input.OpenRouterAPIKey != "" {
		encrypted, err := h.cryptoService.Encrypt(*input.OpenRouterAPIKey)
		if err != nil {
			h.logger.Error("failed to encrypt OpenRouter API key", zap.Error(err))
			response.Error(c, errors.New("failed to encrypt API key"))
			return
		}
		encryptedOpenRouterKey = &encrypted
	} else if input.OpenRouterAPIKey == nil {
		// Keep existing key if not provided
		encryptedOpenRouterKey = currentOpenRouterKey
	}
	// If input.OpenRouterAPIKey is empty string, set to nil (clear the key)

	if input.KIEAPIKey != nil && *input.KIEAPIKey != "" {
		encrypted, err := h.cryptoService.Encrypt(*input.KIEAPIKey)
		if err != nil {
			h.logger.Error("failed to encrypt KIE API key", zap.Error(err))
			response.Error(c, errors.New("failed to encrypt API key"))
			return
		}
		encryptedKIEKey = &encrypted
	} else if input.KIEAPIKey == nil {
		// Keep existing key if not provided
		encryptedKIEKey = currentKIEKey
	}
	// If input.KIEAPIKey is empty string, set to nil (clear the key)

	// Update keys in database
	if err := h.userRepo.UpdateAPIKeys(c.Request.Context(), userID, encryptedOpenRouterKey, encryptedKIEKey); err != nil {
		h.logger.Error("failed to update API keys", zap.Error(err))
		response.Error(c, err)
		return
	}

	h.logger.Info("API keys updated", zap.String("user_id", userID.String()))

	// Return updated status
	response.Success(c, models.APIKeysStatusResponse{
		HasOpenRouterKey: encryptedOpenRouterKey != nil && *encryptedOpenRouterKey != "",
		HasKIEKey:        encryptedKIEKey != nil && *encryptedKIEKey != "",
	})
}

// DeleteAPIKeys removes all API keys for the user
// @Summary Delete API keys
// @Description Removes all API keys for the user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/api-keys [delete]
func (h *AuthHandler) DeleteAPIKeys(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	if err := h.userRepo.DeleteAPIKeys(c.Request.Context(), userID); err != nil {
		h.logger.Error("failed to delete API keys", zap.Error(err))
		response.Error(c, err)
		return
	}

	h.logger.Info("API keys deleted", zap.String("user_id", userID.String()))
	response.NoContent(c)
}

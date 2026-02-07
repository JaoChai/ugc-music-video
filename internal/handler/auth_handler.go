// Package handler provides HTTP handlers for the API.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/jaochai/ugc/internal/external/youtube"
	"github.com/jaochai/ugc/internal/middleware"
	"github.com/jaochai/ugc/internal/models"
	"github.com/jaochai/ugc/internal/repository"
	"github.com/jaochai/ugc/internal/service"
	"github.com/jaochai/ugc/pkg/response"
)

// maxResponseBodySize limits the size of external API response bodies to prevent memory exhaustion
const maxResponseBodySize = 1024 // 1KB

// maxNameLength is the maximum allowed length for user names
const maxNameLength = 100

// maxModelLength is the maximum allowed length for model names
const maxModelLength = 100

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
	authService      service.AuthService
	userRepo         repository.UserRepository
	systemPromptRepo repository.SystemPromptRepository
	cryptoService    service.CryptoService
	youtubeClient    *youtube.Client
	frontendURL      string
	logger           *zap.Logger
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(
	authService service.AuthService,
	userRepo repository.UserRepository,
	systemPromptRepo repository.SystemPromptRepository,
	cryptoService service.CryptoService,
	youtubeClient *youtube.Client,
	frontendURL string,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService:      authService,
		userRepo:         userRepo,
		systemPromptRepo: systemPromptRepo,
		cryptoService:    cryptoService,
		youtubeClient:    youtubeClient,
		frontendURL:      frontendURL,
		logger:           logger,
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
			protected.PATCH("/profile", h.UpdateProfile)
			protected.GET("/api-keys", h.GetAPIKeysStatus)
			protected.PUT("/api-keys", h.UpdateAPIKeys)
			protected.DELETE("/api-keys", h.DeleteAPIKeys)
			protected.POST("/test-openrouter", h.TestOpenRouterConnection)
			protected.POST("/test-kie", h.TestKIEConnection)

			// YouTube OAuth routes
			protected.GET("/youtube/connect", h.YouTubeConnect)
			protected.DELETE("/youtube", h.YouTubeDisconnect)
		}

		// YouTube OAuth callback (not protected — user redirected from Google)
		auth.GET("/youtube/callback", h.YouTubeCallback)
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

	// Check YouTube connection
	hasYouTube := false
	ytToken, err := h.userRepo.GetYouTubeToken(c.Request.Context(), userID)
	if err != nil {
		h.logger.Warn("failed to check YouTube token", zap.Error(err), zap.String("user_id", userID.String()))
	} else if ytToken != nil && *ytToken != "" {
		hasYouTube = true
	}

	response.Success(c, models.APIKeysStatusResponse{
		HasOpenRouterKey: hasOpenRouterKey,
		HasKIEKey:        hasKIEKey,
		HasYouTube:       hasYouTube,
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

// UpdateProfile updates the user's profile (name, openrouter_model)
// @Summary Update user profile
// @Description Updates the user's profile settings
// @Tags auth
// @Accept json
// @Produce json
// @Param input body models.UpdateUserInput true "Profile data to update"
// @Security BearerAuth
// @Success 200 {object} response.Response{data=models.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/profile [patch]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	var input models.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	// Validate input
	if input.Name != nil && len(*input.Name) > maxNameLength {
		response.BadRequest(c, "name must be 100 characters or less")
		return
	}
	if input.OpenRouterModel != nil && len(*input.OpenRouterModel) > maxModelLength {
		response.BadRequest(c, "model name must be 100 characters or less")
		return
	}

	// Get current user
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user", zap.Error(err), zap.String("user_id", userID.String()))
		response.Error(c, err)
		return
	}

	// Update fields if provided
	if input.Name != nil {
		user.Name = input.Name
	}
	if input.OpenRouterModel != nil {
		user.OpenRouterModel = *input.OpenRouterModel
	}

	// Save to database
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		h.logger.Error("failed to update user profile", zap.Error(err), zap.String("user_id", userID.String()))
		response.Error(c, err)
		return
	}

	h.logger.Info("user profile updated", zap.String("user_id", userID.String()))
	response.Success(c, user.ToResponse())
}

// TestConnectionResponse represents the response for API connection tests
type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// TestOpenRouterConnection tests the OpenRouter API connection with user's API key
// @Summary Test OpenRouter API connection
// @Description Tests connectivity to OpenRouter API using the user's saved API key
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=TestConnectionResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/test-openrouter [post]
func (h *AuthHandler) TestOpenRouterConnection(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// Get user's API key
	openRouterKey, _, err := h.userRepo.GetAPIKeys(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get API keys", zap.Error(err), zap.String("user_id", userID.String()))
		response.Error(c, err)
		return
	}

	if openRouterKey == nil || *openRouterKey == "" {
		response.BadRequest(c, "OpenRouter API key not configured")
		return
	}

	// Decrypt the API key
	decryptedKey, err := h.cryptoService.Decrypt(*openRouterKey)
	if err != nil {
		h.logger.Error("failed to decrypt OpenRouter API key", zap.Error(err))
		response.Error(c, errors.New("failed to decrypt API key"))
		return
	}

	// Test the connection by making a simple request to OpenRouter
	success, message := testOpenRouterAPI(c.Request.Context(), decryptedKey, h.logger)

	h.logger.Info("OpenRouter connection test",
		zap.String("user_id", userID.String()),
		zap.Bool("success", success),
	)

	response.Success(c, TestConnectionResponse{
		Success: success,
		Message: message,
	})
}

// TestKIEConnection tests the KIE API connection with user's API key
// @Summary Test KIE API connection
// @Description Tests connectivity to KIE API using the user's saved API key
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=TestConnectionResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/test-kie [post]
func (h *AuthHandler) TestKIEConnection(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// Get user's API key
	_, kieKey, err := h.userRepo.GetAPIKeys(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get API keys", zap.Error(err), zap.String("user_id", userID.String()))
		response.Error(c, err)
		return
	}

	if kieKey == nil || *kieKey == "" {
		response.BadRequest(c, "KIE API key not configured")
		return
	}

	// Decrypt the API key
	decryptedKey, err := h.cryptoService.Decrypt(*kieKey)
	if err != nil {
		h.logger.Error("failed to decrypt KIE API key", zap.Error(err))
		response.Error(c, errors.New("failed to decrypt API key"))
		return
	}

	// Test the connection by making a simple request to KIE
	success, message := testKIEAPI(c.Request.Context(), decryptedKey, h.logger)

	h.logger.Info("KIE connection test",
		zap.String("user_id", userID.String()),
		zap.Bool("success", success),
	)

	response.Success(c, TestConnectionResponse{
		Success: success,
		Message: message,
	})
}

// testOpenRouterAPI tests the OpenRouter API connection
func testOpenRouterAPI(ctx context.Context, apiKey string, logger *zap.Logger) (bool, string) {
	client := &http.Client{Timeout: 10 * time.Second}

	// Make a simple request to list models (lightweight endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		logger.Error("failed to create OpenRouter request", zap.Error(err))
		return false, "Failed to create request"
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("OpenRouter connection failed", zap.Error(err))
		return false, "Connection failed. Please check your network and try again."
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return false, "Invalid API key"
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
		logger.Error("OpenRouter API error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("body", string(body)))
		return false, "API returned an error. Please try again later."
	}

	return true, "Connection successful"
}

// kieCreditsResponse represents the response from KIE credits endpoint
type kieCreditsResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data int    `json:"data"`
}

// testKIEAPI tests the KIE API connection using the credits endpoint
// API docs: https://docs.kie.ai/common-api/get-account-credits
func testKIEAPI(ctx context.Context, apiKey string, logger *zap.Logger) (bool, string) {
	client := &http.Client{Timeout: 10 * time.Second}

	// Use GET /api/v1/chat/credit endpoint per KIE docs
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.kie.ai/api/v1/chat/credit", nil)
	if err != nil {
		logger.Error("failed to create KIE request", zap.Error(err))
		return false, "Failed to create request"
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("KIE connection failed", zap.Error(err))
		return false, "Connection failed. Please check your network and try again."
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))

	// Handle specific error codes per KIE API docs
	switch resp.StatusCode {
	case http.StatusOK:
		var creditsResp kieCreditsResponse
		if err := json.Unmarshal(body, &creditsResp); err != nil {
			logger.Warn("failed to parse KIE credits response", zap.Error(err))
			return true, "Connection successful"
		}
		if creditsResp.Code == 200 {
			return true, fmt.Sprintf("Connection successful. Credits: %d", creditsResp.Data)
		}
		return true, "Connection successful"

	case http.StatusUnauthorized: // 401
		return false, "Invalid API key"

	case http.StatusPaymentRequired: // 402 - Insufficient Credits
		return false, "Insufficient credits in your KIE account"

	case http.StatusUnprocessableEntity: // 422
		return false, "Validation error. Please check your API key format."

	case http.StatusTooManyRequests: // 429
		return false, "Rate limited. Please try again later."

	case 455: // KIE-specific: Service Unavailable
		return false, "KIE service temporarily unavailable. Please try again later."

	case http.StatusInternalServerError: // 500
		return false, "KIE server error. Please try again later."

	case 505: // KIE-specific: Feature Disabled
		return false, "This feature is disabled for your account"

	default:
		logger.Error("KIE API error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("body", string(body)))
		return false, fmt.Sprintf("API error (status %d). Please try again later.", resp.StatusCode)
	}
}

// YouTubeConnect initiates the YouTube OAuth2 flow.
// Returns a URL the frontend should redirect the user to.
func (h *AuthHandler) YouTubeConnect(c *gin.Context) {
	if h.youtubeClient == nil {
		response.BadRequest(c, "YouTube integration is not configured")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// Generate a short-lived JWT as the OAuth state parameter (CSRF protection)
	state, err := h.authService.GenerateShortToken(userID, 10*time.Minute)
	if err != nil {
		h.logger.Error("failed to generate OAuth state token", zap.Error(err))
		response.Error(c, errors.New("failed to initiate YouTube connection"))
		return
	}

	authURL := h.youtubeClient.GetAuthURL(state)

	response.Success(c, gin.H{
		"auth_url": authURL,
	})
}

// settingsRedirect builds a redirect URL to the frontend settings page.
func (h *AuthHandler) settingsRedirect(query string) string {
	if h.frontendURL != "" {
		return h.frontendURL + "/settings?" + query
	}
	return "/settings?" + query
}

// YouTubeCallback handles the OAuth2 callback from Google.
// Exchanges the authorization code for a refresh token, encrypts it, and saves to DB.
func (h *AuthHandler) YouTubeCallback(c *gin.Context) {
	if h.youtubeClient == nil {
		c.Redirect(http.StatusFound, h.settingsRedirect("youtube=error&reason=not_configured"))
		return
	}

	// Check for OAuth error from Google
	if errParam := c.Query("error"); errParam != "" {
		h.logger.Warn("YouTube OAuth error", zap.String("error", errParam))
		c.Redirect(http.StatusFound, h.settingsRedirect("youtube=error&reason="+errParam))
		return
	}

	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.Redirect(http.StatusFound, h.settingsRedirect("youtube=error&reason=missing_params"))
		return
	}

	// Validate the state parameter (JWT) to extract userID and prevent CSRF
	userID, err := h.authService.ValidateShortToken(state)
	if err != nil {
		h.logger.Warn("invalid YouTube OAuth state", zap.Error(err))
		c.Redirect(http.StatusFound, h.settingsRedirect("youtube=error&reason=invalid_state"))
		return
	}

	// Exchange the authorization code for a refresh token
	refreshToken, err := h.youtubeClient.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		h.logger.Error("failed to exchange YouTube OAuth code", zap.Error(err), zap.String("user_id", userID.String()))
		c.Redirect(http.StatusFound, h.settingsRedirect("youtube=error&reason=exchange_failed"))
		return
	}

	// Encrypt the refresh token
	encrypted, err := h.cryptoService.Encrypt(refreshToken)
	if err != nil {
		h.logger.Error("failed to encrypt YouTube refresh token", zap.Error(err))
		c.Redirect(http.StatusFound, h.settingsRedirect("youtube=error&reason=encryption_failed"))
		return
	}

	// Save to database
	if err := h.userRepo.UpdateYouTubeToken(c.Request.Context(), userID, &encrypted); err != nil {
		h.logger.Error("failed to save YouTube token", zap.Error(err), zap.String("user_id", userID.String()))
		c.Redirect(http.StatusFound, h.settingsRedirect("youtube=error&reason=save_failed"))
		return
	}

	h.logger.Info("YouTube connected successfully", zap.String("user_id", userID.String()))
	c.Redirect(http.StatusFound, h.settingsRedirect("youtube=connected"))
}

// YouTubeDisconnect revokes the YouTube OAuth token and removes it from DB.
func (h *AuthHandler) YouTubeDisconnect(c *gin.Context) {
	if h.youtubeClient == nil {
		response.BadRequest(c, "YouTube integration is not configured")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// Get the encrypted token
	encToken, err := h.userRepo.GetYouTubeToken(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get YouTube token", zap.Error(err), zap.String("user_id", userID.String()))
		response.Error(c, errors.New("failed to disconnect YouTube"))
		return
	}

	// Revoke the token if it exists
	if encToken != nil && *encToken != "" {
		refreshToken, err := h.cryptoService.Decrypt(*encToken)
		if err != nil {
			h.logger.Warn("failed to decrypt YouTube token for revocation", zap.Error(err))
		} else {
			if err := h.youtubeClient.RevokeToken(c.Request.Context(), refreshToken); err != nil {
				h.logger.Warn("failed to revoke YouTube token (continuing with disconnect)", zap.Error(err))
			}
		}
	}

	// Remove token from DB
	if err := h.userRepo.UpdateYouTubeToken(c.Request.Context(), userID, nil); err != nil {
		h.logger.Error("failed to remove YouTube token", zap.Error(err), zap.String("user_id", userID.String()))
		response.Error(c, errors.New("failed to disconnect YouTube"))
		return
	}

	h.logger.Info("YouTube disconnected", zap.String("user_id", userID.String()))
	response.NoContent(c)
}

// maxPromptLength is the maximum allowed length for custom prompts
const maxPromptLength = 10000


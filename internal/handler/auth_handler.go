package handler

import (
	"errors"
	"net/http"

	"github.com/example/go-gin-jwt-auth/internal/config"
	"github.com/example/go-gin-jwt-auth/internal/middleware"
	"github.com/example/go-gin-jwt-auth/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RefreshCookieName = "refresh_token"

type AuthHandler struct {
	authService service.AuthService
	cfg         *config.Config
}

func NewAuthHandler(authService service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	h.setRefreshTokenCookie(c, res.RefreshToken, int(h.cfg.RefreshTokenDuration.Seconds()))
	c.JSON(http.StatusCreated, res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate"})
		return
	}

	h.setRefreshTokenCookie(c, res.RefreshToken, int(h.cfg.RefreshTokenDuration.Seconds()))
	c.JSON(http.StatusOK, res)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	// 1. Read refresh token from HttpOnly cookie
	refreshToken, err := c.Cookie(RefreshCookieName)
	if err != nil || refreshToken == "" {
		// Fallback check in body if client sends it explicitly
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&body); err == nil && body.RefreshToken != "" {
			refreshToken = body.RefreshToken
		}
	}

	if refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token is missing"})
		return
	}

	// 2. Perform Refresh Token Rotation with reuse detection
	res, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		// Clear cookie on any refresh failure
		h.clearRefreshTokenCookie(c)

		if errors.Is(err, service.ErrTokenReuseDetected) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Compromised token detected. All sessions revoked for security. Please login again.",
			})
			return
		}
		if errors.Is(err, service.ErrTokenExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired. Please login again."})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// 3. Set the newly rotated Refresh Token into HttpOnly cookie
	h.setRefreshTokenCookie(c, res.RefreshToken, int(h.cfg.RefreshTokenDuration.Seconds()))
	c.JSON(http.StatusOK, res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie(RefreshCookieName)
	if refreshToken != "" {
		_ = h.authService.Logout(refreshToken)
	}

	h.clearRefreshTokenCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user ID"})
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// Helpers for cookie management
func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string, maxAge int) {
	sameSite := http.SameSiteLaxMode
	switch h.cfg.CookieSameSite {
	case "Strict":
		sameSite = http.SameSiteStrictMode
	case "None":
		sameSite = http.SameSiteNoneMode
	}

	c.SetSameSite(sameSite)
	c.SetCookie(
		RefreshCookieName,
		token,
		maxAge,
		"/",
		h.cfg.CookieDomain,
		h.cfg.CookieSecure,
		true, // HttpOnly = true (protects against XSS attacks)
	)
}

func (h *AuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	sameSite := http.SameSiteLaxMode
	switch h.cfg.CookieSameSite {
	case "Strict":
		sameSite = http.SameSiteStrictMode
	case "None":
		sameSite = http.SameSiteNoneMode
	}

	c.SetSameSite(sameSite)
	c.SetCookie(
		RefreshCookieName,
		"",
		-1, // MaxAge < 0 clears the cookie immediately
		"/",
		h.cfg.CookieDomain,
		h.cfg.CookieSecure,
		true,
	)
}

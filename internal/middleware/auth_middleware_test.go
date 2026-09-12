package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/go-gin-jwt-auth/internal/config"
	"github.com/example/go-gin-jwt-auth/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		AccessTokenSecret:   "test-secret-key-for-middleware-32ch",
		AccessTokenDuration: 15 * time.Minute,
	}

	userID := uuid.New()
	userEmail := "user@example.com"
	validToken, _, err := utils.GenerateAccessToken(userID, userEmail, cfg.AccessTokenSecret, cfg.AccessTokenDuration)
	if err != nil {
		t.Fatalf("failed to generate valid token: %v", err)
	}

	expiredToken, _, err := utils.GenerateAccessToken(userID, userEmail, cfg.AccessTokenSecret, -10*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate expired token: %v", err)
	}

	wrongSecretToken, _, err := utils.GenerateAccessToken(userID, userEmail, "wrong-secret-key-1234567890123456", cfg.AccessTokenDuration)
	if err != nil {
		t.Fatalf("failed to generate wrong secret token: %v", err)
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectContext  bool
	}{
		{
			name:           "Valid Bearer token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectContext:  true,
		},
		{
			name:           "Missing Authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name:           "Malformed Authorization header (no Bearer prefix)",
			authHeader:     "Basic " + validToken,
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name:           "Malformed Authorization header (only Bearer)",
			authHeader:     "Bearer",
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name:           "Expired token",
			authHeader:     "Bearer " + expiredToken,
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name:           "Token signed with wrong secret",
			authHeader:     "Bearer " + wrongSecretToken,
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name:           "Garbage token string",
			authHeader:     "Bearer not.a.valid.token",
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(AuthMiddleware(cfg))

			var capturedUserID any
			var capturedUserEmail any

			r.GET("/protected", func(c *gin.Context) {
				capturedUserID, _ = c.Get(ContextUserIDKey)
				capturedUserEmail, _ = c.Get(ContextUserEmailKey)
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected HTTP status %d, got %d. Body: %s", tc.expectedStatus, w.Code, w.Body.String())
			}

			if tc.expectContext {
				if capturedUserID != userID {
					t.Errorf("expected context userID %v, got %v", userID, capturedUserID)
				}
				if capturedUserEmail != userEmail {
					t.Errorf("expected context userEmail %q, got %v", userEmail, capturedUserEmail)
				}
			} else {
				var resp map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("failed to parse error response: %v", err)
				}
				if resp["error"] == "" {
					t.Error("expected non-empty error field in response body")
				}
			}
		})
	}
}

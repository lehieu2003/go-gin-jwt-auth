package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                 string
		method               string
		origin               string
		expectedStatus       int
		expectedAllowOrigin  string
		expectHandlerReached bool
	}{
		{
			name:                 "Standard GET with Origin header",
			method:               http.MethodGet,
			origin:               "http://localhost:3000",
			expectedStatus:       http.StatusOK,
			expectedAllowOrigin:  "http://localhost:3000",
			expectHandlerReached: true,
		},
		{
			name:                 "Standard GET without Origin header (wildcard)",
			method:               http.MethodGet,
			origin:               "",
			expectedStatus:       http.StatusOK,
			expectedAllowOrigin:  "*",
			expectHandlerReached: true,
		},
		{
			name:                 "OPTIONS preflight request",
			method:               http.MethodOptions,
			origin:               "http://frontend.example.com",
			expectedStatus:       http.StatusNoContent,
			expectedAllowOrigin:  "http://frontend.example.com",
			expectHandlerReached: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(CORSMiddleware())

			handlerReached := false
			r.Any("/test", func(c *gin.Context) {
				handlerReached = true
				c.JSON(http.StatusOK, gin.H{"message": "pong"})
			})

			req := httptest.NewRequest(tc.method, "/test", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
			if allowOrigin != tc.expectedAllowOrigin {
				t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tc.expectedAllowOrigin, allowOrigin)
			}

			allowCreds := w.Header().Get("Access-Control-Allow-Credentials")
			if allowCreds != "true" {
				t.Errorf("expected Access-Control-Allow-Credentials 'true', got %q", allowCreds)
			}

			allowMethods := w.Header().Get("Access-Control-Allow-Methods")
			if allowMethods == "" {
				t.Error("expected non-empty Access-Control-Allow-Methods header")
			}

			if handlerReached != tc.expectHandlerReached {
				t.Errorf("expected handlerReached = %v, got %v", tc.expectHandlerReached, handlerReached)
			}
		})
	}
}

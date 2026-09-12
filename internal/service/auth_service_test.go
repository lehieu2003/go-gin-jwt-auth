package service

import (
	"testing"
	"time"

	"github.com/example/go-gin-jwt-auth/internal/config"
	"github.com/example/go-gin-jwt-auth/internal/model"
	"github.com/example/go-gin-jwt-auth/internal/utils"
	"github.com/google/uuid"
)

// In-memory mock for UserRepository
type mockUserRepository struct {
	usersByEmail map[string]*model.User
	usersByID    map[uuid.UUID]*model.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		usersByEmail: make(map[string]*model.User),
		usersByID:    make(map[uuid.UUID]*model.User),
	}
}

func (m *mockUserRepository) Create(user *model.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.usersByEmail[user.Email] = user
	m.usersByID[user.ID] = user
	return nil
}

func (m *mockUserRepository) FindByEmail(email string) (*model.User, error) {
	if user, ok := m.usersByEmail[email]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *mockUserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	if user, ok := m.usersByID[id]; ok {
		return user, nil
	}
	return nil, nil
}

// In-memory mock for TokenRepository
type mockTokenRepository struct {
	tokensByHash map[string]*model.RefreshToken
	tokensByID   map[uuid.UUID]*model.RefreshToken
}

func newMockTokenRepository() *mockTokenRepository {
	return &mockTokenRepository{
		tokensByHash: make(map[string]*model.RefreshToken),
		tokensByID:   make(map[uuid.UUID]*model.RefreshToken),
	}
}

func (m *mockTokenRepository) Create(token *model.RefreshToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	m.tokensByHash[token.TokenHash] = token
	m.tokensByID[token.ID] = token
	return nil
}

func (m *mockTokenRepository) FindByTokenHash(tokenHash string) (*model.RefreshToken, error) {
	if token, ok := m.tokensByHash[tokenHash]; ok {
		return token, nil
	}
	return nil, nil
}

func (m *mockTokenRepository) RevokeToken(id uuid.UUID) error {
	if token, ok := m.tokensByID[id]; ok {
		token.IsRevoked = true
	}
	return nil
}

func (m *mockTokenRepository) RevokeFamily(familyID uuid.UUID) error {
	for _, token := range m.tokensByID {
		if token.FamilyID == familyID {
			token.IsRevoked = true
		}
	}
	return nil
}

func (m *mockTokenRepository) DeleteByUserID(userID uuid.UUID) error {
	for id, token := range m.tokensByID {
		if token.UserID == userID {
			delete(m.tokensByID, id)
			delete(m.tokensByHash, token.TokenHash)
		}
	}
	return nil
}

func (m *mockTokenRepository) DeleteExpiredTokens() error {
	now := time.Now()
	for id, token := range m.tokensByID {
		if now.After(token.ExpiresAt) {
			delete(m.tokensByID, id)
			delete(m.tokensByHash, token.TokenHash)
		}
	}
	return nil
}

func setupAuthService() (AuthService, *mockUserRepository, *mockTokenRepository, *config.Config) {
	cfg := &config.Config{
		AccessTokenSecret:    "unit-test-access-token-secret-key-32ch",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
	}
	userRepo := newMockUserRepository()
	tokenRepo := newMockTokenRepository()
	authSvc := NewAuthService(userRepo, tokenRepo, cfg)
	return authSvc, userRepo, tokenRepo, cfg
}

func TestAuthService_Register(t *testing.T) {
	authSvc, _, _, _ := setupAuthService()

	// 1. Success Registration
	resp, err := authSvc.Register("test@example.com", "Password@123")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if resp == nil || resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatal("expected valid AuthResponse with tokens")
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", resp.User.Email)
	}

	// 2. Duplicate Registration
	_, err = authSvc.Register("test@example.com", "AnotherPassword@123")
	if err != ErrUserAlreadyExists {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestAuthService_Login(t *testing.T) {
	authSvc, _, _, _ := setupAuthService()

	// Setup initial user
	email := "user@example.com"
	password := "CorrectPassword@123"
	_, err := authSvc.Register(email, password)
	if err != nil {
		t.Fatalf("setup registration failed: %v", err)
	}

	// 1. Successful Login
	resp, err := authSvc.Login(email, password)
	if err != nil {
		t.Fatalf("Login failed with valid credentials: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatal("expected non-empty tokens on login")
	}

	// 2. Login with wrong password
	_, err = authSvc.Login(email, "WrongPassword@123")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials for wrong password, got %v", err)
	}

	// 3. Login with non-existent email
	_, err = authSvc.Login("nonexistent@example.com", password)
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials for non-existent user, got %v", err)
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	authSvc, _, tokenRepo, _ := setupAuthService()

	email := "refresh_test@example.com"
	password := "Pass@123456"
	regResp, err := authSvc.Register(email, password)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	rawToken1 := regResp.RefreshToken

	// 1. Success Refresh Rotation
	refreshResp, err := authSvc.RefreshToken(rawToken1)
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if refreshResp.RefreshToken == "" || refreshResp.RefreshToken == rawToken1 {
		t.Fatal("expected new distinct refresh token on rotation")
	}

	// 2. Token Reuse Detection (using rawToken1 again after it has been rotated/revoked)
	_, err = authSvc.RefreshToken(rawToken1)
	if err != ErrTokenReuseDetected {
		t.Errorf("expected ErrTokenReuseDetected on reuse of rotated token, got %v", err)
	}

	// 3. Expired token check
	expiredUser := &model.User{ID: uuid.New(), Email: "expired@example.com"}
	expiredRawToken := "expired-token-raw-string"
	expiredHash := utils.HashToken(expiredRawToken)
	_ = tokenRepo.Create(&model.RefreshToken{
		UserID:    expiredUser.ID,
		TokenHash: expiredHash,
		FamilyID:  uuid.New(),
		IsRevoked: false,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Already expired
	})

	_, err = authSvc.RefreshToken(expiredRawToken)
	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}

	// 4. Non-existent token
	_, err = authSvc.RefreshToken("unknown-token-string")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for unknown token, got %v", err)
	}
}

func TestAuthService_Logout(t *testing.T) {
	authSvc, _, tokenRepo, _ := setupAuthService()

	email := "logout_test@example.com"
	password := "Pass@123456"
	regResp, err := authSvc.Register(email, password)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	rawToken := regResp.RefreshToken
	tokenHash := utils.HashToken(rawToken)

	// Logout
	err = authSvc.Logout(rawToken)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Verify token is revoked in repo
	record, _ := tokenRepo.FindByTokenHash(tokenHash)
	if record == nil || !record.IsRevoked {
		t.Error("expected token to be revoked after logout")
	}

	// Logout with empty string or invalid token should not error
	if err := authSvc.Logout(""); err != nil {
		t.Errorf("expected nil error on empty logout, got %v", err)
	}
}

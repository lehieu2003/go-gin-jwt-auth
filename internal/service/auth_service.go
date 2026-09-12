package service

import (
	"errors"
	"time"

	"github.com/example/go-gin-jwt-auth/internal/config"
	"github.com/example/go-gin-jwt-auth/internal/model"
	"github.com/example/go-gin-jwt-auth/internal/repository"
	"github.com/example/go-gin-jwt-auth/internal/utils"
	"github.com/google/uuid"
)

var (
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid or non-existent refresh token")
	ErrTokenExpired       = errors.New("refresh token has expired")
	ErrTokenReuseDetected = errors.New("refresh token reuse detected: potential security compromise")
)

type AuthResponse struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"-"` // Omitted from JSON body, will be set in HttpOnly cookie
	ExpiresIn    int64       `json:"expires_in"`
}

type AuthService interface {
	Register(email, password string) (*AuthResponse, error)
	Login(email, password string) (*AuthResponse, error)
	RefreshToken(rawRefreshToken string) (*AuthResponse, error)
	Logout(rawRefreshToken string) error
	GetUserByID(userID uuid.UUID) (*model.User, error)
}

type authService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	cfg       *config.Config
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	cfg *config.Config,
) AuthService {
	return &authService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
	}
}

func (s *authService) Register(email, password string) (*AuthResponse, error) {
	existingUser, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        email,
		PasswordHash: hashedPassword,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return s.generateTokenPair(user, uuid.New())
}

func (s *authService) Login(email, password string) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Create a new token family for this login session
	familyID := uuid.New()
	return s.generateTokenPair(user, familyID)
}

// RefreshToken executes Refresh Token Rotation (RTR) with Automatic Reuse Detection
func (s *authService) RefreshToken(rawRefreshToken string) (*AuthResponse, error) {
	if rawRefreshToken == "" {
		return nil, ErrInvalidToken
	}

	tokenHash := utils.HashToken(rawRefreshToken)
	tokenRecord, err := s.tokenRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return nil, err
	}
	if tokenRecord == nil {
		return nil, ErrInvalidToken
	}

	// 1. REUSE DETECTION: If this token was already revoked, a breach is suspected!
	if tokenRecord.IsRevoked {
		// Invalidate all tokens in the entire family to protect the user
		_ = s.tokenRepo.RevokeFamily(tokenRecord.FamilyID)
		return nil, ErrTokenReuseDetected
	}

	// 2. EXPIRY CHECK
	if time.Now().After(tokenRecord.ExpiresAt) {
		_ = s.tokenRepo.RevokeToken(tokenRecord.ID)
		return nil, ErrTokenExpired
	}

	// 3. ROTATE TOKEN:
	// Invalidate the current used token
	if err := s.tokenRepo.RevokeToken(tokenRecord.ID); err != nil {
		return nil, err
	}

	// Find the user associated with this token
	user, err := s.userRepo.FindByID(tokenRecord.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Generate new token pair under the SAME family ID
	return s.generateTokenPair(user, tokenRecord.FamilyID)
}

func (s *authService) Logout(rawRefreshToken string) error {
	if rawRefreshToken == "" {
		return nil
	}

	tokenHash := utils.HashToken(rawRefreshToken)
	tokenRecord, err := s.tokenRepo.FindByTokenHash(tokenHash)
	if err != nil || tokenRecord == nil {
		return nil
	}

	// Invalidate the entire family chain on explicit logout
	return s.tokenRepo.RevokeFamily(tokenRecord.FamilyID)
}

func (s *authService) GetUserByID(userID uuid.UUID) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}

// Helper function to generate access token and rotated refresh token
func (s *authService) generateTokenPair(user *model.User, familyID uuid.UUID) (*AuthResponse, error) {
	// 1. Generate short-lived Access Token
	accessToken, expTime, err := utils.GenerateAccessToken(
		user.ID,
		user.Email,
		s.cfg.AccessTokenSecret,
		s.cfg.AccessTokenDuration,
	)
	if err != nil {
		return nil, err
	}

	// 2. Generate cryptographically secure Refresh Token string
	rawRefreshToken, err := utils.GenerateSecureRandomString(32)
	if err != nil {
		return nil, err
	}

	// 3. Store Refresh Token Hash in Database
	tokenHash := utils.HashToken(rawRefreshToken)
	refreshTokenRecord := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		FamilyID:  familyID,
		IsRevoked: false,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenDuration),
	}

	if err := s.tokenRepo.Create(refreshTokenRecord); err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresIn:    int64(time.Until(expTime).Seconds()),
	}, nil
}

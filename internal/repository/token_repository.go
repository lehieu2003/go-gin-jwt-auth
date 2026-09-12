package repository

import (
	"errors"
	"time"

	"github.com/example/go-gin-jwt-auth/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenRepository interface {
	Create(token *model.RefreshToken) error
	FindByTokenHash(tokenHash string) (*model.RefreshToken, error)
	RevokeToken(id uuid.UUID) error
	RevokeFamily(familyID uuid.UUID) error
	DeleteByUserID(userID uuid.UUID) error
	DeleteExpiredTokens() error
}

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Create(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *tokenRepository) FindByTokenHash(tokenHash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.db.Preload("User").Where("token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *tokenRepository) RevokeToken(id uuid.UUID) error {
	return r.db.Model(&model.RefreshToken{}).Where("id = ?", id).Update("is_revoked", true).Error
}

// RevokeFamily invalidates all tokens within the same family (critical for token reuse detection)
func (r *tokenRepository) RevokeFamily(familyID uuid.UUID) error {
	return r.db.Model(&model.RefreshToken{}).Where("family_id = ?", familyID).Update("is_revoked", true).Error
}

func (r *tokenRepository) DeleteByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.RefreshToken{}).Error
}

func (r *tokenRepository) DeleteExpiredTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&model.RefreshToken{}).Error
}

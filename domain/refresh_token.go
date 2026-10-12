package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// リフレッシュトークンのドメインエンティティ
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// トークンが期限切れかどうかを判定する
func (r *RefreshToken) IsExpired() bool {
	return r.ExpiresAt.Before(time.Now())
}

// トークンが失効済みかどうかを判定する
func (r *RefreshToken) IsRevoked() bool {
	return r.RevokedAt != nil
}

// リフレッシュトークンの DB 操作に関するインターフェース
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeByTokenHash(ctx context.Context, tokenHash string) error
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
}

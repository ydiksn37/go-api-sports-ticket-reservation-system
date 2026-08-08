package domain

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// リフレッシュトークンのドメインエンティティ
// DB の refresh_tokens テーブルと1対1で対応
type RefreshToken struct {
	ID        uuid.UUID    `db:"id"`
	UserID    uuid.UUID    `db:"user_id"`
	TokenHash string       `db:"token_hash"`
	ExpiresAt time.Time    `db:"expires_at"`
	RevokedAt *time.Time   `db:"revoked_at"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

// トークンが期限切れかどうかを判定する
func (r *RefreshToken) IsExpired() bool {
	return r.ExpiresAt.Before(time.Now())
}

// トークンが失効済みかどうかを判定する
func (r *RefreshToken) IsRevoked() bool {
	return r.RevokedAt != nil
}

// リフレッシュトークンの永続化に関するインターフェース
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeByTokenHash(ctx context.Context, tokenHash string) error
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
}

package repository

import (
	"42tokyo-road-to-dena-server/domain"
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// domain.RefreshTokenRepository インターフェースの PostgreSQL 実装
type refreshTokenRepository struct {
	db *sqlx.DB
}

// domain.RefreshTokenRepository の PostgreSQL 実装を返す
func NewRefreshTokenRepository(db *sqlx.DB) domain.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// リフレッシュトークンを DB に保存
func (r *refreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
		token.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("refreshTokenRepository.Create: %w", err)
	}
	return nil
}

// ハッシュ値でリフレッシュトークンを取得
func (r *refreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at, updated_at, deleted_at
		FROM refresh_tokens
		WHERE token_hash = $1
			AND deleted_at IS NULL
			AND revoked_at IS NULL
			AND expires_at > NOW()
	`
	err := r.db.GetContext(ctx, &token, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("refreshTokenRepository.GetByTokenHash: %w", err)
	}
	return &token, nil
}

// ハッシュ値でリフレッシュトークンを失効させる
func (r *refreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE token_hash = $1
			AND deleted_at IS NULL
			AND revoked_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("refreshTokenRepository.RevokeByTokenHash: %w", err)
	}
	return nil
}

// ユーザーIDに紐づく全リフレッシュトークンを失効させる
func (r *refreshTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE user_id = $1
			AND deleted_at IS NULL
			AND revoked_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("refreshTokenRepository.RevokeByUserID: %w", err)
	}
	return nil
}

// コンパイル時に domain.RefreshTokenRepository インターフェースを満たしているか検証
var _ domain.RefreshTokenRepository = (*refreshTokenRepository)(nil)

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ydiksn37/go-api-sports-ticket-reservation-system/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const uniqueViolationCode pq.ErrorCode = "23505"

const (
	usersEmailConstraint    = "users_email_key"
	usersUsernameConstraint = "users_username_key"
)

// domain.UserRepository インターフェースの PostgreSQL 実装
type userRepository struct {
	db *sqlx.DB
}

// コンストラクタ
func NewUserRepository(db *sqlx.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// DB から取得したデータをマッピングするための構造体
type userRow struct {
	ID           uuid.UUID `db:"id"`
	Username     string    `db:"username"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// DB の Row データを domain エンティティに変換する
func toUserDomain(row userRow) *domain.User {
	return &domain.User{
		ID:           row.ID,
		Username:     row.Username,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func classifyUserCreateError(err error) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || pqErr.Code != uniqueViolationCode {
		return err
	}

	switch pqErr.Constraint {
	case usersEmailConstraint:
		return domain.ErrEmailAlreadyExists
	case usersUsernameConstraint:
		return domain.ErrUsernameAlreadyExists
	default:
		return err
	}
}

// 新しいユーザーを DB に保存
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("userRepository.Create: %w", classifyUserCreateError(err))
	}
	return nil
}

// メールアドレスでユーザーを検索する
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var row userRow
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	err := r.db.GetContext(ctx, &row, query, email)
	if err != nil {
		// sql.ErrNoRows はレコードが存在しないことを意味し、エラーではない
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.FindByEmail: %w", err)
	}
	return toUserDomain(row), nil
}

// ユーザー名でユーザーを検索
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var row userRow
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	err := r.db.GetContext(ctx, &row, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.FindByUsername: %w", err)
	}
	return toUserDomain(row), nil
}

// ID でユーザーを検索
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var row userRow
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.FindByID: %w", err)
	}
	return toUserDomain(row), nil
}

// コンパイル時に domain.UserRepository インターフェースを満たしているか検証
var _ domain.UserRepository = (*userRepository)(nil)

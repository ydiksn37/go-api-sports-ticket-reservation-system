package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// システムのユーザーエンティティ
// DB の users テーブルと1対1で対応
type User struct {
	ID           uuid.UUID `db:"id"`
	Username     string    `db:"username"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// ユーザーの永続化に関するインターフェース
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
}

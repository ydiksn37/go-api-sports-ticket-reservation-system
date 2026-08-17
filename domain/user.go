package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// システムのユーザーエンティティ
// DB の users テーブルと1対1で対応
type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ユーザーの DB 操作に関するインターフェース
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
}

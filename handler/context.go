package handler

import (
	"context"

	"github.com/google/uuid"
)

// コンテキストで使用するキーの型
type contextKey string

// コンテキストにユーザー ID を格納するキー
const userIDKey contextKey = "userID"

// コンテキストからユーザー ID を取得する
func getUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}

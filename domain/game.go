package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// チームのエンティティ
// DB の teams テーブルと1対1で対応
type Team struct {
	ID   uuid.UUID
	Name string
}

// 試合のエンティティ
// DB の games テーブルと対応し、JOIN 済みのホーム/アウェイチーム情報を保持する
type Game struct {
	ID         uuid.UUID
	HomeTeamID uuid.UUID
	AwayTeamID uuid.UUID
	GameDate   time.Time
	StartTime  time.Time
	HomeTeam   Team
	AwayTeam   Team
}

// 試合情報の DB 操作に関するインターフェース
type GameRepository interface {
	// 全試合を試合日昇順で取得する（ホーム/アウェイチーム情報を含む）
	FindAll(ctx context.Context) ([]*Game, error)
	// 指定 ID の試合を取得する
	FindByID(ctx context.Context, id uuid.UUID) (*Game, error)
}

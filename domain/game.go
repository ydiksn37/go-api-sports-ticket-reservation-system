package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// チームのエンティティ
type Team struct {
	ID   uuid.UUID
	Name string
}

// 試合のエンティティ
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
	FindAll(ctx context.Context) ([]*Game, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Game, error)
}

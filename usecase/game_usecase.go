package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/ydiksn37/go-api-sports-ticket-reservation-system/domain"

	"github.com/google/uuid"
)

// エラー定義
var (
	// 指定 ID の試合が存在しない場合のエラー
	ErrGameNotFound = errors.New("game not found")
)

// DTO 定義
type ListGamesOutput struct {
	Games []*domain.Game
}

type GetGameInput struct {
	GameID uuid.UUID
}

type GetGameOutput struct {
	Game *domain.Game
}

// 試合情報に関するビジネスロジックのインターフェース
type GameUseCase interface {
	ListGames(ctx context.Context) (*ListGamesOutput, error)
	GetGame(ctx context.Context, input GetGameInput) (*GetGameOutput, error)
}

// DIP のため domain.GameRepository インターフェースに依存
type gameUseCase struct {
	gameRepo domain.GameRepository
}

// コンストラクタ
func NewGameUseCase(gameRepo domain.GameRepository) GameUseCase {
	return &gameUseCase{gameRepo: gameRepo}
}

// 全試合一覧取得のビジネスロジック
func (u *gameUseCase) ListGames(ctx context.Context) (*ListGamesOutput, error) {
	games, err := u.gameRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListGames: %w", err)
	}
	return &ListGamesOutput{Games: games}, nil
}

// 試合詳細取得のビジネスロジック
func (u *gameUseCase) GetGame(ctx context.Context, input GetGameInput) (*GetGameOutput, error) {
	game, err := u.gameRepo.FindByID(ctx, input.GameID)
	if err != nil {
		return nil, fmt.Errorf("GetGame: %w", err)
	}
	if game == nil {
		return nil, ErrGameNotFound
	}
	return &GetGameOutput{Game: game}, nil
}

// コンパイル時に GameUseCase インターフェースを満たしているか検証
var _ GameUseCase = (*gameUseCase)(nil)

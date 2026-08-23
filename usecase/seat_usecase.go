package usecase

import (
	"42tokyo-road-to-dena-server/domain"
	"context"
	"fmt"

	"github.com/google/uuid"
)

// ============================================================
// DTO 定義
// ============================================================

// 座席空き状況取得の入力データ
type ListSeatAvailabilityInput struct {
	GameID uuid.UUID
}

// 座席空き状況取得の出力データ
type ListSeatAvailabilityOutput struct {
	Seats []*domain.SeatAvailability
}

// ============================================================
// インターフェース定義
// Handler はこのインターフェースに依存
// ============================================================

// 座席情報に関するビジネスロジックのインターフェース
type SeatUseCase interface {
	// 指定試合の座席空き状況をグレード別に取得する
	ListSeatAvailability(ctx context.Context, input ListSeatAvailabilityInput) (*ListSeatAvailabilityOutput, error)
}

// ============================================================
// UseCase ロジック
// ============================================================

// DIP のため domain.SeatRepository / domain.GameRepository インターフェースに依存
type seatUseCase struct {
	seatRepo domain.SeatRepository
	gameRepo domain.GameRepository // ゲーム存在確認のため
}

// コンストラクタ
func NewSeatUseCase(seatRepo domain.SeatRepository, gameRepo domain.GameRepository) SeatUseCase {
	return &seatUseCase{seatRepo: seatRepo, gameRepo: gameRepo}
}

// 座席空き状況取得のビジネスロジック
func (u *seatUseCase) ListSeatAvailability(ctx context.Context, input ListSeatAvailabilityInput) (*ListSeatAvailabilityOutput, error) {
	// ゲーム存在確認（存在しない場合は ErrGameNotFound を返す）
	game, err := u.gameRepo.FindByID(ctx, input.GameID)
	if err != nil {
		return nil, fmt.Errorf("ListSeatAvailability: %w", err)
	}
	if game == nil {
		return nil, ErrGameNotFound // game_usecase.go で定義済みのセンチネルエラーを再利用
	}

	seats, err := u.seatRepo.FindAvailabilityByGameID(ctx, input.GameID)
	if err != nil {
		return nil, fmt.Errorf("ListSeatAvailability: %w", err)
	}
	return &ListSeatAvailabilityOutput{Seats: seats}, nil
}

// コンパイル時に SeatUseCase インターフェースを満たしているか検証
var _ SeatUseCase = (*seatUseCase)(nil)

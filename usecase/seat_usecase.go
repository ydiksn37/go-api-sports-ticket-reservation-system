package usecase

import (
	"context"
	"fmt"
	"github.com/ydiksn37/go-api-sports-ticket-reservation-system/domain"

	"github.com/google/uuid"
)

// DTO 定義
type ListSeatAvailabilityInput struct {
	GameID uuid.UUID
}

type ListSeatAvailabilityOutput struct {
	Seats []*domain.SeatAvailability
}

// 座席情報に関するビジネスロジックのインターフェース
type SeatUseCase interface {
	ListSeatAvailability(ctx context.Context, input ListSeatAvailabilityInput) (*ListSeatAvailabilityOutput, error)
}

// DIP のため domain.SeatRepository / domain.GameRepository / domain.ReservationRepository インターフェースに依存
type seatUseCase struct {
	seatRepo        domain.SeatRepository
	gameRepo        domain.GameRepository
	reservationRepo domain.ReservationRepository
}

// コンストラクタ
func NewSeatUseCase(seatRepo domain.SeatRepository, gameRepo domain.GameRepository, reservationRepo domain.ReservationRepository) SeatUseCase {
	return &seatUseCase{seatRepo: seatRepo, gameRepo: gameRepo, reservationRepo: reservationRepo}
}

// 座席空き状況取得のビジネスロジック
func (u *seatUseCase) ListSeatAvailability(ctx context.Context, input ListSeatAvailabilityInput) (*ListSeatAvailabilityOutput, error) {
	if err := u.reservationRepo.ExpireOverdue(ctx); err != nil {
		return nil, fmt.Errorf("ListSeatAvailability: expire overdue: %w", err)
	}

	game, err := u.gameRepo.FindByID(ctx, input.GameID)
	if err != nil {
		return nil, fmt.Errorf("ListSeatAvailability: %w", err)
	}
	if game == nil {
		return nil, ErrGameNotFound
	}

	seats, err := u.seatRepo.FindAvailabilityByGameID(ctx, input.GameID)
	if err != nil {
		return nil, fmt.Errorf("ListSeatAvailability: %w", err)
	}
	return &ListSeatAvailabilityOutput{Seats: seats}, nil
}

// コンパイル時に SeatUseCase インターフェースを満たしているか検証
var _ SeatUseCase = (*seatUseCase)(nil)

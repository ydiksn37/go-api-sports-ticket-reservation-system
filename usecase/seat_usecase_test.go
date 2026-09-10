package usecase_test

import (
	"42tokyo-road-to-dena-server/domain"
	"42tokyo-road-to-dena-server/usecase"
	"42tokyo-road-to-dena-server/usecase/mock"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// ============================================================
// ListSeatAvailability テスト
// ============================================================

func TestListSeatAvailability_Success(t *testing.T) {
	gameID := uuid.New()

	seatRepo := &mock.SeatRepository{
		FindAvailabilityByGameIDFn: func(_ context.Context, _ uuid.UUID) ([]*domain.SeatAvailability, error) {
			return []*domain.SeatAvailability{
				{SeatGrade: "S指定席", Price: 6800, TotalSeats: 10, AvailableSeats: 5, ReservedSeats: 2, SoldSeats: 3},
				{SeatGrade: "A指定席", Price: 5500, TotalSeats: 20, AvailableSeats: 10, ReservedSeats: 2, SoldSeats: 8},
			}, nil
		},
	}
	gameRepo := &mock.GameRepository{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Game, error) {
			return mock.NewGame(id), nil
		},
	}
	reservationRepo := &mock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return nil },
	}

	uc := usecase.NewSeatUseCase(seatRepo, gameRepo, reservationRepo)

	out, err := uc.ListSeatAvailability(context.Background(), usecase.ListSeatAvailabilityInput{GameID: gameID})

	if err != nil {
		t.Fatalf("expected: no error, got %v", err)
	}
	if len(out.Seats) != 2 {
		t.Errorf("expected: 2 seat grades, got %d", len(out.Seats))
	}
}

func TestListSeatAvailability_GameNotFound(t *testing.T) {
	gameRepo := &mock.GameRepository{
		// 試合が見つからない
		FindByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Game, error) { return nil, nil },
	}
	reservationRepo := &mock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return nil },
	}

	uc := usecase.NewSeatUseCase(&mock.SeatRepository{}, gameRepo, reservationRepo)

	_, err := uc.ListSeatAvailability(context.Background(), usecase.ListSeatAvailabilityInput{GameID: uuid.New()})

	if !errors.Is(err, usecase.ErrGameNotFound) {
		t.Errorf("expected: ErrGameNotFound, got %v", err)
	}
}

func TestListSeatAvailability_ExpireOverdueError(t *testing.T) {
	wantErr := fmt.Errorf("expire overdue db error")
	reservationRepo := &mock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return wantErr },
	}

	uc := usecase.NewSeatUseCase(&mock.SeatRepository{}, &mock.GameRepository{}, reservationRepo)

	_, err := uc.ListSeatAvailability(context.Background(), usecase.ListSeatAvailabilityInput{GameID: uuid.New()})

	if !errors.Is(err, wantErr) {
		t.Errorf("expected: expire overdue db error, got %v", err)
	}
}

func TestListSeatAvailability_SeatRepoError(t *testing.T) {
	wantErr := fmt.Errorf("seat db error")
	gameRepo := &mock.GameRepository{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Game, error) {
			return mock.NewGame(id), nil
		},
	}
	seatRepo := &mock.SeatRepository{
		FindAvailabilityByGameIDFn: func(_ context.Context, _ uuid.UUID) ([]*domain.SeatAvailability, error) {
			return nil, wantErr
		},
	}
	reservationRepo := &mock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return nil },
	}

	uc := usecase.NewSeatUseCase(seatRepo, gameRepo, reservationRepo)

	_, err := uc.ListSeatAvailability(context.Background(), usecase.ListSeatAvailabilityInput{GameID: uuid.New()})

	if !errors.Is(err, wantErr) {
		t.Errorf("expected: seat db error, got %v", err)
	}
}

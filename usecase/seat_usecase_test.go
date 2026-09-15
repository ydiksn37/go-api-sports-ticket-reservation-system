package usecase_test

import (
	"42tokyo-road-to-dena-server/domain"
	"42tokyo-road-to-dena-server/usecase"
	repositorymock "42tokyo-road-to-dena-server/usecase/mock"
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

	seatRepo := &repositorymock.SeatRepository{
		FindAvailabilityByGameIDFn: func(_ context.Context, _ uuid.UUID) ([]*domain.SeatAvailability, error) {
			return []*domain.SeatAvailability{
				{SeatGrade: "SS", Price: 6800, TotalSeats: 10, AvailableSeats: 5, ReservedSeats: 2, SoldSeats: 3},
				{SeatGrade: "S", Price: 5500, TotalSeats: 20, AvailableSeats: 10, ReservedSeats: 2, SoldSeats: 8},
			}, nil
		},
	}
	gameRepo := &repositorymock.GameRepository{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Game, error) {
			return repositorymock.NewGame(id), nil
		},
	}
	reservationRepo := &repositorymock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return nil },
	}

	uc := usecase.NewSeatUseCase(seatRepo, gameRepo, reservationRepo)

	out, err := uc.ListSeatAvailability(context.Background(), usecase.ListSeatAvailabilityInput{GameID: gameID})

	if err != nil {
		t.Fatalf("want: no error, got: %v", err)
	}
	if len(out.Seats) != 2 {
		t.Errorf("want: 2 seat grades, got: %d", len(out.Seats))
	}
}

func TestListSeatAvailability_GameNotFound(t *testing.T) {
	gameRepo := &repositorymock.GameRepository{
		// 試合が見つからない
		FindByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Game, error) { return nil, nil },
	}
	reservationRepo := &repositorymock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return nil },
	}

	uc := usecase.NewSeatUseCase(&repositorymock.SeatRepository{}, gameRepo, reservationRepo)

	_, err := uc.ListSeatAvailability(context.Background(), usecase.ListSeatAvailabilityInput{GameID: uuid.New()})

	if !errors.Is(err, usecase.ErrGameNotFound) {
		t.Errorf("want: ErrGameNotFound, got: %v", err)
	}
}

func TestListSeatAvailability_ExpireOverdueError(t *testing.T) {
	wantErr := fmt.Errorf("expire overdue db error")
	reservationRepo := &repositorymock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return wantErr },
	}

	uc := usecase.NewSeatUseCase(&repositorymock.SeatRepository{}, &repositorymock.GameRepository{}, reservationRepo)

	_, err := uc.ListSeatAvailability(context.Background(), usecase.ListSeatAvailabilityInput{GameID: uuid.New()})

	if !errors.Is(err, wantErr) {
		t.Errorf("want: expire overdue db error, got: %v", err)
	}
}

func TestListSeatAvailability_SeatRepoError(t *testing.T) {
	wantErr := fmt.Errorf("seat db error")
	gameRepo := &repositorymock.GameRepository{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Game, error) {
			return repositorymock.NewGame(id), nil
		},
	}
	seatRepo := &repositorymock.SeatRepository{
		FindAvailabilityByGameIDFn: func(_ context.Context, _ uuid.UUID) ([]*domain.SeatAvailability, error) {
			return nil, wantErr
		},
	}
	reservationRepo := &repositorymock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return nil },
	}

	uc := usecase.NewSeatUseCase(seatRepo, gameRepo, reservationRepo)

	_, err := uc.ListSeatAvailability(context.Background(), usecase.ListSeatAvailabilityInput{GameID: uuid.New()})

	if !errors.Is(err, wantErr) {
		t.Errorf("want: seat db error, got: %v", err)
	}
}

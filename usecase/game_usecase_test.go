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
// ListGames テスト
// ============================================================

func TestListGames_Success(t *testing.T) {
	gameID1 := uuid.New()
	gameID2 := uuid.New()

	gameRepo := &mock.GameRepository{
		FindAllFn: func(_ context.Context) ([]*domain.Game, error) {
			return []*domain.Game{
				mock.NewGame(gameID1),
				mock.NewGame(gameID2),
			}, nil
		},
	}
	uc := usecase.NewGameUseCase(gameRepo)

	out, err := uc.ListGames(context.Background())

	if err != nil {
		t.Fatalf("expected: no error, got %v", err)
	}
	if len(out.Games) != 2 {
		t.Errorf("expected: 2 games, got %d", len(out.Games))
	}
}

func TestListGames_Empty(t *testing.T) {
	gameRepo := &mock.GameRepository{
		FindAllFn: func(_ context.Context) ([]*domain.Game, error) {
			return []*domain.Game{}, nil
		},
	}
	uc := usecase.NewGameUseCase(gameRepo)

	out, err := uc.ListGames(context.Background())

	if err != nil {
		t.Fatalf("expected: no error, got %v", err)
	}
	if len(out.Games) != 0 {
		t.Errorf("expected: 0 games, got %d", len(out.Games))
	}
}

func TestListGames_RepoError(t *testing.T) {
	wantErr := fmt.Errorf("db connection error")
	gameRepo := &mock.GameRepository{
		FindAllFn: func(_ context.Context) ([]*domain.Game, error) {
			return nil, wantErr
		},
	}
	uc := usecase.NewGameUseCase(gameRepo)

	_, err := uc.ListGames(context.Background())

	if !errors.Is(err, wantErr) {
		t.Errorf("expected: db connection error, got %v", err)
	}
}

// ============================================================
// GetGame テスト
// ============================================================

func TestGetGame_Success(t *testing.T) {
	gameID := uuid.New()
	expected := mock.NewGame(gameID)

	gameRepo := &mock.GameRepository{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Game, error) {
			return expected, nil
		},
	}
	uc := usecase.NewGameUseCase(gameRepo)

	out, err := uc.GetGame(context.Background(), usecase.GetGameInput{GameID: gameID})

	if err != nil {
		t.Fatalf("expected: no error, got %v", err)
	}
	if out.Game.ID != gameID {
		t.Errorf("expected: game ID %v, got %v", gameID, out.Game.ID)
	}
}

func TestGetGame_NotFound(t *testing.T) {
	gameRepo := &mock.GameRepository{
		// 試合が見つからない
		FindByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Game, error) { return nil, nil },
	}
	uc := usecase.NewGameUseCase(gameRepo)

	_, err := uc.GetGame(context.Background(), usecase.GetGameInput{GameID: uuid.New()})

	if !errors.Is(err, usecase.ErrGameNotFound) {
		t.Errorf("expected: ErrGameNotFound, got %v", err)
	}
}

func TestGetGame_RepoError(t *testing.T) {
	wantErr := fmt.Errorf("db error")
	gameRepo := &mock.GameRepository{
		FindByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Game, error) {
			return nil, wantErr
		},
	}
	uc := usecase.NewGameUseCase(gameRepo)

	_, err := uc.GetGame(context.Background(), usecase.GetGameInput{GameID: uuid.New()})

	if !errors.Is(err, wantErr) {
		t.Errorf("expected: db connection error, got %v", err)
	}
}

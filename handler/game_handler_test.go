package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ydiksn37/go-api-sports-ticket-reservation-system/domain"
	handlermock "github.com/ydiksn37/go-api-sports-ticket-reservation-system/handler/mock"
	"github.com/ydiksn37/go-api-sports-ticket-reservation-system/usecase"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

// handler テスト用の Handler を組み立てるヘルパー
func newTestHandler(deps Deps) *Handler {
	return New(deps)
}

// 認証済みユーザー ID を context に注入したリクエストを返すヘルパー
func withUserID(r *http.Request, id uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), userIDKey, id)
	return r.WithContext(ctx)
}

// ============================================================
// GET /api/games テスト
// ============================================================

func TestListGames_Handler_Success(t *testing.T) {
	gameID := uuid.New()
	ucMock := &handlermock.GameUseCase{
		ListGamesFn: func(_ context.Context) (*usecase.ListGamesOutput, error) {
			return &usecase.ListGamesOutput{
				Games: []*domain.Game{
					{
						ID:       gameID,
						GameDate: time.Now(),
						HomeTeam: domain.Team{ID: uuid.New(), Name: "広島東洋カープ"},
						AwayTeam: domain.Team{ID: uuid.New(), Name: "読売ジャイアンツ"},
					},
					{
						ID:       gameID,
						GameDate: time.Now(),
						HomeTeam: domain.Team{ID: uuid.New(), Name: "福岡ソフトバンクホークス"},
						AwayTeam: domain.Team{ID: uuid.New(), Name: "千葉ロッテマリーンズ"},
					},
				},
			}, nil
		},
	}
	h := newTestHandler(Deps{GameUseCase: ucMock})

	req := httptest.NewRequest(http.MethodGet, "/api/games", nil)
	rec := httptest.NewRecorder()
	h.ListGames(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want: 200, got: %d", rec.Code)
	}
	var resp []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("want: 2, got: %d", len(resp))
	}
}

func TestListGames_Handler_Empty(t *testing.T) {
	ucMock := &handlermock.GameUseCase{
		ListGamesFn: func(_ context.Context) (*usecase.ListGamesOutput, error) {
			return &usecase.ListGamesOutput{Games: []*domain.Game{}}, nil
		},
	}
	h := newTestHandler(Deps{GameUseCase: ucMock})

	req := httptest.NewRequest(http.MethodGet, "/api/games", nil)
	rec := httptest.NewRecorder()
	h.ListGames(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want: 200, got: %d", rec.Code)
	}
	var resp []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
	if resp == nil {
		t.Error("empty list should be [] but got null")
	}
}

func TestListGames_Handler_InternalError(t *testing.T) {
	ucMock := &handlermock.GameUseCase{
		ListGamesFn: func(_ context.Context) (*usecase.ListGamesOutput, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	h := newTestHandler(Deps{GameUseCase: ucMock})

	req := httptest.NewRequest(http.MethodGet, "/api/games", nil)
	rec := httptest.NewRecorder()
	h.ListGames(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("want: 500, got: %d", rec.Code)
	}
}

// ============================================================
// GET /api/games/{id} テスト
// ============================================================

func TestGetGame_Handler_Success(t *testing.T) {
	gameID := uuid.New()
	ucMock := &handlermock.GameUseCase{
		GetGameFn: func(_ context.Context, _ usecase.GetGameInput) (*usecase.GetGameOutput, error) {
			return &usecase.GetGameOutput{
				Game: &domain.Game{
					ID:       gameID,
					GameDate: time.Now(),
					HomeTeam: domain.Team{ID: uuid.New(), Name: "広島東洋カープ"},
					AwayTeam: domain.Team{ID: uuid.New(), Name: "読売ジャイアンツ"},
				},
			}, nil
		},
	}
	h := newTestHandler(Deps{GameUseCase: ucMock})

	req := httptest.NewRequest(http.MethodGet, "/api/games/"+gameID.String(), nil)
	req.SetPathValue("id", gameID.String())
	rec := httptest.NewRecorder()
	h.GetGame(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want: 200, got: %d", rec.Code)
	}
}

func TestGetGame_Handler_InvalidUUID(t *testing.T) {
	h := newTestHandler(Deps{GameUseCase: &handlermock.GameUseCase{}})

	req := httptest.NewRequest(http.MethodGet, "/api/games/not-a-uuid", nil)
	req.SetPathValue("id", "not-a-uuid")
	rec := httptest.NewRecorder()
	h.GetGame(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("want: 400, got: %d", rec.Code)
	}
}

func TestGetGame_Handler_NotFound(t *testing.T) {
	gameID := uuid.New()
	ucMock := &handlermock.GameUseCase{
		GetGameFn: func(_ context.Context, _ usecase.GetGameInput) (*usecase.GetGameOutput, error) {
			return nil, usecase.ErrGameNotFound
		},
	}
	h := newTestHandler(Deps{GameUseCase: ucMock})

	req := httptest.NewRequest(http.MethodGet, "/api/games/"+gameID.String(), nil)
	req.SetPathValue("id", gameID.String())
	rec := httptest.NewRecorder()
	h.GetGame(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("want: 404, got: %d", rec.Code)
	}
}

func TestGetGame_Handler_InternalError(t *testing.T) {
	gameID := uuid.New()
	ucMock := &handlermock.GameUseCase{
		GetGameFn: func(_ context.Context, _ usecase.GetGameInput) (*usecase.GetGameOutput, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	h := newTestHandler(Deps{GameUseCase: ucMock})

	req := httptest.NewRequest(http.MethodGet, "/api/games/"+gameID.String(), nil)
	req.SetPathValue("id", gameID.String())
	rec := httptest.NewRecorder()
	h.GetGame(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("want: 500, got: %d", rec.Code)
	}
}

// ============================================================
// GET /api/games/{id}/seats テスト
// ============================================================

func TestListSeatAvailability_Handler_Success(t *testing.T) {
	gameID := uuid.New()
	ucMock := &handlermock.SeatUseCase{
		ListSeatAvailabilityFn: func(_ context.Context, _ usecase.ListSeatAvailabilityInput) (*usecase.ListSeatAvailabilityOutput, error) {
			return &usecase.ListSeatAvailabilityOutput{
				Seats: []*domain.SeatAvailability{
					{SeatGrade: "SS", Price: 6800, TotalSeats: 10, AvailableSeats: 5},
					{SeatGrade: "S", Price: 5500, TotalSeats: 20, AvailableSeats: 10},
					{SeatGrade: "A", Price: 4000, TotalSeats: 30, AvailableSeats: 15},
					{SeatGrade: "B", Price: 2500, TotalSeats: 40, AvailableSeats: 20},
				},
			}, nil
		},
	}
	h := newTestHandler(Deps{SeatUseCase: ucMock})

	req := httptest.NewRequest(http.MethodGet, "/api/games/"+gameID.String()+"/seats", nil)
	req.SetPathValue("id", gameID.String())
	rec := httptest.NewRecorder()
	h.ListSeatAvailability(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want: 200, got: %d", rec.Code)
	}
	var resp []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
	if len(resp) != 4 {
		t.Errorf("want: 4, got: %d", len(resp))
	}
}

func TestListSeatAvailability_Handler_InvalidUUID(t *testing.T) {
	h := newTestHandler(Deps{SeatUseCase: &handlermock.SeatUseCase{}})

	req := httptest.NewRequest(http.MethodGet, "/api/games/bad-id/seats", nil)
	req.SetPathValue("id", "bad-id")
	rec := httptest.NewRecorder()
	h.ListSeatAvailability(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("want: 400, got: %d", rec.Code)
	}
}

func TestListSeatAvailability_Handler_GameNotFound(t *testing.T) {
	gameID := uuid.New()
	ucMock := &handlermock.SeatUseCase{
		ListSeatAvailabilityFn: func(_ context.Context, _ usecase.ListSeatAvailabilityInput) (*usecase.ListSeatAvailabilityOutput, error) {
			return nil, usecase.ErrGameNotFound
		},
	}
	h := newTestHandler(Deps{SeatUseCase: ucMock})

	req := httptest.NewRequest(http.MethodGet, "/api/games/"+gameID.String()+"/seats", nil)
	req.SetPathValue("id", gameID.String())
	rec := httptest.NewRecorder()
	h.ListSeatAvailability(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("want: 404, got: %d", rec.Code)
	}
}

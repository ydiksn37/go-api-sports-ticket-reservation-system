package handler

import (
	"42tokyo-road-to-dena-server/domain"
	handlermock "42tokyo-road-to-dena-server/handler/mock"
	"42tokyo-road-to-dena-server/usecase"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// テストヘルパー
// ============================================================

// CreateReservation リクエスト JSON を組み立てるヘルパー
func createReservationJSON(gameID string, seats []map[string]interface{}) *bytes.Buffer {
	body := map[string]interface{}{
		"game_id": gameID,
		"seats":   seats,
	}
	b, _ := json.Marshal(body)
	return bytes.NewBuffer(b)
}

// テスト用のダミー Reservation を生成するヘルパー
func newDummyReservation(userID, gameID uuid.UUID) *domain.Reservation {
	now := time.Now()
	return &domain.Reservation{
		ID:        uuid.New(),
		UserID:    userID,
		GameID:    gameID,
		Status:    "pending",
		ExpiresAt: now.Add(15 * time.Minute),
		CreatedAt: now,
		UpdatedAt: now,
		Tickets:   []*domain.Ticket{},
	}
}

// ============================================================
// POST /api/reservations テスト
// ============================================================

func TestCreateReservation_Handler_InvalidGameID(t *testing.T) {
	userID := uuid.New()
	h := newTestHandler(Deps{ReservationUseCase: &handlermock.ReservationUseCase{}})

	body := createReservationJSON("not-a-uuid", []map[string]interface{}{
		{"seat_grade": "S指定席", "quantity": 2},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/reservations", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.CreateReservation(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("want: 400, got: %d", rec.Code)
	}
}

func TestCreateReservation_Handler_EmptySeats(t *testing.T) {
	userID := uuid.New()
	h := newTestHandler(Deps{ReservationUseCase: &handlermock.ReservationUseCase{}})

	body := createReservationJSON(uuid.New().String(), []map[string]interface{}{})
	req := httptest.NewRequest(http.MethodPost, "/api/reservations", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.CreateReservation(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("want: 400, got: %d", rec.Code)
	}
}

func TestCreateReservation_Handler_InvalidQuantity(t *testing.T) {
	userID := uuid.New()
	h := newTestHandler(Deps{ReservationUseCase: &handlermock.ReservationUseCase{}})

	body := createReservationJSON(uuid.New().String(), []map[string]interface{}{
		{"seat_grade": "S指定席", "quantity": 0}, // 0 は無効
	})
	req := httptest.NewRequest(http.MethodPost, "/api/reservations", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.CreateReservation(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("want: 400, got: %d", rec.Code)
	}
}

func TestCreateReservation_Handler_GameNotFound(t *testing.T) {
	userID := uuid.New()
	ucMock := &handlermock.ReservationUseCase{
		CreateReservationFn: func(_ context.Context, _ usecase.CreateReservationInput) (*usecase.CreateReservationOutput, error) {
			return nil, usecase.ErrGameNotFound
		},
	}
	h := newTestHandler(Deps{ReservationUseCase: ucMock})

	body := createReservationJSON(uuid.New().String(), []map[string]interface{}{
		{"seat_grade": "S指定席", "quantity": 1},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/reservations", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.CreateReservation(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("want: 404, got: %d", rec.Code)
	}
}

func TestCreateReservation_Handler_InsufficientSeats(t *testing.T) {
	userID := uuid.New()
	ucMock := &handlermock.ReservationUseCase{
		CreateReservationFn: func(_ context.Context, _ usecase.CreateReservationInput) (*usecase.CreateReservationOutput, error) {
			return nil, usecase.ErrInsufficientSeats
		},
	}
	h := newTestHandler(Deps{ReservationUseCase: ucMock})

	body := createReservationJSON(uuid.New().String(), []map[string]interface{}{
		{"seat_grade": "S指定席", "quantity": 100},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/reservations", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.CreateReservation(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("want: 409, got: %d", rec.Code)
	}
}

func TestCreateReservation_Handler_Success(t *testing.T) {
	userID := uuid.New()
	gameID := uuid.New()
	ucMock := &handlermock.ReservationUseCase{
		CreateReservationFn: func(_ context.Context, _ usecase.CreateReservationInput) (*usecase.CreateReservationOutput, error) {
			return &usecase.CreateReservationOutput{
				Reservation: newDummyReservation(userID, gameID),
			}, nil
		},
	}
	h := newTestHandler(Deps{ReservationUseCase: ucMock})

	body := createReservationJSON(gameID.String(), []map[string]interface{}{
		{"seat_grade": "S指定席", "quantity": 2},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/reservations", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.CreateReservation(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want: 200, got: %d", rec.Code)
	}
}

// ============================================================
// GET /api/reservations/{id} テスト
// ============================================================

func TestGetReservation_Handler_InvalidUUID(t *testing.T) {
	userID := uuid.New()
	h := newTestHandler(Deps{ReservationUseCase: &handlermock.ReservationUseCase{}})

	req := httptest.NewRequest(http.MethodGet, "/api/reservations/bad-id", nil)
	req.SetPathValue("id", "bad-id")
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.GetReservation(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("want: 400, got: %d", rec.Code)
	}
}

func TestGetReservation_Handler_NotFound(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	ucMock := &handlermock.ReservationUseCase{
		GetReservationFn: func(_ context.Context, _ usecase.GetReservationInput) (*usecase.GetReservationOutput, error) {
			return nil, usecase.ErrReservationNotFound
		},
	}
	h := newTestHandler(Deps{ReservationUseCase: ucMock})

	req := httptest.NewRequest(http.MethodGet, "/api/reservations/"+reservationID.String(), nil)
	req.SetPathValue("id", reservationID.String())
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.GetReservation(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("want: 404, got: %d", rec.Code)
	}
}

func TestGetReservation_Handler_Forbidden(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	ucMock := &handlermock.ReservationUseCase{
		GetReservationFn: func(_ context.Context, _ usecase.GetReservationInput) (*usecase.GetReservationOutput, error) {
			return nil, usecase.ErrReservationForbidden
		},
	}
	h := newTestHandler(Deps{ReservationUseCase: ucMock})

	req := httptest.NewRequest(http.MethodGet, "/api/reservations/"+reservationID.String(), nil)
	req.SetPathValue("id", reservationID.String())
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.GetReservation(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("want: 403, got: %d", rec.Code)
	}
}

func TestGetReservation_Handler_Success(t *testing.T) {
	userID := uuid.New()
	gameID := uuid.New()
	reservationID := uuid.New()
	ucMock := &handlermock.ReservationUseCase{
		GetReservationFn: func(_ context.Context, _ usecase.GetReservationInput) (*usecase.GetReservationOutput, error) {
			res := newDummyReservation(userID, gameID)
			res.ID = reservationID
			return &usecase.GetReservationOutput{Reservation: res}, nil
		},
	}
	h := newTestHandler(Deps{ReservationUseCase: ucMock})

	req := httptest.NewRequest(http.MethodGet, "/api/reservations/"+reservationID.String(), nil)
	req.SetPathValue("id", reservationID.String())
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.GetReservation(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want: 200, got: %d", rec.Code)
	}
}

// ============================================================
// PUT /api/reservations/{id}/purchase テスト
// ============================================================

func TestPurchaseReservation_Handler_Expired(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	ucMock := &handlermock.ReservationUseCase{
		PurchaseReservationFn: func(_ context.Context, _ usecase.PurchaseReservationInput) (*usecase.PurchaseReservationOutput, error) {
			return nil, usecase.ErrReservationExpired
		},
	}
	h := newTestHandler(Deps{ReservationUseCase: ucMock})

	req := httptest.NewRequest(http.MethodPut, "/api/reservations/"+reservationID.String()+"/purchase", nil)
	req.SetPathValue("id", reservationID.String())
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.PurchaseReservation(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("want: 409, got: %d", rec.Code)
	}
}

func TestPurchaseReservation_Handler_NotPending(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	ucMock := &handlermock.ReservationUseCase{
		PurchaseReservationFn: func(_ context.Context, _ usecase.PurchaseReservationInput) (*usecase.PurchaseReservationOutput, error) {
			return nil, usecase.ErrReservationNotPending
		},
	}
	h := newTestHandler(Deps{ReservationUseCase: ucMock})

	req := httptest.NewRequest(http.MethodPut, "/api/reservations/"+reservationID.String()+"/purchase", nil)
	req.SetPathValue("id", reservationID.String())
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.PurchaseReservation(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("want: 409, got: %d", rec.Code)
	}
}

func TestPurchaseReservation_Handler_InternalError(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	ucMock := &handlermock.ReservationUseCase{
		PurchaseReservationFn: func(_ context.Context, _ usecase.PurchaseReservationInput) (*usecase.PurchaseReservationOutput, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	h := newTestHandler(Deps{ReservationUseCase: ucMock})

	req := httptest.NewRequest(http.MethodPut, "/api/reservations/"+reservationID.String()+"/purchase", nil)
	req.SetPathValue("id", reservationID.String())
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.PurchaseReservation(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("want: 500, got: %d", rec.Code)
	}
}

func TestPurchaseReservation_Handler_Success(t *testing.T) {
	userID := uuid.New()
	gameID := uuid.New()
	reservationID := uuid.New()
	ucMock := &handlermock.ReservationUseCase{
		PurchaseReservationFn: func(_ context.Context, _ usecase.PurchaseReservationInput) (*usecase.PurchaseReservationOutput, error) {
			res := newDummyReservation(userID, gameID)
			res.ID = reservationID
			res.Status = "confirmed"
			return &usecase.PurchaseReservationOutput{Reservation: res}, nil
		},
	}
	h := newTestHandler(Deps{ReservationUseCase: ucMock})

	req := httptest.NewRequest(http.MethodPut, "/api/reservations/"+reservationID.String()+"/purchase", nil)
	req.SetPathValue("id", reservationID.String())
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.PurchaseReservation(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want: 200, got: %d", rec.Code)
	}
}

// ============================================================
// DELETE /api/reservations/{id} テスト
// ============================================================

func TestCancelReservation_Handler_AlreadyPurchased(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	ucMock := &handlermock.ReservationUseCase{
		CancelReservationFn: func(_ context.Context, _ usecase.CancelReservationInput) error {
			return usecase.ErrReservationAlreadyPurchased
		},
	}
	h := newTestHandler(Deps{ReservationUseCase: ucMock})

	req := httptest.NewRequest(http.MethodDelete, "/api/reservations/"+reservationID.String(), nil)
	req.SetPathValue("id", reservationID.String())
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.CancelReservation(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("want: 409, got: %d", rec.Code)
	}
}

func TestCancelReservation_Handler_Forbidden(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	ucMock := &handlermock.ReservationUseCase{
		CancelReservationFn: func(_ context.Context, _ usecase.CancelReservationInput) error {
			return usecase.ErrReservationForbidden
		},
	}
	h := newTestHandler(Deps{ReservationUseCase: ucMock})

	req := httptest.NewRequest(http.MethodDelete, "/api/reservations/"+reservationID.String(), nil)
	req.SetPathValue("id", reservationID.String())
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.CancelReservation(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("want: 403, got: %d", rec.Code)
	}
}

func TestCancelReservation_Handler_Success(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	ucMock := &handlermock.ReservationUseCase{
		CancelReservationFn: func(_ context.Context, _ usecase.CancelReservationInput) error {
			return nil
		},
	}
	h := newTestHandler(Deps{ReservationUseCase: ucMock})

	req := httptest.NewRequest(http.MethodDelete, "/api/reservations/"+reservationID.String(), nil)
	req.SetPathValue("id", reservationID.String())
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.CancelReservation(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("want: 204, got: %d", rec.Code)
	}
}

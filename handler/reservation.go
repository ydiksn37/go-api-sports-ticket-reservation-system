package handler

import (
	"42tokyo-road-to-dena-server/domain"
	"42tokyo-road-to-dena-server/usecase"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// レスポンス用 DTO
// docs/openapi.yaml の TicketResponse / ReservationResponse スキーマに対応
type ticketResponse struct {
	ID        uuid.UUID `json:"id"`
	SeatID    uuid.UUID `json:"seat_id"`
	Price     int       `json:"price"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type reservationResponse struct {
	ID        uuid.UUID        `json:"id"`
	GameID    uuid.UUID        `json:"game_id"`
	Status    string           `json:"status"`
	ExpiresAt time.Time        `json:"expires_at"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Tickets   []ticketResponse `json:"tickets"`
}

// レスポンス用 DTO に変換するヘルパー
func toTicketResponse(t *domain.Ticket) ticketResponse {
	return ticketResponse{
		ID:        t.ID,
		SeatID:    t.SeatID,
		Price:     t.Price,
		Status:    t.Status,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func toReservationResponse(r *domain.Reservation) reservationResponse {
	tickets := make([]ticketResponse, 0, len(r.Tickets))
	for _, t := range r.Tickets {
		tickets = append(tickets, toTicketResponse(t))
	}
	return reservationResponse{
		ID:        r.ID,
		GameID:    r.GameID,
		Status:    r.Status,
		ExpiresAt: r.ExpiresAt,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
		Tickets:   tickets,
	}
}

// リクエスト用 DTO
// docs/openapi.yaml の CreateReservationRequest スキーマに対応
type seatRequestBody struct {
	SeatGrade string `json:"seat_grade"`
	Quantity  int    `json:"quantity"`
}

type createReservationRequestBody struct {
	GameID string            `json:"game_id"`
	Seats  []seatRequestBody `json:"seats"`
}

// POST /api/reservations のハンドラー
// 認証済み（AuthRequired ミドルウェアで認証）ユーザーの予約を作成
func (h *Handler) CreateReservation(w http.ResponseWriter, r *http.Request) {
	// 認証済みユーザー ID の取得
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		h.respondError(w, errors.New("unauthorized"), http.StatusUnauthorized) // 401
		return
	}

	// リクエスト JSON のデコード
	var body createReservationRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.respondError(w, errors.New("invalid request body"), http.StatusBadRequest) // 400
		return
	}

	// game_id のバリデーション
	gameID, err := uuid.Parse(body.GameID)
	if err != nil {
		h.respondError(w, errors.New("invalid game_id"), http.StatusBadRequest) // 400
		return
	}

	// seats のバリデーション
	if len(body.Seats) == 0 {
		h.respondError(w, errors.New("seats must not be empty"), http.StatusBadRequest) // 400
		return
	}
	for _, s := range body.Seats {
		if s.SeatGrade == "" {
			h.respondError(w, errors.New("seat_grade must not be empty"), http.StatusBadRequest) // 400
			return
		}
		if s.Quantity <= 0 {
			h.respondError(w, errors.New("quantity must be greater than 0"), http.StatusBadRequest) // 400
			return
		}
	}

	// SeatRequest の組み立て
	seatRequests := make([]usecase.SeatRequest, 0, len(body.Seats))
	for _, s := range body.Seats {
		seatRequests = append(seatRequests, usecase.SeatRequest{
			SeatGrade: s.SeatGrade,
			Quantity:  s.Quantity,
		})
	}

	// UseCase の呼び出し
	output, err := h.reservationUseCase.CreateReservation(r.Context(), usecase.CreateReservationInput{
		UserID: userID,
		GameID: gameID,
		Seats:  seatRequests,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrGameNotFound):
			h.respondError(w, err, http.StatusNotFound) // 404
		case errors.Is(err, usecase.ErrInsufficientSeats):
			h.respondError(w, err, http.StatusConflict) // 409
		default:
			h.respondError(w, errors.New("internal server error"), http.StatusInternalServerError) // 500
		}
		return
	}

	// レスポンスを返す
	h.respondJSON(w, toReservationResponse(output.Reservation), http.StatusOK) // 200
}

package handler

import (
	"42tokyo-road-to-dena-server/domain"
	"42tokyo-road-to-dena-server/usecase"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// レスポンス用 DTO
type teamResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type gameResponse struct {
	ID        uuid.UUID    `json:"id"`
	GameDate  time.Time    `json:"game_date"`
	StartTime time.Time    `json:"start_time"`
	HomeTeam  teamResponse `json:"home_team"`
	AwayTeam  teamResponse `json:"away_team"`
}

// レスポンス用 DTO に変換するヘルパー
func toGameResponse(g *domain.Game) gameResponse {
	return gameResponse{
		ID:        g.ID,
		GameDate:  g.GameDate,
		StartTime: g.StartTime,
		HomeTeam:  teamResponse{ID: g.HomeTeam.ID, Name: g.HomeTeam.Name},
		AwayTeam:  teamResponse{ID: g.AwayTeam.ID, Name: g.AwayTeam.Name},
	}
}

// GET /api/games のハンドラー
func (h *Handler) ListGames(w http.ResponseWriter, r *http.Request) {
	output, err := h.gameUseCase.ListGames(r.Context())
	if err != nil {
		h.respondError(w, errors.New("internal server error"), http.StatusInternalServerError) // 500
		return
	}

	resp := make([]gameResponse, 0, len(output.Games))
	for _, g := range output.Games {
		resp = append(resp, toGameResponse(g))
	}

	h.respondJSON(w, resp, http.StatusOK) // 200
}

// GET /api/games/{id} のハンドラー
func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, errors.New("invalid game id"), http.StatusBadRequest) // 400
		return
	}

	// UseCase の呼び出し
	output, err := h.gameUseCase.GetGame(r.Context(), usecase.GetGameInput{GameID: id})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrGameNotFound):
			h.respondError(w, err, http.StatusNotFound) // 404
		default:
			h.respondError(w, errors.New("internal server error"), http.StatusInternalServerError) // 500
		}
		return
	}

	h.respondJSON(w, toGameResponse(output.Game), http.StatusOK) // 200
}

// レスポンス用 DTO
type seatAvailabilityResponse struct {
	SeatGrade      string `json:"seat_grade"`
	Price          int    `json:"price"`
	TotalSeats     int    `json:"total_seats"`
	AvailableSeats int    `json:"available_seats"`
	ReservedSeats  int    `json:"reserved_seats"`
	SoldSeats      int    `json:"sold_seats"`
}

// レスポンス用 DTO に変換するヘルパー
func toSeatAvailabilityResponse(s *domain.SeatAvailability) seatAvailabilityResponse {
	return seatAvailabilityResponse{
		SeatGrade:      s.SeatGrade,
		Price:          s.Price,
		TotalSeats:     s.TotalSeats,
		AvailableSeats: s.AvailableSeats,
		ReservedSeats:  s.ReservedSeats,
		SoldSeats:      s.SoldSeats,
	}
}

// GET /api/games/{id}/seats のハンドラー
func (h *Handler) ListSeatAvailability(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, errors.New("invalid game id"), http.StatusBadRequest) // 400
		return
	}

	// UseCase の呼び出し
	output, err := h.seatUseCase.ListSeatAvailability(r.Context(), usecase.ListSeatAvailabilityInput{GameID: id})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrGameNotFound):
			h.respondError(w, err, http.StatusNotFound) // 404
		default:
			h.respondError(w, errors.New("internal server error"), http.StatusInternalServerError) // 500
		}
		return
	}

	resp := make([]seatAvailabilityResponse, 0, len(output.Seats))
	for _, s := range output.Seats {
		resp = append(resp, toSeatAvailabilityResponse(s))
	}

	h.respondJSON(w, resp, http.StatusOK) // 200
}

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
// docs/openapi.yaml の TeamResponse / GameResponse スキーマに対応

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

	// 空一覧でも null ではなく [] を返す
	resp := make([]gameResponse, 0, len(output.Games))
	for _, g := range output.Games {
		resp = append(resp, toGameResponse(g))
	}

	// レスポンスを返す
	h.respondJSON(w, resp, http.StatusOK) // 200
}

// GET /api/games/{id} のハンドラー
func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	// パスパラメータの取得・バリデーション
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

	// レスポンスを返す
	h.respondJSON(w, toGameResponse(output.Game), http.StatusOK) // 200
}

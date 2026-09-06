package handler

import (
	"42tokyo-road-to-dena-server/infrastructure/authbundle"
	"42tokyo-road-to-dena-server/usecase"
	"encoding/json"
	"log"
	"net/http"
)

// Handler が必要とする依存関係をまとめた構造体
type Deps struct {
	AuthBundle         *authbundle.AuthBundle
	AuthConfig         *authbundle.AuthConfig
	UserUseCase        usecase.UserUseCase
	GameUseCase        usecase.GameUseCase
	SeatUseCase        usecase.SeatUseCase
	ReservationUseCase usecase.ReservationUseCase
}

type Handler struct {
	authBundleService  *authbundle.AuthBundle
	authConfig         *authbundle.AuthConfig
	userUseCase        usecase.UserUseCase
	gameUseCase        usecase.GameUseCase
	seatUseCase        usecase.SeatUseCase
	reservationUseCase usecase.ReservationUseCase
}

func New(deps Deps) *Handler {
	return &Handler{
		authBundleService:  deps.AuthBundle,
		authConfig:         deps.AuthConfig,
		userUseCase:        deps.UserUseCase,
		gameUseCase:        deps.GameUseCase,
		seatUseCase:        deps.SeatUseCase,
		reservationUseCase: deps.ReservationUseCase,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// ルーティング
	mux.HandleFunc("GET /health", h.HealthCheck)

	// ユーザー登録
	mux.HandleFunc("POST /api/user/signup", h.Signup)

	// ユーザー情報取得（認証あり）
	mux.Handle("GET /api/user/me", h.AuthRequired(http.HandlerFunc(h.Me)))

	// 試合情報取得
	mux.HandleFunc("GET /api/games", h.ListGames)
	mux.HandleFunc("GET /api/games/{id}", h.GetGame)

	// 座席空き状況取得
	mux.HandleFunc("GET /api/games/{id}/seats", h.ListSeatAvailability)

	// 予約管理（認証あり）
	mux.Handle("POST /api/reservations", h.AuthRequired(http.HandlerFunc(h.CreateReservation)))
	mux.Handle("GET /api/reservations", h.AuthRequired(http.HandlerFunc(h.ListReservations)))
	mux.Handle("GET /api/reservations/{id}", h.AuthRequired(http.HandlerFunc(h.GetReservation)))
	// 予約確定（決済処理）
	mux.Handle("PUT /api/reservations/{id}/purchase", h.AuthRequired(http.HandlerFunc(h.PurchaseReservation)))
	// 予約キャンセル
	mux.Handle("DELETE /api/reservations/{id}", h.AuthRequired(http.HandlerFunc(h.CancelReservation)))

	// Swagger / OpenAPI 配信
	registerDocsRoutes(mux)

	return mux
}

func (h *Handler) respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *Handler) respondError(w http.ResponseWriter, err error, status int) {
	response := map[string]string{
		"error": err.Error(),
	}
	h.respondJSON(w, response, status)
}

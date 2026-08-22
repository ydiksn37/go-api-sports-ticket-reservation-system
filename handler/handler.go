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
	AuthBundle  *authbundle.AuthBundle
	AuthConfig  *authbundle.AuthConfig
	UserUseCase usecase.UserUseCase
	GameUseCase usecase.GameUseCase
	SeatUseCase usecase.SeatUseCase
}

type Handler struct {
	authBundleService *authbundle.AuthBundle
	authConfig        *authbundle.AuthConfig
	userUseCase       usecase.UserUseCase
	gameUseCase       usecase.GameUseCase
	seatUseCase       usecase.SeatUseCase
}

func New(deps Deps) *Handler {
	return &Handler{
		authBundleService: deps.AuthBundle,
		authConfig:        deps.AuthConfig,
		userUseCase:       deps.UserUseCase,
		gameUseCase:       deps.GameUseCase,
		seatUseCase:       deps.SeatUseCase,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// ルーティング
	mux.HandleFunc("GET /health", h.HealthCheck)

	// ユーザー登録
	mux.HandleFunc("POST /api/user/signup", h.Signup)

	// 認証が必要なエンドポイント（AuthRequired ミドルウェアを適用）
	// AuthRequired が http.Handler を返すため Handle を使用
	mux.Handle("GET /api/user/me", h.AuthRequired(http.HandlerFunc(h.Me)))

	// 試合情報取得
	mux.HandleFunc("GET /api/games", h.ListGames)
	mux.HandleFunc("GET /api/games/{id}", h.GetGame)

	// 座席空き状況取得
	mux.HandleFunc("GET /api/games/{id}/seats", h.ListSeatAvailability)

	// Swagger/OpenAPI 配信
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

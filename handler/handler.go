package handler

import (
	"42tokyo-road-to-dena-server/authbundle"
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
}

type Handler struct {
	authBundleService *authbundle.AuthBundle
	authConfig        *authbundle.AuthConfig
	userUseCase       usecase.UserUseCase
}

func New(deps Deps) *Handler {
	return &Handler{
		authBundleService: deps.AuthBundle,
		authConfig:        deps.AuthConfig,
		userUseCase:       deps.UserUseCase,
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

package handler

import (
	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/usecase"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Deps は Handler が必要とする依存関係をまとめた構造体
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

	// Swagger/OpenAPI 配信
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		repoRoot, _ := os.Getwd()
		p := filepath.Join(repoRoot, "docs", "openapi.yaml")
		w.Header().Set("Content-Type", "application/yaml")
		http.ServeFile(w, r, p)
	})
	mux.HandleFunc("GET /swagger", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("docs", "swagger", "index.html"))
	})
	mux.HandleFunc("GET /swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("docs", "swagger", "index.html"))
	})

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

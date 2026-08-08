package handler

import (
	"net/http"
	"os"
	"path/filepath"
)

// Swagger UI と OpenAPI YAML を提供するルートを mux に登録
func registerDocsRoutes(mux *http.ServeMux) {
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
}

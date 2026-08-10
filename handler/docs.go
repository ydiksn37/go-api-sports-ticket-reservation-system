package handler

import (
	"net/http"
)

// Swagger UI と OpenAPI YAML を提供するルートを mux に登録
const (
	SwaggerDir  = "./docs/swagger"
	OpenAPIPath = "./docs/openapi.yaml"
)

// Swagger UI と OpenAPI YAML を提供するハンドラを mux に登録
func registerDocsRoutes(mux *http.ServeMux) {
	mux.Handle("GET /swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir(SwaggerDir))))
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, OpenAPIPath)
	})
}

package handler

import (
	"42tokyo-road-to-dena-server/infrastructure/authbundle"
	"context"
	"net/http"
	"strings"
)

// アクセストークンを検証し、context に userID を注入する
func (h *Handler) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := h.extractTokenFromRequest(r)

		if token == "" {
			h.respondError(w, authbundle.ErrUnauthorized, http.StatusUnauthorized) // 401
			return
		}

		if h.authBundleService == nil {
			h.respondError(w, authbundle.ErrUnauthorized, http.StatusUnauthorized) // 401
			return
		}

		claims, err := h.authBundleService.ValidateAccessToken(token)
		if err != nil {
			h.respondError(w, authbundle.ErrUnauthorized, http.StatusUnauthorized) // 401
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ヘッダまたは Cookie からトークンを取得する
func (h *Handler) extractTokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	cookie, err := r.Cookie("access_token")
	if err == nil {
		return cookie.Value
	}

	return ""
}

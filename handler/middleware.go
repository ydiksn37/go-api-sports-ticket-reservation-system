package handler

import (
	"42tokyo-road-to-dena-server/infrastructure/authbundle"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

type requestLogState struct {
	userID string
}

type requestLogStateKey struct{}

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
		if state, ok := r.Context().Value(requestLogStateKey{}).(*requestLogState); ok {
			state.userID = claims.UserID.String()
		}
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

type responseWriter struct {
	http.ResponseWriter
	status  int
	errBody bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status >= 400 {
		rw.errBody.Write(b)
	}
	return rw.ResponseWriter.Write(b)
}

func oneLineBody(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, []byte(body)); err == nil {
		return compact.String()
	}

	return strings.NewReplacer("\r", `\r`, "\n", `\n`).Replace(body)
}

func Logger(accessLog, errorLog *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			state := &requestLogState{}
			r = r.WithContext(context.WithValue(r.Context(), requestLogStateKey{}, state))

			bodyBytes, _ := io.ReadAll(io.LimitReader(r.Body, 1024))
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			bodyStr := oneLineBody(string(bodyBytes))
			if r.Method == http.MethodPost && r.URL.Path == "/api/user/signup" {
				bodyStr = "-"
			}

			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			userID := getUserIDStr(state)

			// アクセスログ: 全リクエストを記録
			accessLog.Printf("[ACCESS] method=%s url=%s ip=%s user_id=%s status=%d body=%s",
				r.Method, r.URL.Path, r.RemoteAddr, userID, rw.status, bodyStr)

			// エラーログ: 4xx/5xx のみ記録
			if rw.status >= 400 {
				errorLog.Printf("[ERROR] method=%s url=%s ip=%s user_id=%s status=%d body_req=%s body_res=%s",
					r.Method, r.URL.Path, r.RemoteAddr, userID, rw.status, bodyStr, oneLineBody(rw.errBody.String()))
			}
		})
	}
}

// 未認証リクエストでは "-" を返す
func getUserIDStr(state *requestLogState) string {
	if state.userID == "" {
		return "-"
	}
	return state.userID
}

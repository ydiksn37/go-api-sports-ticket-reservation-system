package handler

import (
	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/usecase"
	"encoding/json"
	"errors"
	"net/http"
)

// docs/openapi.yaml 仕様の SignupRequest スキーマと対応
type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// POST /api/user/signup のハンドラー
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	// リクエストのデコード
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, errors.New("invalid request body"), http.StatusBadRequest) // 400
		return
	}

	// バリデーション
	if req.Name == "" || req.Email == "" || req.Password == "" {
		h.respondError(w, errors.New("name, email, and password are required"), http.StatusBadRequest) // 400
		return
	}

	// UseCase の呼び出し
	output, err := h.userUseCase.Signup(r.Context(), usecase.SignupInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		// UseCase でのエラーを HTTP ステータスにマッピング
		switch {
		case errors.Is(err, usecase.ErrEmailAlreadyExists):
			h.respondError(w, err, http.StatusConflict) // 409
		case errors.Is(err, usecase.ErrUsernameAlreadyExists):
			h.respondError(w, err, http.StatusConflict) // 409
		default:
			h.respondError(w, errors.New("internal server error"), http.StatusInternalServerError) // 500
		}
		return
	}

	// トークン生成
	// authbundle を使ったトークン生成は Handler 層で行う
	// UseCase はトークンの具体的な実装（JWT等）を知らない
	// todo: authbundle/auth_bundle.go は将来的な分割が必要
	accessToken, err := h.authBundleService.GenerateAccessToken(output.UserID)
	if err != nil {
		h.respondError(w, errors.New("failed to generate access token"), http.StatusInternalServerError) // 500
		return
	}
	refreshToken, err := h.authBundleService.GenerateRefreshToken(r.Context(), output.UserID)
	if err != nil {
		h.respondError(w, errors.New("failed to generate refresh token"), http.StatusInternalServerError) // 500
		return
	}

	// Cookie にトークンをセット
	authbundle.SetAuthCookies(w, accessToken, refreshToken, h.authConfig)

	// レスポンスを返す
	// authbundle.AuthTokensResponse に user_id も含める形で返す
	h.respondJSON(w, map[string]interface{}{
		"user_id":       output.UserID,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, http.StatusCreated) // 201
}

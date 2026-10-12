package handler

import (
	"42tokyo-road-to-dena-server/usecase"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

// docs/openapi.yaml 仕様の SignupRequest スキーマと対応
type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
}

// レスポンス用 DTO に変換するヘルパー
func toSignupResponse(output *usecase.SignupOutput, accessToken, refreshToken string) SignupResponse {
	return SignupResponse{
		UserID:       output.UserID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

type MeResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

// レスポンス用 DTO に変換するヘルパー
func toMeResponse(output *usecase.MeOutput) MeResponse {
	return MeResponse{
		ID:       output.UserID,
		Username: output.Username,
		Email:    output.Email,
	}
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
	if !isValidEmail(req.Email) {
		h.respondError(w, errors.New("invalid email format"), http.StatusBadRequest) // 400
		return
	}
	if !isValidUsername(req.Name) {
		h.respondError(w, errors.New("username must be 3-50 characters and contain only alphanumeric, underscore, or hyphen"), http.StatusBadRequest) // 400
		return
	}
	if !isValidPassword(req.Password) {
		h.respondError(w, errors.New("password must be between 8 and 72 characters"), http.StatusBadRequest) // 400
		return
	}

	// UseCase の呼び出し
	output, err := h.userUseCase.Signup(r.Context(), usecase.SignupInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
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
	setAuthCookies(w, accessToken, refreshToken, h.authConfig)

	// レスポンスを返す
	h.respondJSON(w, toSignupResponse(output, accessToken, refreshToken), http.StatusCreated) // 201
}

// GET /api/user/me のハンドラー
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		h.respondError(w, errors.New("unauthorized"), http.StatusUnauthorized) // 401
		return
	}

	// UseCase の呼び出し
	output, err := h.userUseCase.Me(r.Context(), usecase.MeInput{UserID: userID})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			h.respondError(w, err, http.StatusNotFound) // 404
		default:
			h.respondError(w, errors.New("internal server error"), http.StatusInternalServerError) // 500
		}
		return
	}

	// レスポンスを返す
	h.respondJSON(w, toMeResponse(output), http.StatusOK) // 200
}

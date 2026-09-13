package handler

import (
	handlermock "42tokyo-road-to-dena-server/handler/mock"
	"42tokyo-road-to-dena-server/usecase"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// ============================================================
// POST /api/user/signup テスト
// ============================================================

// Signup リクエスト JSON を組み立てるヘルパー
func signupBody(name, email, password string) *bytes.Buffer {
	body := map[string]string{
		"name":     name,
		"email":    email,
		"password": password,
	}
	b, _ := json.Marshal(body)
	return bytes.NewBuffer(b)
}

func TestSignup_Handler_MissingName(t *testing.T) {
	h := newTestHandler(Deps{UserUseCase: &handlermock.UserUseCase{}})

	req := httptest.NewRequest(http.MethodPost, "/api/user/signup", signupBody("", "fakeman@example.com", "pass"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Signup(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("want: 400, got: %d", rec.Code)
	}
}

func TestSignup_Handler_MissingEmail(t *testing.T) {
	h := newTestHandler(Deps{UserUseCase: &handlermock.UserUseCase{}})

	req := httptest.NewRequest(http.MethodPost, "/api/user/signup", signupBody("fakeman", "", "pass"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Signup(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("want: 400, got: %d", rec.Code)
	}
}

func TestSignup_Handler_MissingPassword(t *testing.T) {
	h := newTestHandler(Deps{UserUseCase: &handlermock.UserUseCase{}})

	req := httptest.NewRequest(http.MethodPost, "/api/user/signup", signupBody("fakeman", "fakeman@example.com", ""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Signup(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("want: 400, got: %d", rec.Code)
	}
}

func TestSignup_Handler_InvalidJSON(t *testing.T) {
	h := newTestHandler(Deps{UserUseCase: &handlermock.UserUseCase{}})

	req := httptest.NewRequest(http.MethodPost, "/api/user/signup", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Signup(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("want: 400, got: %d", rec.Code)
	}
}

func TestSignup_Handler_EmailConflict(t *testing.T) {
	ucMock := &handlermock.UserUseCase{
		SignupFn: func(_ context.Context, _ usecase.SignupInput) (*usecase.SignupOutput, error) {
			return nil, usecase.ErrEmailAlreadyExists
		},
	}
	h := newTestHandler(Deps{UserUseCase: ucMock})

	req := httptest.NewRequest(http.MethodPost, "/api/user/signup", signupBody("fakeman", "fakeman@example.com", "pass"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Signup(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("want: 409, got: %d", rec.Code)
	}
}

func TestSignup_Handler_UsernameConflict(t *testing.T) {
	ucMock := &handlermock.UserUseCase{
		SignupFn: func(_ context.Context, _ usecase.SignupInput) (*usecase.SignupOutput, error) {
			return nil, usecase.ErrUsernameAlreadyExists
		},
	}
	h := newTestHandler(Deps{UserUseCase: ucMock})

	req := httptest.NewRequest(http.MethodPost, "/api/user/signup", signupBody("fakeman", "fakeman@example.com", "pass"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Signup(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("want: 409, got: %d", rec.Code)
	}
}

// ============================================================
// GET /api/user/me テスト
// ============================================================

func TestMe_Handler_Success(t *testing.T) {
	userID := uuid.New()
	ucMock := &handlermock.UserUseCase{
		MeFn: func(_ context.Context, input usecase.MeInput) (*usecase.MeOutput, error) {
			return &usecase.MeOutput{
				UserID:   input.UserID,
				Username: "fakeman",
				Email:    "fakeman@example.com",
			}, nil
		},
	}
	h := newTestHandler(Deps{UserUseCase: ucMock})

	req := httptest.NewRequest(http.MethodGet, "/api/user/me", nil)
	req = withUserID(req, userID) // context に userID を注入
	rec := httptest.NewRecorder()
	h.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want: 200, got: %d", rec.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
	if resp["username"] != "fakeman" {
		t.Errorf("want: username=fakeman, got: %v", resp["username"])
	}
}

func TestMe_Handler_NoUserIDInContext(t *testing.T) {
	// context に userID がない場合 → 401
	h := newTestHandler(Deps{UserUseCase: &handlermock.UserUseCase{}})

	req := httptest.NewRequest(http.MethodGet, "/api/user/me", nil)
	// withUserID を呼ばない = userID なし
	rec := httptest.NewRecorder()
	h.Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want: 401, got: %d", rec.Code)
	}
}

func TestMe_Handler_UserNotFound(t *testing.T) {
	userID := uuid.New()
	ucMock := &handlermock.UserUseCase{
		MeFn: func(_ context.Context, _ usecase.MeInput) (*usecase.MeOutput, error) {
			return nil, usecase.ErrUserNotFound
		},
	}
	h := newTestHandler(Deps{UserUseCase: ucMock})

	req := httptest.NewRequest(http.MethodGet, "/api/user/me", nil)
	req = withUserID(req, userID)
	rec := httptest.NewRecorder()
	h.Me(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("want: 404, got: %d", rec.Code)
	}
}

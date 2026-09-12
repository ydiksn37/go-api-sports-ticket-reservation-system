package usecase_test

import (
	"42tokyo-road-to-dena-server/domain"
	"42tokyo-road-to-dena-server/usecase"
	repositorymock "42tokyo-road-to-dena-server/usecase/mock"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// ============================================================
// Signup テスト
// ============================================================

func TestSignup_Success(t *testing.T) {
	userRepo := &repositorymock.UserRepository{
		// メール・ユーザー名ともに重複なし
		FindByEmailFn:    func(_ context.Context, _ string) (*domain.User, error) { return nil, nil },
		FindByUsernameFn: func(_ context.Context, _ string) (*domain.User, error) { return nil, nil },
		CreateFn:         func(_ context.Context, _ *domain.User) error { return nil },
	}
	uc := usecase.NewUserUseCase(userRepo, &repositorymock.PasswordHasher{})

	out, err := uc.Signup(context.Background(), usecase.SignupInput{
		Name:     "fakeman",
		Email:    "fakeman@example.com",
		Password: "password",
	})

	if err != nil {
		t.Fatalf("expected: no error, got %v", err)
	}
	if out == nil {
		t.Fatal("expected: output, got nil")
	}
	if out.UserID == uuid.Nil {
		t.Error("expected: non-nil UserID")
	}
}

func TestSignup_EmailAlreadyExists(t *testing.T) {
	userRepo := &repositorymock.UserRepository{
		// メール重複あり
		FindByEmailFn: func(_ context.Context, email string) (*domain.User, error) {
			return repositorymock.NewUser(uuid.New(), "fakeman", email), nil
		},
	}
	uc := usecase.NewUserUseCase(userRepo, &repositorymock.PasswordHasher{})

	_, err := uc.Signup(context.Background(), usecase.SignupInput{
		Name:     "fakeman",
		Email:    "fakeman@example.com",
		Password: "password",
	})

	if !errors.Is(err, usecase.ErrEmailAlreadyExists) {
		t.Errorf("expected: ErrEmailAlreadyExists, got %v", err)
	}
}

func TestSignup_UsernameAlreadyExists(t *testing.T) {
	userRepo := &repositorymock.UserRepository{
		// メールは重複なし・ユーザー名は重複あり
		FindByEmailFn: func(_ context.Context, _ string) (*domain.User, error) { return nil, nil },
		FindByUsernameFn: func(_ context.Context, username string) (*domain.User, error) {
			return repositorymock.NewUser(uuid.New(), username, "fakeman@example.com"), nil
		},
	}
	uc := usecase.NewUserUseCase(userRepo, &repositorymock.PasswordHasher{})

	_, err := uc.Signup(context.Background(), usecase.SignupInput{
		Name:     "fakeman",
		Email:    "fakeman@example.com",
		Password: "password",
	})

	if !errors.Is(err, usecase.ErrUsernameAlreadyExists) {
		t.Errorf("expected: ErrUsernameAlreadyExists, got %v", err)
	}
}

func TestSignup_HashError(t *testing.T) {
	wantErr := errors.New("bcrypt failure")
	userRepo := &repositorymock.UserRepository{
		FindByEmailFn:    func(_ context.Context, _ string) (*domain.User, error) { return nil, nil },
		FindByUsernameFn: func(_ context.Context, _ string) (*domain.User, error) { return nil, nil },
	}
	hasher := &repositorymock.PasswordHasher{
		HashPasswordFn: func(_ string) (string, error) { return "", wantErr },
	}
	uc := usecase.NewUserUseCase(userRepo, hasher)

	_, err := uc.Signup(context.Background(), usecase.SignupInput{
		Name:     "fakeman",
		Email:    "fakeman@example.com",
		Password: "password",
	})

	if !errors.Is(err, wantErr) {
		t.Errorf("expected: bcrypt failure error, got %v", err)
	}
}

func TestSignup_RepoFindByEmailError(t *testing.T) {
	wantErr := fmt.Errorf("db connection error")
	userRepo := &repositorymock.UserRepository{
		FindByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, wantErr
		},
	}
	uc := usecase.NewUserUseCase(userRepo, &repositorymock.PasswordHasher{})

	_, err := uc.Signup(context.Background(), usecase.SignupInput{
		Name:     "fakeman",
		Email:    "fakeman@example.com",
		Password: "password",
	})

	if !errors.Is(err, wantErr) {
		t.Errorf("expected: db connection error, got %v", err)
	}
}

// ============================================================
// Me テスト
// ============================================================

func TestMe_Success(t *testing.T) {
	userID := uuid.New()
	expected := repositorymock.NewUser(userID, "fakeman", "fakeman@example.com")

	userRepo := &repositorymock.UserRepository{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			return expected, nil
		},
	}
	uc := usecase.NewUserUseCase(userRepo, &repositorymock.PasswordHasher{})

	out, err := uc.Me(context.Background(), usecase.MeInput{UserID: userID})

	if err != nil {
		t.Fatalf("expected: no error, got %v", err)
	}
	if out.UserID != expected.ID {
		t.Errorf("expected: UserID %v, got %v", userID, out.UserID)
	}
	if out.Username != expected.Username {
		t.Errorf("expected: Username %q, got %q", expected.Username, out.Username)
	}
	if out.Email != expected.Email {
		t.Errorf("expected: Email %q, got %q", expected.Email, out.Email)
	}
}

func TestMe_NotFound(t *testing.T) {
	userRepo := &repositorymock.UserRepository{
		// ユーザーが見つからない
		FindByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.User, error) { return nil, nil },
	}
	uc := usecase.NewUserUseCase(userRepo, &repositorymock.PasswordHasher{})

	_, err := uc.Me(context.Background(), usecase.MeInput{UserID: uuid.New()})

	if !errors.Is(err, usecase.ErrUserNotFound) {
		t.Errorf("expected: ErrUserNotFound, got %v", err)
	}
}

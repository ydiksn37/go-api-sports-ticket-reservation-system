package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/ydiksn37/go-api-sports-ticket-reservation-system/domain"
	"time"

	"github.com/google/uuid"
)

// エラー定義
var (
	ErrEmailAlreadyExists    = domain.ErrEmailAlreadyExists
	ErrUsernameAlreadyExists = domain.ErrUsernameAlreadyExists
	ErrUserNotFound          = errors.New("user not found")
)

// DTO 定義
type SignupInput struct {
	Name     string
	Email    string
	Password string
}

type SignupOutput struct {
	UserID uuid.UUID
}

type MeInput struct {
	UserID uuid.UUID
}

type MeOutput struct {
	UserID   uuid.UUID
	Username string
	Email    string
}

// ユーザーに関するビジネスロジックのインターフェース
type UserUseCase interface {
	Signup(ctx context.Context, input SignupInput) (*SignupOutput, error)
	Me(ctx context.Context, input MeInput) (*MeOutput, error)
}

// DIP のため domain.UserRepository / domain.PasswordHasher インターフェースに依存
type userUseCase struct {
	userRepo domain.UserRepository
	hasher   domain.PasswordHasher
}

// コンストラクタ
func NewUserUseCase(userRepo domain.UserRepository, hasher domain.PasswordHasher) UserUseCase {
	return &userUseCase{userRepo: userRepo, hasher: hasher}
}

// ユーザー登録のビジネスロジック
func (u *userUseCase) Signup(ctx context.Context, input SignupInput) (*SignupOutput, error) {
	existingByEmail, err := u.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("Signup: find by email: %w", err)
	}
	if existingByEmail != nil {
		return nil, ErrEmailAlreadyExists
	}
	existingByUsername, err := u.userRepo.FindByUsername(ctx, input.Name)
	if err != nil {
		return nil, fmt.Errorf("Signup: find by username: %w", err)
	}
	if existingByUsername != nil {
		return nil, ErrUsernameAlreadyExists
	}
	hashedPassword, err := u.hasher.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("Signup: hash password: %w", err)
	}
	// User エンティティの生成
	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		Username:     input.Name,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	// DB への保存
	if err := u.userRepo.Create(ctx, user); err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			return nil, ErrEmailAlreadyExists
		case errors.Is(err, domain.ErrUsernameAlreadyExists):
			return nil, ErrUsernameAlreadyExists
		}
		return nil, fmt.Errorf("Signup: create user: %w", err)
	}
	return &SignupOutput{UserID: user.ID}, nil
}

// 認証済みユーザーの情報取得ビジネスロジック
func (u *userUseCase) Me(ctx context.Context, input MeInput) (*MeOutput, error) {
	user, err := u.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("Me: find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return &MeOutput{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

// コンパイル時に UserUseCase インターフェースを満たしているか検証
var _ UserUseCase = (*userUseCase)(nil)

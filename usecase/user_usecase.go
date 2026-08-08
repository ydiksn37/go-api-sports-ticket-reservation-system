package usecase

import (
	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// センチネルエラー定義
// ============================================================

var (
	// メールアドレスが既に登録されている場合のエラー
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ユーザー名が既に登録されている場合のエラー
	ErrUsernameAlreadyExists = errors.New("username already exists")
)

// ============================================================
// DTO 定義
// ============================================================

// Signup ユースケースへの入力データ（json タグを持たない）
type SignupInput struct {
	Name     string
	Email    string
	Password string
}

// Signup ユースケースからの出力データ
// トークン生成は Handler 側で行う（authbundle に依存しない）
type SignupOutput struct {
	UserID uuid.UUID
}

// ============================================================
// インターフェース定義
// Handler はこのインターフェースに依存
// ============================================================

// ユーザーに関するビジネスロジックのインターフェース
type UserUseCase interface {
	Signup(ctx context.Context, input SignupInput) (*SignupOutput, error)
}

// ============================================================
// UseCase ロジック
// ============================================================

// DIP のため domain.UserRepository インターフェースに依存
type userUseCase struct {
	userRepo domain.UserRepository
}

// コンストラクタ
func NewUserUseCase(userRepo domain.UserRepository) UserUseCase {
	return &userUseCase{userRepo: userRepo}
}

// ユーザー登録のビジネスロジック
func (u *userUseCase) Signup(ctx context.Context, input SignupInput) (*SignupOutput, error) {
	// メールアドレスの重複チェック
	existingByEmail, err := u.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("Signup: find by email: %w", err)
	}
	if existingByEmail != nil {
		// センチネルエラーを返す
		return nil, ErrEmailAlreadyExists
	}
	// ユーザー名の重複チェック
	existingByUsername, err := u.userRepo.FindByUsername(ctx, input.Name)
	if err != nil {
		return nil, fmt.Errorf("Signup: find by username: %w", err)
	}
	if existingByUsername != nil {
		// センチネルエラーを返す
		return nil, ErrUsernameAlreadyExists
	}
	// パスワードのハッシュ化
	hashedPassword, err := authbundle.HashPassword(input.Password)
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
		return nil, fmt.Errorf("Signup: create user: %w", err)
	}
	return &SignupOutput{UserID: user.ID}, nil
}

// コンパイル時に UserUseCase インターフェースを満たしているか検証
var _ UserUseCase = (*userUseCase)(nil)

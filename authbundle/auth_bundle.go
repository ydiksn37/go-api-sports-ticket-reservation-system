package authbundle

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"42tokyo-road-to-dena-server/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// エラー定数
var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidInput = errors.New("invalid input")
	ErrInternal     = errors.New("internal server error")
)

// ============================================================================
// 認証機能統合ファイル
// 認証サービス（JWT・パスワード）に特化した機能を提供する
// ============================================================================

// リクエスト/レスポンス型
type AuthLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

type AuthTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// JWT 認証・Cookie の設定をまとめた構造体
type AuthConfig struct {
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	CookieDomain string
	CookieSecure bool
}

// ============================================================================
// RefreshTokenStore - authbundle 提供の DB アクセス実装
// domain.RefreshTokenRepository インターフェースを満たす
// ============================================================================

type RefreshTokenStore struct {
	db *sqlx.DB
}

// RefreshTokenStore のコンストラクタ
func NewRefreshTokenStore(db *sqlx.DB) *RefreshTokenStore {
	return &RefreshTokenStore{db: db}
}

// リフレッシュトークンを DB に保存
func (st *RefreshTokenStore) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := st.db.ExecContext(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
		token.UpdatedAt,
	)
	return err
}

// ハッシュ値でリフレッシュトークンを取得
func (st *RefreshTokenStore) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at, updated_at, deleted_at
		FROM refresh_tokens
		WHERE token_hash = $1
			AND deleted_at IS NULL
			AND revoked_at IS NULL
			AND expires_at > NOW()
	`
	err := st.db.GetContext(ctx, &token, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

// ハッシュ値でリフレッシュトークンを失効させる
func (st *RefreshTokenStore) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE token_hash = $1
			AND deleted_at IS NULL
			AND revoked_at IS NULL
	`
	_, err := st.db.ExecContext(ctx, query, tokenHash)
	return err
}

// ユーザーIDに紐づく全リフレッシュトークンを失効させる
func (st *RefreshTokenStore) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE user_id = $1
			AND deleted_at IS NULL
			AND revoked_at IS NULL
	`
	_, err := st.db.ExecContext(ctx, query, userID)
	return err
}

// コンパイル時に domain.RefreshTokenRepository インターフェースを満たしているか検証
var _ domain.RefreshTokenRepository = (*RefreshTokenStore)(nil)

// ============================================================================
// AuthBundle - 認証サービスの統括構造体
// ============================================================================

type AuthBundle struct {
	cfg               *AuthConfig
	refreshTokenStore domain.RefreshTokenRepository // インターフェースに依存（DIP適用）
}

// AuthBundle のコンストラクタ
func NewAuthBundle(cfg *AuthConfig, refreshTokenStore domain.RefreshTokenRepository) *AuthBundle {
	return &AuthBundle{
		cfg:               cfg,
		refreshTokenStore: refreshTokenStore,
	}
}

// ============================================================================
// JWT（アクセストークン）
// ============================================================================

// JWT のクレーム定義
type AuthClaims struct {
	UserID uuid.UUID `json:"sub"`
	jwt.RegisteredClaims
}

// アクセストークン（JWT）を生成する
func (a *AuthBundle) GenerateAccessToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := AuthClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(a.cfg.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    a.cfg.JWTIssuer,
			Audience:  []string{a.cfg.JWTAudience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("%w: failed to sign token: %v", ErrInternal, err)
	}

	return tokenString, nil
}

// アクセストークンを検証し、クレームを返す
func (a *AuthBundle) ValidateAccessToken(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrUnauthorized, token.Header["alg"])
		}
		return []byte(a.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrUnauthorized
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !ok {
		return nil, ErrUnauthorized
	}

	// Issuer 検証
	if claims.Issuer != a.cfg.JWTIssuer {
		return nil, ErrUnauthorized
	}

	// Audience 検証
	validAudience := false
	for _, aud := range claims.Audience {
		if aud == a.cfg.JWTAudience {
			validAudience = true
			break
		}
	}
	if !validAudience {
		return nil, ErrUnauthorized
	}

	return claims, nil
}

// ============================================================================
// リフレッシュトークン
// ============================================================================

// リフレッシュトークンを生成し DB に保存する
func (a *AuthBundle) GenerateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// ランダムトークン生成
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("%w: failed to generate random token: %v", ErrInternal, err)
	}
	tokenString := hex.EncodeToString(tokenBytes)

	// ハッシュ化して DB に保存
	hash := sha256.Sum256([]byte(tokenString))
	tokenHash := hex.EncodeToString(hash[:])

	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(a.cfg.RefreshTTL),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := a.refreshTokenStore.Create(ctx, refreshToken); err != nil {
		return "", fmt.Errorf("%w: failed to save refresh token: %v", ErrInternal, err)
	}

	return tokenString, nil
}

// リフレッシュトークンを検証する
func (a *AuthBundle) ValidateRefreshToken(ctx context.Context, tokenString string) (*domain.RefreshToken, error) {
	// ハッシュ化して DB 検索
	hash := sha256.Sum256([]byte(tokenString))
	tokenHash := hex.EncodeToString(hash[:])

	token, err := a.refreshTokenStore.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, ErrUnauthorized
	}

	if token.IsExpired() {
		return nil, ErrUnauthorized
	}

	if token.IsRevoked() {
		return nil, ErrUnauthorized
	}

	return token, nil
}

// 古いリフレッシュトークンを失効させ、新しいトークンを発行する
func (a *AuthBundle) RotateRefreshToken(ctx context.Context, oldTokenString string) (string, error) {
	// 旧トークン検証
	oldToken, err := a.ValidateRefreshToken(ctx, oldTokenString)
	if err != nil {
		return "", err
	}

	// 旧トークン失効
	hash := sha256.Sum256([]byte(oldTokenString))
	tokenHash := hex.EncodeToString(hash[:])
	if err := a.refreshTokenStore.RevokeByTokenHash(ctx, tokenHash); err != nil {
		return "", fmt.Errorf("%w: failed to revoke old token: %v", ErrInternal, err)
	}

	// 新トークン生成
	newToken, err := a.GenerateRefreshToken(ctx, oldToken.UserID)
	if err != nil {
		return "", fmt.Errorf("%w: failed to generate new token: %v", ErrInternal, err)
	}

	return newToken, nil
}

// ============================================================================
// パスワード（純粋関数）
// ============================================================================

// パスワードを bcrypt でハッシュ化する
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// パスワードとハッシュを比較して一致するか検証する
func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

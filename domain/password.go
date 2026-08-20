package domain

// パスワードのハッシュ化・検証を抽象化するインターフェース
// usecase 層はこのインターフェースに依存
type PasswordHasher interface {
	// パスワードをハッシュ化して返す
	HashPassword(password string) (string, error)
	// パスワードとハッシュが一致するか検証する
	CheckPassword(password, hash string) bool
}

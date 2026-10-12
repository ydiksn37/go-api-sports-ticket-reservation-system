package domain

// パスワードのハッシュ化・検証を抽象化するインターフェース
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
}

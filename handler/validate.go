package handler

import (
	"regexp"
	"unicode/utf8"
)

// ============================================================
// email バリデーション
// ============================================================

// RFC 5322 の簡易版にもとづく email 形式の正規表現
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// email が有効な形式かチェックする
func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// ============================================================
// username バリデーション
// ============================================================

// 英数字・アンダースコア・ハイフンのみ、3〜50文字を許可する正規表現
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]{3,50}$`)

// username が有効な形式かチェックする
func isValidUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

// ============================================================
// password バリデーション
// ============================================================

const (
	passwordMinLength = 8
	// bcrypt の入力は最大 72 バイトまで有効
	passwordMaxLength = 72
)

// password の文字数が有効な範囲かチェックする
// 文字種は問わず、8〜72 文字（Unicode のルーン数）を許可する
func isValidPassword(password string) bool {
	length := utf8.RuneCountInString(password)
	return length >= passwordMinLength && length <= passwordMaxLength
}

// ============================================================
// seat_grade バリデーション
// ============================================================

// DB の seed データに対応した有効な座席グレードの一覧
var validSeatGrades = map[string]bool{
	"SS": true,
	"S":  true,
	"A":  true,
	"B":  true,
}

// 座席グレードが有効な値かチェックする
func isValidSeatGrade(grade string) bool {
	return validSeatGrades[grade]
}

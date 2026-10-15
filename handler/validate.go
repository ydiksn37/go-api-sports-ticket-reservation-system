package handler

import (
	"regexp"
	"unicode/utf8"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// 英数字・アンダースコア・ハイフンのみ、3〜50文字を許可する正規表現
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]{3,50}$`)

func isValidUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

const (
	passwordMinLength = 8
	passwordMaxLength = 72
)

func isValidPassword(password string) bool {
	length := utf8.RuneCountInString(password)
	return length >= passwordMinLength && length <= passwordMaxLength
}

var validSeatGrades = map[string]bool{
	"SS": true,
	"S":  true,
	"A":  true,
	"B":  true,
}

func isValidSeatGrade(grade string) bool {
	return validSeatGrades[grade]
}

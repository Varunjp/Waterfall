package validation

import (
	"net/mail"
	"strings"
	"unicode"
)

func IsVaildEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil 
}

func IsValidName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false 
	}

	firstRune := []rune(name)[0]
	return !unicode.IsDigit(firstRune)
}
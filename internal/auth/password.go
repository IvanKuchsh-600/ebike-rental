package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword — превращает пароль в bcrypt-хеш.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword — проверяет, что пароль соответствует хешу.
func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

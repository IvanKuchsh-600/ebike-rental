package auth

import "fmt"

// JWTGenerator — адаптер для использования GenerateToken через интерфейс TokenGenerator.
type JWTGenerator struct {
	secret      string
	expiryHours int
}

func NewJWTGenerator(secret string, expiryHours int) *JWTGenerator {
	return &JWTGenerator{
		secret:      secret,
		expiryHours: expiryHours,
	}
}

func (g *JWTGenerator) Generate(userID int64, email, role string) (string, error) {
	token, err := GenerateToken(g.secret, g.expiryHours, userID, email, role)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return token, nil
}

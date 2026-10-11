package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/IvanKuchsh-600/ebike-rental/internal/auth"
	"github.com/IvanKuchsh-600/ebike-rental/internal/domain/user"
)

// contextKey — приватный тип для ключей в context.
// Используем свой тип, чтобы избежать коллизий с ключами из других пакетов.
type contextKey string

const (
	contextKeyUserID contextKey = "user_id"
	contextKeyEmail  contextKey = "email"
	contextKeyRole   contextKey = "role"
)

// RequireAuth — middleware, проверяющий JWT.
// Кладёт user_id, email, role в контекст запроса.
func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Достаём заголовок Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		// 2. Проверяем формат "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenString := parts[1]

		// 3. Парсим и валидируем токен
		claims, err := auth.ParseToken(secret, tokenString)
		if err != nil {
			if errors.Is(err, auth.ErrExpiredToken) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// 4. Кладём данные в контекст
		c.Set(string(contextKeyUserID), claims.UserID)
		c.Set(string(contextKeyEmail), claims.Email)
		c.Set(string(contextKeyRole), claims.Role)

		c.Next()
	}
}

// RequireRole — middleware, проверяющий роль.
// Должен использоваться ПОСЛЕ RequireAuth.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *gin.Context) {
		role, exists := c.Get(string(contextKeyRole))
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		roleStr, ok := role.(string)
		if !ok || !allowed[roleStr] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Next()
	}
}

// UserIDFromContext — достаёт user_id из контекста.
func UserIDFromContext(c *gin.Context) (int64, bool) {
	v, exists := c.Get(string(contextKeyUserID))
	if !exists {
		return 0, false
	}

	id, ok := v.(int64)
	return id, ok
}

// RoleFromContext — достаёт role из контекста.
func RoleFromContext(c *gin.Context) (string, bool) {
	v, exists := c.Get(string(contextKeyRole))
	if !exists {
		return "", false
	}
	role, ok := v.(string)
	return role, ok
}

// UserFromContext — достаёт User из контекста (используется редко).
// Если нужно только user_id — используй UserIDFromContext.
func UserFromContext(c *gin.Context) (*user.User, bool) {
	id, ok := UserIDFromContext(c)
	if !ok {
		return nil, false
	}
	role, _ := RoleFromContext(c)
	email, _ := c.Get(string(contextKeyEmail))
	emailStr, _ := email.(string)

	return &user.User{
		ID:    id,
		Email: emailStr,
		Role:  role,
	}, true
}

// ContextWithUserID — кладёт user_id в context.Context (не gin.Context).
// Нужно, когда user_id надо передать в usecase через обычный ctx.
func ContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, contextKeyUserID, userID)
}

// UserIDFromStdContext — достаёт user_id из context.Context.
func UserIDFromStdContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(contextKeyUserID)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

// src/backend/internal/infrastructure/middleware/auth.go
package middleware

import (
	"message-service/internal/domain/token"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	tokenRepo token.Repository
}

func NewAuthMiddleware(tokenRepo token.Repository) *AuthMiddleware {
	return &AuthMiddleware{
		tokenRepo: tokenRepo,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for POST /api/tokens (token creation)
		if c.Request.Method == "POST" && c.Request.URL.Path == "/api/tokens" {
			c.Next()
			return
		}

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication format"})
			return
		}

		tokenString := parts[1]

		// Validate token
		tkn, err := m.tokenRepo.FindByToken(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error occurred during token validation"})
			return
		}

		if tkn == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Store token information in context (if needed)
		c.Set("token", tkn)
		c.Set("token_id", tkn.ID.Hex())

		c.Next()
	}
}

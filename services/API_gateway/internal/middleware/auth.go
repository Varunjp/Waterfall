package middleware

import (
	"api_gateway/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(secret string) gin.HandlerFunc {

	publicPaths := map[string]bool{
		"/health": true,
		"/api/v1/jobs": true,
		"/api/v1/admin/login": true,
		"/api/v1/users/login":true,
		"/api/v1/users/password/reset/request": true,
		"/api/v1/users/password/reset/verify": true,
		"/api/v1/users/password/reset": true,
		"/api/v1/apps": true,
		"/billing/webhook":true,
		"/billing/checkout":true,
	}

	return func(c *gin.Context) {

		if publicPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing auth header"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := utils.ValidateToken(token, secret)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		c.Set("user", claims)
		c.Next()
	}
}
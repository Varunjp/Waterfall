package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const AppIDKey contextKey = "app_id"

func AuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// apiKey := r.Header.Get("X-API-KEY")

		// if apiKey == "" {
		// 	http.Error(w, "missing api key", http.StatusUnauthorized)
		// 	return
		// }

		// Example: validate api key and get app id
		// appID := ValidateAPIKey(apiKey)

		// if appID == "" {
		// 	http.Error(w, "invalid api key", http.StatusUnauthorized)
		// 	return
		// }
		jwtSecret := os.Getenv("JWT_SECRET")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrAbortHandler
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		appID, ok := claims["app_id"].(string)
		if !ok || appID == "" {
			http.Error(w, "app_id missing in token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), AppIDKey,appID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetAppID(ctx context.Context) string {

	val := ctx.Value(AppIDKey)

	if val == nil {
		return ""
	}

	return val.(string)
}
package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const AppIDKey contextKey = "app_id"

func AuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		apiKey := r.Header.Get("X-API-KEY")

		if apiKey == "" {
			http.Error(w, "missing api key", http.StatusUnauthorized)
			return
		}

		// Example: validate api key and get app id
		// appID := ValidateAPIKey(apiKey)

		// if appID == "" {
		// 	http.Error(w, "invalid api key", http.StatusUnauthorized)
		// 	return
		// }

		ctx := context.WithValue(r.Context(), AppIDKey,apiKey)

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
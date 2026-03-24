package utils

import (
	"admin_service/internal/transport/grpc/interceptors"
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
)


func GetAppIDFromContext(ctx context.Context)(string,error) {
	claims,ok := ctx.Value(interceptors.ClaimsKey).(jwt.MapClaims)
	if !ok {
		return "",errors.New("claims not found in context")
	}

	appID,ok := claims["app_id"].(string)
	if !ok || appID == "" {
		return "",errors.New("app_id missing in token")
	}

	return appID,nil 
}
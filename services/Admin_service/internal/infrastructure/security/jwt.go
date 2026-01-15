package security

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(secret, userID, role, appID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":userID,
		"role": role,
		"app_id": appID,
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	return token.SignedString([]byte(secret))
}
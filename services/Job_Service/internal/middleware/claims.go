package middleware

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	AppID string `json:"app_id"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}
package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ResetClaims struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

func GenerateResetToken(email string)(string,error) {
	secret := os.Getenv("JWT_SECRET")

	claims := ResetClaims {
		Email: email,
		Purpose: "password_reset",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)

	return token.SignedString(secret)
}

func ParseResetToken(tokenString string)(string,error) {

	secret := os.Getenv("JWT_SECRET")

	token,err := jwt.ParseWithClaims(
		tokenString,
		&ResetClaims{},
		func(token *jwt.Token)(interface{},error) {
			return secret,nil 
		},
	)

	if err != nil {
		return "",err 
	}

	claims,ok := token.Claims.(*ResetClaims)
	if !ok || !token.Valid {
		return "",errors.New("invalid token")
	}

	if claims.Purpose != "password_reset" {
		return "",errors.New("invalid token purpose")
	}

	return claims.Email, nil 
}
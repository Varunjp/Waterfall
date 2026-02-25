package errors

import "errors"

var (
	ErrInvalidRole = errors.New("invalid role")
	ErrAppEmailAlreadyExists = errors.New("app email already exists")
	ErrAppNameAlreadyExists  = errors.New("app name already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
)
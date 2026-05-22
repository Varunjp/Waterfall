package errors

import "errors"

var (
	ErrInvalidRole           = errors.New("invalid role")
	ErrAppEmailAlreadyExists = errors.New("app email already exists")
	ErrAppNameAlreadyExists  = errors.New("app name already exists")
	ErrPlanIDRequired        = errors.New("plan id is required")
	ErrPlanNotFound          = errors.New("plan not found")
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrUserBlocked           = errors.New("User is not active, please contact administrator")
)

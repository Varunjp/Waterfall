package domain

import "errors"

var (
	ErrForbidden        = errors.New("forbidden")
	ErrMaxRetryExceeded = errors.New("max retry exceeded")
	ErrNotFound         = errors.New("resource not found")
)
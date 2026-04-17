package middleware

import (
	"context"
	"errors"
)

type ctxKey string

const (
	ctxAppID ctxKey = "app_id"
	ctxRole  ctxKey = "role"
)

var (
	ErrMissingAppID = errors.New("app_id not found in context")
	ErrMissingRole  = errors.New("role not found in context")
)

func AppIDFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(ctxAppID)
	if v == nil {
		return "", ErrMissingAppID
	}

	appID, _ := v.(string)
	if appID == "" {
		return "", ErrMissingAppID
	}

	return appID, nil
}

func RoleFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(ctxRole)
	if v == nil {
		return "", ErrMissingRole
	}

	role, _ := v.(string)
	if role == "" {
		return "", ErrMissingRole
	}

	return role, nil
}

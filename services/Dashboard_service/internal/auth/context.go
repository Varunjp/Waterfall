package auth

import (
	"context"
	"errors"
)

type ctxKey string

const (
	ctxAppID ctxKey = "app_id"
	ctxRole  ctxKey = "role"
	ctxSub   ctxKey = "sub"
)

func SubjectFromContext(ctx any) string {
	return ctx.(interface {
		Value(key any) any
	}).Value(ctxSub).(string)
}

var ErrMissingAppID = errors.New("app_id not found in context")
var ErrMissingRole = errors.New("role not found in context")

func AppIDFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(ctxAppID)
	if v == nil {
		return "", ErrMissingAppID
	}
	return v.(string), nil
}

func RoleFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(ctxRole)
	if v == nil {
		return "", ErrMissingRole
	}
	return v.(string), nil
}
package auth

type ctxKey string

const (
	ctxAppID ctxKey = "app_id"
	ctxRole  ctxKey = "role"
	ctxSub   ctxKey = "sub"
)

func AppIDFromContext(ctx any) string {
	return ctx.(interface {
		Value(key any) any
	}).Value(ctxAppID).(string)
}

func RoleFromContext(ctx any) string {
	return ctx.(interface {
		Value(key any) any
	}).Value(ctxRole).(string)
}

func SubjectFromContext(ctx any) string {
	return ctx.(interface {
		Value(key any) any
	}).Value(ctxSub).(string)
}
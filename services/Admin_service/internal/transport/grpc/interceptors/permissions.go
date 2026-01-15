package interceptors

var MethodPermissions = map[string][]string{
	// Platform admin only
	"/admin.AdminService/Login":     {"platform_admin"},
	"/admin.AppService/RegisterApp": {"platform_admin"},
	"/admin.AppService/ListApps":    {"platform_admin"},
	"/admin.AppService/BlockApp":    {"platform_admin"},
	"/admin.AppService/UnblockApp":  {"platform_admin"},

	// App management
	"/admin.AppUserService/CreateUser": {"super_admin", "admin"},
	"/admin.AppUserService/ListUsers":  {"super_admin", "admin", "viewer"},
}

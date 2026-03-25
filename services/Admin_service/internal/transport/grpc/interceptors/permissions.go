package interceptors

var MethodPermissions = map[string][]string{
	// Platform admin only
	"/admin.AdminService/Login":      {"platform_admin"},
	"/admin.AppService/RegisterApp":  {"platform_admin"},
	"/admin.AppService/ListApps":     {"platform_admin"},
	"/admin.AppService/BlockApp":     {"platform_admin"},
	"/admin.AppService/UnblockApp":   {"platform_admin"},
	"/admin.AdminService/CreatePlan": {"platform_admin"},
	"/admin.AdminService/ListPlans":  {"platform_admin"},
	"/admin.AdminService/UpdatePlan": {"platform_admin"},

	// App management
	"/admin.AppUserService/CreateUser":       {"super_admin"},
	"/admin.AppUserService/ListUsers":        {"super_admin", "admin", "viewer"},
	"/admin.AppUserService/UpdateUserStatus": {"super_admin", "admin"},
}

var publicMethods = map[string]bool{
	"/admin.AppService/RegisterApp":   true,
	"/admin.AdminService/Login":       true,
	"/admin.AppUserService/AppLogin":  true,
	"/admin.AppUserService/ListPlans": true,
}

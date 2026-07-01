package enums

type UserStatus string

const (
	SuperAdmin UserStatus = "super_admin"
	Admin      UserStatus = "admin"
	Viewer     UserStatus = "viewer"
)

var ValidUserRoles = map[string]bool{
	string(SuperAdmin): true,
	string(Admin):      true,
	string(Viewer):     true,
}

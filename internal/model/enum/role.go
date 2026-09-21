package enum

const (
	UserRoleAdmin   UserRole = "admin"
	UserRoleUser    UserRole = "user"
	UserRoleUnknown UserRole = "unknown"
)

type UserRole string

func (r UserRole) String() string { return string(r) }

func ParseUserRole(v string) UserRole {
	switch v {
	case "admin":
		return UserRoleAdmin
	case "user":
		return UserRoleUser
	default:
		return UserRoleUnknown
	}
}

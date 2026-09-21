package enum

const (
	StatusApproved UserStatus = "approved" // 已通过
	StatusBlocked  UserStatus = "blocked"  // 拉黑中
	StatusUnknown  UserStatus = "unknown"
)

type UserStatus string

func (s UserStatus) String() string { return string(s) }

func ParseUserStatus(v string) UserStatus {
	switch v {
	case "approved":
		return StatusApproved
	case "blocked":
		return StatusBlocked
	default:
		return StatusUnknown
	}
}

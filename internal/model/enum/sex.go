package enum

const (
	UserSexMale   UserSex = "male"
	UserSexFemale UserSex = "female"
	UserSexSecret UserSex = "secret"
)

type UserSex string

func (s UserSex) String() string { return string(s) }

func ParseUserSex(v string) UserSex {
	switch v {
	case "male":
		return UserSexMale
	case "female":
		return UserSexFemale
	case "secret":
		return UserSexSecret
	default:
		return UserSexSecret
	}
}

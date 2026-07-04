package repo

import (
	"gorm.io/gorm"
)

// IUserRepo defines the interface for user-related data access operations.

type UserPepo struct {
	DB *gorm.DB
}

func NewUserPepo(client *SQLClient) *UserPepo {
	return &UserPepo{
		DB: client.DB,
	}
}

func (u *UserPepo) Login() {

}

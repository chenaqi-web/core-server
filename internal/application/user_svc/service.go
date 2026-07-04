// Package user provides application-level use cases for the user subdomain.
package user_svc

import (
	"backend/core-server/internal/domain/user_domain"
)

type UserService struct {
	repo user_domain.UserDomain
}

func NewUserService(repo user_domain.UserDomain) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (u *UserService) Login() {
	u.repo.Login()
}

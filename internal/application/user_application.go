package application

import (
	"context"
	"core-server/internal/domain"
	"core-server/internal/infras/cache"
	"core-server/internal/model/entity"
	"errors"
	"strings"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserDisabled       = errors.New("user is disabled")
	ErrEmailTooFrequent   = errors.New("email code sent too frequently")
)

type UserService struct {
	repo  domain.UserRepoDomain
	cache *cache.CacheClient
}

func NewUserService(
	repo domain.UserRepoDomain,
	cacheClient *cache.CacheClient,
) *UserService {
	return &UserService{
		repo:  repo,
		cache: cacheClient,
	}
}

func (s *UserService) Login(ctx context.Context, username, password string) (*entity.User, error) {
	// todo 这里还需要判断密码，后续再写
	user, err := s.repo.GetByName(ctx, strings.TrimSpace(username))
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) EmailLogin(ctx context.Context, email string) (*entity.User, error) {
	user, err := s.repo.GetByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		return nil, err
	}
	return user, nil
}

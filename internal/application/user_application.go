package application

import (
	"backend/core-server/internal/config"
	"backend/core-server/internal/domain"
	"backend/core-server/internal/infras/clog"
)

// 编排层，编排逻辑

type UserService struct {
	cfg   *config.Config
	log   *clog.Log
	repo  domain.UserRepoDomain
	cache domain.UserCacheDomain
}

func NewUserService(
	log *clog.Log,
	repo domain.UserRepoDomain,
	userCache domain.UserCacheDomain,
	cfg *config.Config,
) (*UserService, error) {
	return &UserService{
		cfg:   cfg,
		repo:  repo,
		cache: userCache,
		log:   log,
	}, nil
}

// =====================================================================================================================
// login

// =====================================================================================================================
// 点赞操作

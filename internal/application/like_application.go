package application

import (
	"backend/core-server/internal/domain"
	"backend/core-server/internal/infras/cache"
)

type LikeService struct {
	repo  domain.LikeDomain
	cache *cache.ILikeCache
}

func NewLikeService(repo domain.LikeDomain, cache *cache.ILikeCache) *LikeService {
	return &LikeService{
		repo:  repo,
		cache: cache,
	}
}

func (like *LikeService) Like() {

}

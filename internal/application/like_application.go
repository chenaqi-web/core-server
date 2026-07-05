package application

import (
	"backend/core-server/internal/domain/like_domain"
	"backend/core-server/internal/infras/cache"
)

type LikeService struct {
	repo  like_domain.LikeDomain
	cache *cache.ILikeCache
}

func NewLikeService(repo like_domain.LikeDomain, cache *cache.ILikeCache) *LikeService {
	return &LikeService{
		repo:  repo,
		cache: cache,
	}
}

func (like *LikeService) Like() {

}

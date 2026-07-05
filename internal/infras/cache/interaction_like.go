package cache

type ILikeCache struct {
	Cache *CacheClient
}

func NewILikeCache(cache *CacheClient) *ILikeCache {
	return &ILikeCache{
		Cache: cache,
	}
}

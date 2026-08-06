package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"core-server/internal/infras/cache/scripts"

	"github.com/redis/go-redis/v9"
)

const (
	// zset 最大长度
	defaultMaxLikeSetSize      int64         = 50
	defaultLikeListExpiration  time.Duration = 7 * 24 * time.Hour
	defaultLikeCountExpiration time.Duration = 7 * 24 * time.Hour
)

// 记录用户的点赞关系的列表key
func thumbUpListKey(userID uint64, objectType string) string {
	return fmt.Sprintf("like:list:%d:%s", userID, objectType)
}

// 记录对象的点赞数量key
func objectThumbUpCountKey(objectID uint64, objectType string) string {
	return fmt.Sprintf("like:count:%s:%d", objectType, objectID)
}

type ILikeCache struct {
	*CacheClient        // 继承
	maxLikeSetSize      int64
	likeListExpiration  time.Duration
	likeCountExpiration time.Duration
	thumbUpScript       *redis.Script
	cancelThumbUpScript *redis.Script
}

func NewILikeCache(cache *CacheClient) *ILikeCache {
	return &ILikeCache{
		CacheClient:         cache,
		maxLikeSetSize:      defaultMaxLikeSetSize,
		likeListExpiration:  defaultLikeListExpiration,
		likeCountExpiration: defaultLikeCountExpiration,
		thumbUpScript:       redis.NewScript(scripts.ThumbUpLuaScript),
		cancelThumbUpScript: redis.NewScript(scripts.CancelThumbUpLuaScript),
	}
}

func (c *ILikeCache) randomLikeCountExpiration() time.Duration {
	return randomExpiration(c.likeCountExpiration, 0.3)
}

func (c *ILikeCache) randomLikeListExpiration() time.Duration {
	return randomExpiration(c.likeListExpiration, 0.3)
}

//======================================================================================================================
// 点赞操作

func (c *ILikeCache) ThumbUp(ctx context.Context, userID uint64, objectType string, objectID uint64, score int64) error {
	keyZSet := thumbUpListKey(userID, objectType)
	keyCount := objectThumbUpCountKey(objectID, objectType)
	objectIDStr := strconv.FormatUint(objectID, 10)

	keys := []string{keyZSet, keyCount}
	argv := []interface{}{
		c.maxLikeSetSize,
		score,
		objectIDStr,
		int64(c.randomLikeListExpiration().Seconds()),
		int64(c.randomLikeCountExpiration().Seconds()),
	}

	res, err := c.thumbUpScript.Run(ctx, c.Cache, keys, argv...).Result()
	if err != nil {
		return err
	}

	arr, ok := res.([]interface{})
	if !ok {
		return fmt.Errorf("unexpected result type from lua script: %T", res)
	}
	if len(arr) < 2 {
		return fmt.Errorf("unexpected array length for results, expected at least 2 but got %d", len(arr))
	}

	code, ok := arr[0].(int64)
	if !ok {
		return fmt.Errorf("unexpected type for result code: %T", arr[0])
	}
	msg, ok := arr[1].(string)
	if !ok {
		return fmt.Errorf("unexpected type for result message: %T", arr[1])
	}
	if code == 0 {
		return fmt.Errorf("%w: %s", ErrLuaScriptExecFailure, msg)
	}
	return nil
}

func (c *ILikeCache) CancelThumbUp(ctx context.Context, userID uint64, objectType string, objectID uint64) (int, int64, error) {
	keyZSet := thumbUpListKey(userID, objectType)
	keyCount := objectThumbUpCountKey(objectID, objectType)
	objectIDStr := strconv.FormatUint(objectID, 10)

	keys := []string{keyZSet, keyCount}
	argv := []interface{}{objectIDStr}

	result, err := c.cancelThumbUpScript.Run(ctx, c.Cache, keys, argv...).Result()
	if err != nil {
		return 0, 0, err
	}

	arr, ok := result.([]interface{})
	if !ok || len(arr) < 2 {
		return 0, 0, fmt.Errorf("unexpected result type from lua script: %T", result)
	}

	code, ok := arr[0].(int64)
	if !ok || code == 0 {
		return 0, 0, ErrKeyNotFound
	}

	score, ok := arr[1].(string)
	if !ok {
		return 1, 0, nil
	}
	scoreInt, err := strconv.ParseInt(score, 10, 64)
	if err != nil {
		return 1, 0, nil
	}
	return 1, scoreInt, nil
}

func (c *ILikeCache) ExistZSetMember(ctx context.Context, userID uint64, objectType string, objectID uint64) (bool, error) {
	keyZSet := thumbUpListKey(userID, objectType)
	objectIDStr := strconv.FormatUint(objectID, 10)

	_, err := c.Cache.ZScore(ctx, keyZSet, objectIDStr).Result()
	if errors.Is(err, redis.Nil) {
		return false, ErrKeyNotFound
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *ILikeCache) CompensationCountDecr(ctx context.Context, objectID uint64, objectType string) error {
	keyCount := objectThumbUpCountKey(objectID, objectType)
	_, err := c.Cache.Decr(ctx, keyCount).Result()
	return err
}

func (c *ILikeCache) CompensationCountIncr(ctx context.Context, objectID uint64, objectType string) error {
	keyCount := objectThumbUpCountKey(objectID, objectType)
	_, err := c.Cache.Incr(ctx, keyCount).Result()
	return err
}

//======================================================================================================================
// 点赞列表

func (c *ILikeCache) SetLikeList(ctx context.Context, userID uint64, objectType string, objectIDs []uint64, scores []float64) error {
	if len(objectIDs) == 0 {
		return nil
	}
	if len(objectIDs) != len(scores) {
		return fmt.Errorf("objectIDs and scores length mismatch")
	}

	keyZSet := thumbUpListKey(userID, objectType)
	zs := make([]redis.Z, 0, len(objectIDs))
	for i, objectID := range objectIDs {
		zs = append(zs, redis.Z{Member: strconv.FormatUint(objectID, 10), Score: scores[i]})
	}

	pipe := c.Cache.Pipeline()
	pipe.Del(ctx, keyZSet)
	pipe.ZAdd(ctx, keyZSet, zs...)
	pipe.Expire(ctx, keyZSet, c.randomLikeListExpiration())
	_, err := pipe.Exec(ctx)
	return err
}

func (c *ILikeCache) PageQueryObjects(ctx context.Context, userID uint64, objectType string, page, size int) ([]uint64, error) {
	key := thumbUpListKey(userID, objectType)
	start := int64((page - 1) * size)
	stop := start + int64(size) - 1
	objectIDs, err := c.Cache.ZRangeArgs(ctx, redis.ZRangeArgs{Key: key, Start: start, Stop: stop, Rev: true}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	res := make([]uint64, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		parsed, parseErr := strconv.ParseUint(objectID, 10, 64)
		if parseErr != nil {
			return nil, parseErr
		}
		res = append(res, parsed)
	}
	return res, nil
}

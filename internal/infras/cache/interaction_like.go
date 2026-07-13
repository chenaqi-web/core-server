package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"backend/core-server/internal/infras/cache/scripts"
	"backend/core-server/internal/model/entity"

	"github.com/redis/go-redis/v9"
)

const (
	defaultMaxLikeSetSize      int64 = 50
	defaultLikeListExpiration        = 7 * 24 * time.Hour
	defaultLikeCountExpiration       = 7 * 24 * time.Hour
)

func thumbUpListKey(userID, objectType string) string {
	return fmt.Sprintf("like:list:%s:%s", userID, objectType)
}

func objectThumbUpCountKey(objectID, objectType string) string {
	return fmt.Sprintf("like:count:%s:%s", objectType, objectID)
}

func userThumbUpCountKey(userID, objectType string) string {
	return fmt.Sprintf("like:user:count:%s:%s", userID, objectType)
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

func (c *ILikeCache) ThumbUp(ctx context.Context, userID, objectType, objectID string, score int64) error {
	keyZSet := thumbUpListKey(userID, objectType)
	keyCount := objectThumbUpCountKey(objectID, objectType)

	keys := []string{keyZSet, keyCount}
	argv := []interface{}{
		c.maxLikeSetSize,
		score,
		objectID,
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

func (c *ILikeCache) CancelThumbUp(ctx context.Context, userID, objectType, objectID string) (int, int64, error) {
	keyZSet := thumbUpListKey(userID, objectType)
	keyCount := objectThumbUpCountKey(objectID, objectType)

	keys := []string{keyZSet, keyCount}
	argv := []interface{}{objectID}

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

func (c *ILikeCache) SetLikeList(ctx context.Context, userID, objectType string, interactions []*entity.InteractionLike) error {
	if len(interactions) == 0 {
		return nil
	}

	keyZSet := thumbUpListKey(userID, objectType)
	zs := make([]redis.Z, 0, len(interactions))

	pipe := c.Cache.Pipeline()
	pipe.Del(ctx, keyZSet)
	for _, interaction := range interactions {
		zs = append(zs, redis.Z{
			Member: interaction.ObjectID,
			Score:  float64(interaction.Version),
		})
	}

	pipe.ZAdd(ctx, keyZSet, zs...)
	pipe.Expire(ctx, keyZSet, c.randomLikeListExpiration())

	_, err := pipe.Exec(ctx)
	return err
}

func (c *ILikeCache) ExistZSetMember(ctx context.Context, userID, objectType, objectID string) (bool, error) {
	keyZSet := thumbUpListKey(userID, objectType)

	_, err := c.Cache.ZScore(ctx, keyZSet, objectID).Result()
	if errors.Is(err, redis.Nil) {
		return false, ErrKeyNotFound
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *ILikeCache) BatchExistZSetMembers(ctx context.Context, userID, objectType string, objectIDs []string) ([]string, error) {
	key := thumbUpListKey(userID, objectType)
	return batchExistsZSetMembers(ctx, c.Cache, key, objectIDs)
}

func (c *ILikeCache) PageQueryObjects(ctx context.Context, userID, objectType string, page, size int) ([]string, error) {
	key := thumbUpListKey(userID, objectType)
	objectIDs, err := c.Cache.ZRevRange(ctx, key, int64(page), int64(size)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return objectIDs, nil
}

func (c *ILikeCache) CompensationCountDecr(ctx context.Context, objectID, objectType string) error {
	keyCount := objectThumbUpCountKey(objectID, objectType)
	_, err := c.Cache.Decr(ctx, keyCount).Result()
	return err
}

func (c *ILikeCache) CompensationCountIncr(ctx context.Context, objectID, objectType string) error {
	keyCount := objectThumbUpCountKey(objectID, objectType)
	_, err := c.Cache.Incr(ctx, keyCount).Result()
	return err
}

func (c *ILikeCache) SetUserThumbUpTotalCount(ctx context.Context, userID, objectType string, count int64) error {
	key := userThumbUpCountKey(userID, objectType)
	_, err := c.Cache.Set(ctx, key, strconv.FormatInt(count, 10), c.randomLikeCountExpiration()).Result()
	return err
}

func (c *ILikeCache) SetNXObjectLikedCount(ctx context.Context, objectID, objectType string, count int64) error {
	keyCount := objectThumbUpCountKey(objectID, objectType)
	_, err := c.Cache.SetNX(ctx, keyCount, strconv.FormatInt(count, 10), c.randomLikeCountExpiration()).Result()
	return err
}

func (c *ILikeCache) SetNXObjectsLikedCountFromDB(ctx context.Context, records []*entity.InteractionCount) error {
	if len(records) == 0 {
		return nil
	}

	pipe := c.Cache.Pipeline()
	for _, record := range records {
		keyCount := objectThumbUpCountKey(record.ObjectID, record.ObjectType.String())
		pipe.SetNX(ctx, keyCount, record.Count, c.randomLikeCountExpiration())
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (c *ILikeCache) ExpireObjectLikedCount(ctx context.Context, objectType, objectID string) error {
	countKey := objectThumbUpCountKey(objectID, objectType)
	_, err := c.Cache.Expire(ctx, countKey, c.randomLikeCountExpiration()).Result()
	return err
}

func (c *ILikeCache) QueryUserLikeTotalCount(ctx context.Context, userID, objectType string) (int64, error) {
	key := userThumbUpCountKey(userID, objectType)
	total, err := c.Cache.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return -1, ErrKeyNotFound
	}
	if err != nil {
		return 0, err
	}

	count, err := strconv.ParseInt(total, 10, 64)
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (c *ILikeCache) GetObjectLikedCount(ctx context.Context, objectType, objectID string) (int64, error) {
	countKey := objectThumbUpCountKey(objectID, objectType)
	count, err := c.Cache.Get(ctx, countKey).Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrKeyNotFound
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(count, 10, 64)
}

func (c *ILikeCache) QueryObjectLikedCountInBatch(ctx context.Context, objectType string, objectIDs []string) ([]int64, []string, map[string]int, error) {
	if len(objectIDs) == 0 {
		return nil, nil, nil, fmt.Errorf("invalid empty id list")
	}

	pipe := c.Cache.Pipeline()
	cmdList := make([]*redis.StringCmd, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		countKey := objectThumbUpCountKey(objectID, objectType)
		cmd := pipe.Get(ctx, countKey)
		cmdList = append(cmdList, cmd)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, nil, nil, err
	}

	counts := make([]int64, len(objectIDs))
	missedIDs := make([]string, 0)
	missIndex := make(map[string]int)

	for i, cmd := range cmdList {
		val, err := cmd.Int64()
		if errors.Is(err, redis.Nil) {
			missedIDs = append(missedIDs, objectIDs[i])
			missIndex[objectIDs[i]] = i
			continue
		}
		if err != nil {
			return nil, nil, nil, err
		}
		counts[i] = val
	}

	return counts, missedIDs, missIndex, nil
}

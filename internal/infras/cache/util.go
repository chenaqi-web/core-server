package cache

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

func randomExpiration(base time.Duration, jitter float64) time.Duration {
	if base <= 0 {
		return 0
	}
	delta := float64(base) * jitter
	offset := (rand.Float64()*2 - 1) * delta
	return base + time.Duration(offset)
}

func batchExistsZSetMembers(ctx context.Context, client *redis.Client, key string, objectIDs []string) ([]string, error) {
	if len(objectIDs) == 0 {
		return nil, nil
	}

	pipe := client.Pipeline()
	cmds := make([]*redis.FloatCmd, len(objectIDs))
	for i, objectID := range objectIDs {
		cmds[i] = pipe.ZScore(ctx, key, objectID)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	exists := make([]string, 0, len(objectIDs))
	for i, cmd := range cmds {
		if cmd.Err() == nil {
			exists = append(exists, objectIDs[i])
		}
	}
	return exists, nil
}

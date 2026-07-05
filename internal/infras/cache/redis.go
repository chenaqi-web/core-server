package cache

import (
	"backend/core-server/internal/config"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheClient struct {
	Cache *redis.Client
}

// NewClient, 参数设置参考
// https://aws.amazon.com/cn/blogs/china/all-roads-lead-to-rome-use-go-redis-to-connect-amazon-elasticache-for-redis-cluster/
func NewClient(cfg *config.Config) *CacheClient {
	options := redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,

		// 连接池容量以及闲置连接数量
		// PoolSize: ,
		MinIdleConns: 10, // 在启动阶段创建指定数量的空闲连接，并长期维持空闲状态的连接数不少于指定数量

		// 超时设置
		DialTimeout:  5 * time.Second, // 连接建立超时时间，默认5秒
		ReadTimeout:  3 * time.Second, // 读超时，默认3秒，-1表示取消读超时
		WriteTimeout: 3 * time.Second, // 写超时，默认3秒

		// 命令执行失败时的重试策略
		MaxRetries:      3,                      // 命令执行失败时，最多重试次数，默认为0时不重试
		MinRetryBackoff: 8 * time.Millisecond,   // 重试间隔时间的下限，默认8毫秒，-1表示取消间隔
		MaxRetryBackoff: 512 * time.Millisecond, // 重试间隔时间的上限，默认512毫秒，-1表示取消间隔
	}

	rdb := redis.NewClient(&options)

	return &CacheClient{
		Cache: rdb,
	}
}

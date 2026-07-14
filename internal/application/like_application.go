package application

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"backend/core-server/internal/config"
	"backend/core-server/internal/domain"
	"backend/core-server/internal/infras/cache"
	"backend/core-server/internal/infras/mq/kafka"
	"backend/core-server/internal/model/entity"
	"backend/core-server/internal/model/enum"
	"backend/core-server/internal/model/event"

	"github.com/avast/retry-go"
)

type LikeService struct {
	cfg      *config.Config
	repo     domain.LikeDomain
	cache    *cache.ILikeCache
	producer *kafka.SyncProducer
}

func NewLikeService(
	repo domain.LikeDomain,
	likeCache *cache.ILikeCache,
	producer *kafka.SyncProducer,
	cfg *config.Config,
) (*LikeService, error) {
	return &LikeService{
		cfg:      cfg,
		repo:     repo,
		cache:    likeCache,
		producer: producer,
	}, nil
}

// =====================================================================================================================
// 点赞操作

func (s *LikeService) HasThumbUp(ctx context.Context, userID, objectType, objectID string) (bool, error) {
	// 1. 首先判断是否点赞(在zset里面)
	exist, err := s.cache.ExistZSetMember(ctx, userID, objectType, objectID)
	if err != nil && !errors.Is(err, cache.ErrKeyNotFound) {
		// todo log 降级
	}
	if exist {
		return true, nil
	}

	// 2. 如果有问题，降级查数据库
	interaction, err := s.repo.QueryWithCondition(
		ctx,
		userID,
		objectType,
		objectID,
		entity.LikeStatusTypeThumbUp.String(),
	)
	if err != nil {
		return false, err
	}
	return interaction != nil, nil
}

func (s *LikeService) ThumbUp(ctx context.Context, userID, objectType, objectID, objectOwnerID string) error {
	// 1. 先查询缓存，判断是否点赞
	exists, err := s.HasThumbUp(ctx, userID, objectType, objectID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyLiked
	}

	// 2. 不存在（a.有记录，但是状态为nothing b.无记录）
	score := time.Now().UnixMicro()
	if err := s.cache.ThumbUp(ctx, userID, objectType, objectID, score); err != nil {
		return err
	}

	// 3，提交任务（异步落库）
	payload := &event.EventUserThumbUp{
		Timestamp:     score,
		UserID:        userID,
		ObjectType:    objectType,
		ObjectID:      objectID,
		ObjectOwnerID: objectOwnerID,
		Status:        entity.LikeStatusTypeThumbUp.String(),
	}
	eventBytes, _ := json.Marshal(payload)

	// 如果说发送失败，则补偿
	if err := s.sendMessage(&event.Message{
		UserID:    userID,
		EventType: enum.MessageEventTypeUserThumbUp.String(),
		Body:      eventBytes,
	}); err != nil {
		_, _, _ = s.cache.CancelThumbUp(ctx, userID, objectType, objectID)
		return err
	}
	return nil
}

func (s *LikeService) CancelThumbUp(ctx context.Context, userID, objectType, objectID, objectOwnerID string) error {
	// 1. 先查询缓存，有就删除
	// result 就返回两个值 0和1，0表示没有，1表示有且成功删除
	result, score, err := s.cache.CancelThumbUp(ctx, userID, objectType, objectID)
	if err != nil && !errors.Is(err, cache.ErrKeyNotFound) {
		return err
	}

	// 当缓存没有，查数据库(说明1.缓存过期 2.冷数据)
	var res *entity.InteractionLike
	if result == 0 {
		res, err = s.repo.QueryWithCondition(ctx, userID, objectType, objectID, entity.LikeStatusTypeThumbUp.String())
		if err != nil {
			return err
		}
		if res == nil {
			return ErrLikeNotExists
		}
	}

	// 数据库有或者缓存有，异步发送去删除,res一定不为空
	// 这里只要result ！= 0就表示缓存有，这个CancelThumbUp函数只返回0和1
	if res != nil || result == 1 {
		payload := &event.EventUserCancelThumbUp{
			Timestamp:        time.Now().UnixMicro(),
			UserID:           userID,
			ObjectType:       objectType,
			ObjectID:         objectID,
			ObjectOwnerID:    objectOwnerID,
			IsDeletedInCache: result,
		}
		body, _ := json.Marshal(payload)

		if err := s.sendMessage(&event.Message{
			UserID:    userID,
			EventType: enum.MessageEventTypeUserCancelThumbUp.String(),
			Body:      body,
		}); err != nil {
			// 只有当出错误删，才进行补偿
			if result == 1 {
				_ = s.cache.ThumbUp(ctx, userID, objectType, objectID, score)
			}
			return err
		}
	}
	return nil
}

func (s *LikeService) sendMessage(msg *event.Message) error {
	// 1.对整个msg编码
	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// 2.拿到对应topic
	topic, _ := s.cfg.Kafka.LikeTopicName()

	// 3.发送消息到mq，重试3次
	err = retry.Do(func() error {
		return s.producer.SendMessage(topic, msg.UserID, value)
	},
		retry.Attempts(3),
		retry.MaxDelay(10*time.Second),
		retry.DelayType(retry.BackOffDelay),
	)
	if err != nil {
		return err
	}
	return nil
}

// =====================================================================================================================
// 点赞列表方面

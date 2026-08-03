package jobdbsync

import (
	"context"
	"core-server/internal/model/entity"
	"core-server/internal/model/enum"

	"core-server/internal/model/event"

	"go.uber.org/zap"
)

func (c *MessageQueueConsumer) handleUserLike(ctx context.Context, msg *event.EventUserThumbUp) error {
	ent := &entity.InteractionLike{
		UserID:     msg.UserID,
		ObjectType: enum.ParseObjectType(msg.ObjectType),
		ObjectID:   msg.ObjectID,
		Status:     entity.ParseLikeStatusType(msg.Status),
		Version:    msg.Timestamp,
	}
	// upsert
	err := c.likeRepo.WithTransaction(ctx, func(ctx context.Context) error {
		res, err := c.likeRepo.Upsert(ctx, ent)
		if err != nil {
			_, _, cacheErr := c.likeCache.CancelThumbUp(ctx, msg.UserID, msg.ObjectType, msg.ObjectID)
			if cacheErr != nil {
				c.logger.Error("cache cancel thumb up failed", zap.String("user_id", msg.UserID), zap.Error(cacheErr))
			}
			return err
		}
		// 如果等于 0则无需更新计数，直接返回即可
		if res == 0 {
			return nil
		}

		// 计数聚合器，后续统一写入(前提是成功写入了)
		c.CountAggregator.Push(ctx, enum.InteractionTypeLike.String(), msg.ObjectType, msg.ObjectID)
		return nil
	})
	if err != nil {
		c.logger.Error("handle user like failed", zap.Any("msg", msg), zap.Error(err))
		return err
	}
	return nil
}

func (c *MessageQueueConsumer) handleUserCancelThumbUp(ctx context.Context, msg *event.EventUserCancelThumbUp) error {
	ent := &entity.InteractionLike{
		UserID:     msg.UserID,
		ObjectType: enum.ParseObjectType(msg.ObjectType),
		ObjectID:   msg.ObjectID,
		Version:    msg.Timestamp,
		Status:     entity.LikeStatusTypeNothing,
	}

	mark := msg.IsDeletedInCache
	affected := 0
	condition := entity.LikeStatusTypeThumbUp.String()

	err := c.likeRepo.WithTransaction(ctx, func(ctx context.Context) error {
		var err error
		// 条件更新，且它的状态必须是点赞
		affected, err = c.likeRepo.UpdateWithCondition(ctx, condition, ent)
		if err != nil {
			return err
		}
		if affected == 1 {
			c.CountAggregator.Pop(ctx, enum.InteractionTypeLike.String(), msg.ObjectType, msg.ObjectID)
		}
		return nil
	})
	if err != nil {
		c.logger.Error("handle user cancel like failed", zap.Any("msg", msg), zap.Error(err))
		return err
	}

	// 缓存没有，数据库有，需要处理缓存 -1
	if affected == 1 && mark == 0 {
		if err := c.likeCache.CompensationCountDecr(ctx, msg.ObjectID, msg.ObjectType); err != nil {
			return err
		}
	}
	return nil
}

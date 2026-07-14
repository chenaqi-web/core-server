package jobdbsync

import (
	"backend/core-server/internal/model/entity"
	"backend/core-server/internal/model/enum"
	"context"

	"backend/core-server/internal/model/event"
)

func (c *MessageQueueConsumer) handleUserLike(ctx context.Context, msg *event.EventUserThumbUp) error {
	ent := &entity.InteractionLike{
		UserID:        msg.UserID,
		ObjectType:    enum.ParseObjectType(msg.ObjectType),
		ObjectID:      msg.ObjectID,
		ObjectOwnerID: msg.ObjectOwnerID,
		Status:        entity.ParseLikeStatusType(msg.Status),
		Version:       msg.Timestamp,
	}

	err := c.likeRepo.WithTransaction(ctx, func(ctx context.Context) error {
		res, err := c.likeRepo.Upsert(ctx, ent)
		if err != nil {
			_, _, cacheErr := c.likeCache.CancelThumbUp(ctx, msg.UserID, msg.ObjectType, msg.ObjectID)
			if cacheErr != nil {
				c.logger.Error("cache cancel thumb up failed", "user_id", msg.UserID, "err", cacheErr)
			}
			return err
		}
		if res == 0 {
			return nil
		}

		c.likeCountAggregator.Push(ctx, enum.InteractionTypeLike.String(), msg.ObjectType, msg.ObjectID)
		return nil
	})
	if err != nil {
		c.logger.Error("handle user like failed", "msg", msg, "err", err)
		return err
	}
	return nil
}

func (c *MessageQueueConsumer) handleUserCancelThumbUp(ctx context.Context, msg *event.EventUserCancelThumbUp) error {
	ent := &entity.InteractionLike{
		UserID:        msg.UserID,
		ObjectType:    enum.ParseObjectType(msg.ObjectType),
		ObjectID:      msg.ObjectID,
		ObjectOwnerID: msg.ObjectOwnerID,
		Version:       msg.Timestamp,
		Status:        entity.LikeStatusTypeNothing,
	}

	mark := msg.IsDeletedInCache
	affected := 0
	condition := entity.LikeStatusTypeThumbUp.String()

	err := c.likeRepo.WithTransaction(ctx, func(ctx context.Context) error {
		var err error
		affected, err = c.likeRepo.UpdateWithCondition(ctx, condition, ent)
		if err != nil {
			return err
		}
		if affected == 1 {
			c.likeCountAggregator.Pop(ctx, enum.InteractionTypeLike.String(), msg.ObjectType, msg.ObjectID)
		}
		return nil
	})
	if err != nil {
		c.logger.Error("handle user cancel like failed", "msg", msg, "err", err)
		return err
	}

	if affected == 1 && mark == 0 {
		if err := c.likeCache.CompensationCountDecr(ctx, msg.ObjectID, msg.ObjectType); err != nil {
			return err
		}
	}
	return nil
}

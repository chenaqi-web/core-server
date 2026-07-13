package jobdbsync

import (
	"context"
	"log/slog"

	"backend/core-server/internal/domain"
	"backend/core-server/internal/infras/cache"
	jobaggregator "backend/core-server/internal/jobs/job-aggregator"
	"backend/core-server/internal/model/entity"
	"backend/core-server/internal/model/enum"
	"backend/core-server/internal/model/event"
)

type LikeHandlerDependencies struct {
	LikeRepo     domain.LikeDomain
	LikeCache    *cache.ILikeCache
	LikeCountAgg *jobaggregator.ObjectCountAggregator
	Logger       *slog.Logger
}

func HandleUserLike(ctx context.Context, msg *event.EventUserThumbUp, deps *LikeHandlerDependencies) error {
	ent := &entity.InteractionLike{
		UserID:        msg.UserID,
		ObjectType:    enum.ParseObjectType(msg.ObjectType),
		ObjectID:      msg.ObjectID,
		ObjectOwnerID: msg.ObjectOwnerID,
		Status:        entity.ParseLikeStatusType(msg.Status),
		Version:       msg.Timestamp,
	}

	err := deps.LikeRepo.WithTransaction(ctx, func(ctx context.Context) error {
		res, err := deps.LikeRepo.Upsert(ctx, ent)
		if err != nil {
			_, _, cacheErr := deps.LikeCache.CancelThumbUp(ctx, msg.UserID, msg.ObjectType, msg.ObjectID)
			if cacheErr != nil {
				deps.Logger.Error("cache cancel thumb up failed", "user_id", msg.UserID, "err", cacheErr)
			}
			return err
		}
		if res == 0 {
			return nil
		}

		deps.LikeCountAgg.Push(ctx, enum.InteractionTypeLike.String(), msg.ObjectType, msg.ObjectID)
		return nil
	})
	if err != nil {
		deps.Logger.Error("handle user like failed", "msg", msg, "err", err)
		return err
	}
	return nil
}

func HandleUserCancelLike(ctx context.Context, msg *event.EventUserCancelThumbUp, deps *LikeHandlerDependencies) error {
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

	err := deps.LikeRepo.WithTransaction(ctx, func(ctx context.Context) error {
		var err error
		affected, err = deps.LikeRepo.UpdateWithCondition(ctx, condition, ent)
		if err != nil {
			return err
		}
		if affected == 1 {
			deps.LikeCountAgg.Pop(ctx, enum.InteractionTypeLike.String(), msg.ObjectType, msg.ObjectID)
		}
		return nil
	})
	if err != nil {
		deps.Logger.Error("handle user cancel like failed", "msg", msg, "err", err)
		return err
	}

	if affected == 1 && mark == 0 {
		if err := deps.LikeCache.CompensationCountDecr(ctx, msg.ObjectID, msg.ObjectType); err != nil {
			return err
		}
	}
	return nil
}

func (c *MessageQueueConsumer) NewLikeHandlerDependencies() *LikeHandlerDependencies {
	return &LikeHandlerDependencies{
		LikeRepo:     c.likeRepo,
		LikeCache:    c.likeCache,
		LikeCountAgg: c.likeCountAggregator,
		Logger:       c.logger,
	}
}

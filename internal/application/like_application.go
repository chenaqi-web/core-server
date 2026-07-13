package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"backend/core-server/internal/config"
	"backend/core-server/internal/domain"
	"backend/core-server/internal/infras/cache"
	"backend/core-server/internal/infras/mq/kafka"
	"backend/core-server/internal/model/entity"
	"backend/core-server/internal/model/enum"
	"backend/core-server/internal/model/event"
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

func (s *LikeService) HasThumbUp(ctx context.Context, userID, objectType, objectID string) (bool, error) {
	exist, err := s.cache.ExistZSetMember(ctx, userID, objectType, objectID)
	if err != nil && !errors.Is(err, cache.ErrKeyNotFound) {
		return false, err
	}
	if exist {
		return true, nil
	}

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
	exists, err := s.HasThumbUp(ctx, userID, objectType, objectID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyLiked
	}

	score := time.Now().UnixMicro()
	if err := s.cache.ThumbUp(ctx, userID, objectType, objectID, score); err != nil {
		return err
	}

	payload := &event.EventUserThumbUp{
		Timestamp:     score,
		UserID:        userID,
		ObjectType:    objectType,
		ObjectID:      objectID,
		ObjectOwnerID: objectOwnerID,
		Status:        entity.LikeStatusTypeThumbUp.String(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		_, _, _ = s.cache.CancelThumbUp(ctx, userID, objectType, objectID)
		return err
	}

	if err := s.sendMessage(&event.Message{
		UserID:    userID,
		EventType: enum.MessageEventTypeUserThumbUp.String(),
		Body:      body,
	}); err != nil {
		_, _, _ = s.cache.CancelThumbUp(ctx, userID, objectType, objectID)
		return err
	}
	return nil
}

func (s *LikeService) CancelThumbUp(ctx context.Context, userID, objectType, objectID, objectOwnerID string) error {
	result, score, err := s.cache.CancelThumbUp(ctx, userID, objectType, objectID)
	if err != nil && !errors.Is(err, cache.ErrKeyNotFound) {
		return err
	}

	var res *entity.InteractionLike
	if result == 0 {
		res, err = s.repo.QueryWithCondition(
			ctx,
			userID,
			objectType,
			objectID,
			entity.LikeStatusTypeThumbUp.String(),
		)
		if err != nil {
			return err
		}
		if res == nil {
			return ErrLikeNotExists
		}
	}

	if res != nil || result == 1 {
		ownerID := objectOwnerID
		if ownerID == "" && res != nil {
			ownerID = res.ObjectOwnerID
		}

		payload := &event.EventUserCancelThumbUp{
			Timestamp:        time.Now().UnixMicro(),
			UserID:           userID,
			ObjectType:       objectType,
			ObjectID:         objectID,
			ObjectOwnerID:    ownerID,
			IsDeletedInCache: result,
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		if err := s.sendMessage(&event.Message{
			UserID:    userID,
			EventType: enum.MessageEventTypeUserCancelThumbUp.String(),
			Body:      body,
		}); err != nil {
			if result == 1 {
				_ = s.cache.ThumbUp(ctx, userID, objectType, objectID, score)
			}
			return err
		}
	}
	return nil
}

func (s *LikeService) sendMessage(msg *event.Message) error {
	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	topic, err := s.cfg.Kafka.LikeTopicName()
	if err != nil {
		return fmt.Errorf("parse like topic: %w", err)
	}

	return s.producer.SendMessage(topic, msg.UserID, value)
}

// =====================================================================================================================

package application

import (
	"context"
	"core-server/internal/infras/clog"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"core-server/internal/config"
	"core-server/internal/domain"
	"core-server/internal/infras/cache"
	"core-server/internal/infras/mq/kafka"
	"core-server/internal/model/aggregate"
	"core-server/internal/model/entity"
	"core-server/internal/model/enum"
	"core-server/internal/model/event"

	"github.com/avast/retry-go"
	"go.uber.org/zap"
)

type LikeService struct {
	cfg         *config.Config
	log         *clog.Log
	producer    *kafka.SyncProducer
	repo        domain.LikeRepoDomain
	cache       domain.LikeCacheDomain
	articleRepo domain.ArticleRepoDomain
	userRepo    domain.UserRepoDomain
}

func NewLikeService(
	log *clog.Log,
	repo domain.LikeRepoDomain,
	likeCache domain.LikeCacheDomain,
	producer *kafka.SyncProducer,
	articleRepo domain.ArticleRepoDomain,
	userRepo domain.UserRepoDomain,
	cfg *config.Config,
) (*LikeService, error) {
	return &LikeService{
		cfg:         cfg,
		repo:        repo,
		cache:       likeCache,
		producer:    producer,
		articleRepo: articleRepo,
		userRepo:    userRepo,
		log:         log,
	}, nil
}

// =====================================================================================================================
// 点赞状态

func (s *LikeService) HasThumbUp(ctx context.Context, userID uint64, objectType string, objectID uint64) (bool, error) {
	// 1. 首先判断是否点赞(在zset里面)
	exist, err := s.cache.ExistZSetMember(ctx, userID, objectType, objectID)
	if err != nil && !errors.Is(err, cache.ErrKeyNotFound) {
		s.log.Error("Error checking if thumbnail exist", zap.Error(err))
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

func (s *LikeService) BatchHasThumbUp(ctx context.Context, userID uint64, objectType string, objectIDs []uint64) (map[uint64]bool, error) {
	statuses := make(map[uint64]bool, len(objectIDs))
	for _, objectID := range objectIDs {
		liked, err := s.HasThumbUp(ctx, userID, objectType, objectID)
		if err != nil {
			return nil, err
		}
		statuses[objectID] = liked
	}
	return statuses, nil
}

// =====================================================================================================================
// 点赞操作

func (s *LikeService) ThumbUp(ctx context.Context, userID uint64, objectType string, objectID uint64) error {
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
		Timestamp:  score,
		UserID:     userID,
		ObjectType: objectType,
		ObjectID:   objectID,
		Status:     entity.LikeStatusTypeThumbUp.String(),
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

func (s *LikeService) CancelThumbUp(ctx context.Context, userID uint64, objectType string, objectID uint64) error {
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
			return nil
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
		return s.producer.SendMessage(topic, strconv.FormatUint(msg.UserID, 10), value)
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

func (s *LikeService) UserLikeList(ctx context.Context, userID uint64, objectType string, page, pageSize int) ([]*aggregate.ArticleAggregate, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 1) 直接从 user 表查询用户点赞总数，不再走缓存
	total, err := s.userRepo.GetLikeCount(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	offset := (page - 1) * pageSize
	if int64(offset) >= total {
		return nil, total, nil
	}

	// 2) 先查 zset 热数据
	cachedIDs, cacheErr := s.cache.PageQueryObjects(ctx, userID, objectType, page, pageSize)
	if cacheErr == nil && len(cachedIDs) == pageSize {
		articles, err := s.loadArticlesByIDs(ctx, cachedIDs)
		if err != nil {
			return nil, 0, err
		}
		return articles, total, nil
	}

	// 3) 缓存没有命中，或者最后一页不足 pageSize，或者超过 zset 大小，则查数据库
	likes, err := s.repo.PageQueryLikeObjects(ctx, userID, objectType, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	ids := make([]uint64, 0, len(likes))
	scores := make([]float64, 0, len(likes))
	for _, like := range likes {
		ids = append(ids, like.ObjectID)
		scores = append(scores, float64(like.Version))
	}

	// 4) 只要本次分页范围仍在热数据预算内，就回填 zset
	if err := s.cache.SetLikeList(ctx, userID, objectType, ids, scores); err != nil {
		s.log.Error("set like list cache failed", zap.Error(err))
	}

	articles, err := s.loadArticlesByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	return articles, total, nil
}

// todo 这个加载数据可能后续在改，可以换成批量处理，暂时够用

func (s *LikeService) loadArticlesByIDs(ctx context.Context, ids []uint64) ([]*aggregate.ArticleAggregate, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	articles := make([]*aggregate.ArticleAggregate, 0, len(ids))
	for _, id := range ids {
		article, err := s.articleRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if article == nil {
			continue
		}
		author, err := s.userRepo.GetByID(ctx, article.AuthorID)
		if err != nil {
			return nil, err
		}
		articles = append(articles, aggregate.NewArticleAggregate(article, author))
	}
	return articles, nil
}

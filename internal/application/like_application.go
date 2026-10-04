package application

import (
	"context"
	"core-server/internal/infras/clog"
	"time"

	"core-server/internal/config"
	"core-server/internal/domain"
	"core-server/internal/model/aggregate"
	"core-server/internal/model/entity"
	"core-server/internal/model/enum"

	"github.com/avast/retry-go"
)

type LikeService struct {
	cfg          *config.Config
	log          *clog.Log
	repo         domain.LikeRepoDomain
	articleRepo  domain.ArticleRepoDomain
	userRepo     domain.UserRepoDomain
	countService *CountService
}

func NewLikeService(
	log *clog.Log,
	repo domain.LikeRepoDomain,
	articleRepo domain.ArticleRepoDomain,
	userRepo domain.UserRepoDomain,
	countService *CountService,
	cfg *config.Config,
) (*LikeService, error) {
	return &LikeService{
		cfg:          cfg,
		repo:         repo,
		articleRepo:  articleRepo,
		userRepo:     userRepo,
		countService: countService,
		log:          log,
	}, nil
}

// =====================================================================================================================
// 点赞状态

func (s *LikeService) HasThumbUp(ctx context.Context, userID uint64, objectType string, objectID uint64) (bool, error) {
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

// =====================================================================================================================
// 点赞操作

func (s *LikeService) ThumbUp(ctx context.Context, userID uint64, objectType string, objectID uint64) error {
	// 首先判断点赞是否已经存在
	exists, err := s.HasThumbUp(ctx, userID, objectType, objectID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyLiked
	}

	like := &entity.InteractionLike{
		UserID:     userID,
		ObjectType: enum.ParseObjectType(objectType),
		ObjectID:   objectID,
		Status:     entity.LikeStatusTypeThumbUp,
		Version:    time.Now().UnixMicro(),
	}

	mark := true

	// 开始操作点赞
	err = s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		affected, err := s.repo.Upsert(ctx, like)
		if err != nil {
			return err
		}
		// upsert 0表示无变化或未生效，1表示成功，2表示更新了原有的点赞状态（这里1，2都是1）
		if affected == 0 {
			// 无影响，则计数不需要+1
			mark = false
			return nil
		}

		if err := s.countService.UpdateObjectTypeWithInteractionType(ctx, objectType, enum.InteractionTypeLike.String(), objectID, 1); err != nil {
			return err
		}

		// 更新用户点赞数
		if err := s.userRepo.UpdateLikeCount(ctx, userID, 1); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("Error in ThumbUp transaction")
		return err
	}

	// 没问题异步增加作者的获赞数
	if mark {
		s.asyncUpdateObjectAuthorReceiveLikeCount(objectType, objectID, 1)
	}

	return nil
}

func (s *LikeService) CancelThumbUp(ctx context.Context, userID uint64, objectType string, objectID uint64) error {
	exists, err := s.HasThumbUp(ctx, userID, objectType, objectID)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	like := &entity.InteractionLike{
		UserID:     userID,
		ObjectType: enum.ParseObjectType(objectType),
		ObjectID:   objectID,
		Status:     entity.LikeStatusTypeNothing,
		Version:    time.Now().UnixMicro(),
	}

	mark := true

	err = s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		affected, err := s.repo.UpdateWithCondition(ctx, entity.LikeStatusTypeThumbUp.String(), like)
		if err != nil {
			return err
		}
		if affected == 0 {
			// 无影响，则计数不需要-1
			mark = false
			return nil
		}
		if err := s.countService.UpdateObjectTypeWithInteractionType(ctx, objectType, enum.InteractionTypeLike.String(), objectID, -1); err != nil {
			return err
		}
		if err := s.userRepo.UpdateLikeCount(ctx, userID, -1); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("Error in CancelThumbUp transaction")
		return err
	}

	// 没问题异步减少作者的获赞数
	if mark {
		s.asyncUpdateObjectAuthorReceiveLikeCount(objectType, objectID, -1)
	}

	return nil
}

func (s *LikeService) asyncUpdateObjectAuthorReceiveLikeCount(objectType string, objectID uint64, delta int) {
	go func() {
		syncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := retry.Do(func() error {
			authorID, err := s.getObjectAuthorID(syncCtx, objectType, objectID)
			if err != nil {
				return err
			}
			if authorID == 0 {
				return nil
			}
			return s.userRepo.UpdateReceiveLikeCount(syncCtx, authorID, int64(delta))
		},
			retry.Attempts(3),
			retry.MaxDelay(10*time.Second),
			retry.DelayType(retry.BackOffDelay),
		)
		if err != nil {
			s.log.Error("Error in asyncUpdateObjectAuthorReceiveLikeCount")
			return
		}
	}()
}

func (s *LikeService) getObjectAuthorID(ctx context.Context, objectType string, objectID uint64) (uint64, error) {
	switch enum.ParseObjectType(objectType) {
	case enum.ObjectTypeArticle:
		article, err := s.articleRepo.GetByID(ctx, objectID)
		if err != nil {
			return 0, err
		}
		if article == nil || article.Article == nil {
			return 0, ErrArticleNotFound
		}
		return article.Article.AuthorID, nil
	default:
		// 后续新增 objectType 时，在这里补充对应对象的作者查询。
		return 0, nil
	}
}

// =====================================================================================================================
// 点赞列表方面

func (s *LikeService) UserLikeList(ctx context.Context, userID uint64, objectType string, page, pageSize int) ([]*aggregate.ArticleAggregate, error) {
	offset := (page - 1) * pageSize

	likes, err := s.repo.PageQueryLikeObjects(ctx, userID, objectType, offset, pageSize)
	if err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, len(likes))
	for _, like := range likes {
		ids = append(ids, like.ObjectID)
	}

	articles, err := s.articleRepo.ListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return articles, nil
}

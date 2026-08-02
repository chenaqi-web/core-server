package application

import (
	"context"
	"errors"

	"core-server/internal/domain"
	"core-server/internal/model/entity"
)

const (
	defaultPageSize   = 10
	previewReplyLimit = 3
)

var (
	ErrAlreadyLiked  = errors.New("already liked")
	ErrLikeNotExists = errors.New("like not exists")

	ErrCategoryTypeNotFound = errors.New("category type not found")
	ErrCategoryNotFound     = errors.New("category not found")
	ErrArticleNotFound      = errors.New("article not found")

	ErrCommentNotFound = errors.New("comment not found")
	ErrCommentInvalid  = errors.New("comment content invalid")
)

func Page(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func Size(size int) int {
	if size <= 0 {
		return defaultPageSize
	}
	return size
}

func CollectCommentUserIDs(comments []*entity.Comment) []uint64 {
	if len(comments) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(comments))
	seen := make(map[uint64]struct{})
	for _, c := range comments {
		if _, ok := seen[c.UserID]; ok {
			continue
		}
		seen[c.UserID] = struct{}{}
		ids = append(ids, c.UserID)
	}
	return ids
}

func CollectArticleAuthorIDs(articles []*entity.Article) []uint64 {
	if len(articles) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(articles))
	seen := make(map[uint64]struct{})
	for _, a := range articles {
		if _, ok := seen[a.AuthorID]; ok {
			continue
		}
		seen[a.AuthorID] = struct{}{}
		ids = append(ids, a.AuthorID)
	}
	return ids
}

func LoadUserMap(ctx context.Context, userRepo domain.UserRepoDomain, userIDs []uint64) (map[uint64]*entity.User, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	users, err := userRepo.ListByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	userMap := make(map[uint64]*entity.User, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}
	return userMap, nil
}

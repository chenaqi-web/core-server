package repov2

import (
	"core-server/internal/infras/repov2/ent"
	"core-server/internal/model/entity"
	"database/sql"
)

func toEntityUser(node *ent.User) *entity.User {
	if node == nil {
		return nil
	}
	value := &entity.User{
		ID:        node.ID,
		CreatedAt: node.CreatedAt,
		UpdatedAt: node.UpdatedAt,
		Name:      node.Name,
		Password:  node.Password,
		Phone:     node.Phone,
		Avatar:    node.Avatar,
		Email:     node.Email,
		Role:      node.Role,
		Sex:       node.Sex,
		Status:    node.Status,
	}
	if node.Birthday != nil {
		value.Birthday = *node.Birthday
	}
	if node.DeletedAt != nil {
		value.DeletedAt = sql.NullTime{Time: *node.DeletedAt, Valid: true}
	}
	return value
}

func toEntityUsers(nodes []*ent.User) []*entity.User {
	users := make([]*entity.User, 0, len(nodes))
	for _, node := range nodes {
		users = append(users, toEntityUser(node))
	}
	return users
}

func toEntityUserStat(stat *ent.UserStat) *entity.UserStat {
	if stat == nil {
		return nil
	}
	return &entity.UserStat{
		UserID:            stat.UserID,
		ArticleCount:      stat.ArticleCount,
		FollowersCount:    stat.FollowersCount,
		FollowingCount:    stat.FollowingCount,
		LikeCount:         stat.LikeCount,
		ReceiveLikeCount:  stat.ReceiveLikeCount,
		FavorCount:        stat.FavorCount,
		ReceiveFavorCount: stat.ReceiveFavorCount,
	}
}

func toEntityCategory(node *ent.Category) *entity.Category {
	if node == nil {
		return nil
	}
	value := &entity.Category{
		ID:        node.ID,
		CreatedAt: node.CreatedAt,
		UpdatedAt: node.UpdatedAt,
		ParentID:  node.ParentID,
		Name:      node.Name,
	}
	return value
}

func toEntityCategories(nodes []*ent.Category) []*entity.Category {
	items := make([]*entity.Category, 0, len(nodes))
	for _, node := range nodes {
		items = append(items, toEntityCategory(node))
	}
	return items
}

func toEntityArticle(node *ent.Article) *entity.Article {
	if node == nil {
		return nil
	}
	value := &entity.Article{
		ID:           node.ID,
		CreatedAt:    node.CreatedAt,
		UpdatedAt:    node.UpdatedAt,
		Title:        node.Title,
		Summary:      node.Summary,
		Content:      node.Content,
		CoverImage:   node.CoverImage,
		AuthorID:     node.AuthorID,
		CategoryID:   node.CategoryID,
		IsTop:        node.IsTop,
		IsPublished:  node.IsPublished,
		Visibility:   uint32(node.Visibility),
		ViewCount:    uint64(node.ViewCount),
		LikeCount:    uint64(node.LikeCount),
		CommentCount: uint64(node.CommentCount),
	}
	if node.DeletedAt != nil {
		value.DeletedAt = sql.NullTime{Time: *node.DeletedAt, Valid: true}
	}
	if node.PublishedAt != nil {
		value.PublishedAt = sql.NullTime{Time: *node.PublishedAt, Valid: true}
	}
	return value
}

func toEntityArticles(nodes []*ent.Article) []*entity.Article {
	items := make([]*entity.Article, 0, len(nodes))
	for _, node := range nodes {
		items = append(items, toEntityArticle(node))
	}
	return items
}

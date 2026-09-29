package repov2

import (
	"core-server/internal/infras/repov2/ent"
	"core-server/internal/model/aggregate"
	"core-server/internal/model/entity"
	"core-server/internal/model/enum"
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
		Role:      enum.ParseUserRole(node.Role.String()),
		Sex:       enum.ParseUserSex(node.Sex.String()),
		Status:    enum.ParseUserStatus(node.Status.String()),
		Signature: node.Signature,
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
		ViewCount:    uint64(node.ViewCount),
		LikeCount:    uint64(node.LikeCount),
		FavorCount:   uint64(node.FavorCount),
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

func toEntityComment(node *ent.Comment) *entity.Comment {
	if node == nil {
		return nil
	}
	value := &entity.Comment{
		ID:         node.ID,
		ArticleID:  node.ArticleID,
		UserID:     node.UserID,
		ParentID:   node.ParentID,
		RootID:     node.RootID,
		ReplyToID:  node.ReplyToID,
		Content:    node.Content,
		LikeCount:  node.LikeCount,
		ChildCount: node.ChildCount,
		CreatedAt:  node.CreatedAt,
	}
	if node.DeletedAt != nil {
		value.DeletedAt = sql.NullTime{Time: *node.DeletedAt, Valid: true}
	}
	return value
}

func toEntityComments(nodes []*ent.Comment) []*entity.Comment {
	items := make([]*entity.Comment, 0, len(nodes))
	for _, node := range nodes {
		items = append(items, toEntityComment(node))
	}
	return items
}

func toEntityCommentAggregates(nodes []*ent.Comment) []*aggregate.CommentAggregate {
	items := make([]*aggregate.CommentAggregate, 0, len(nodes))
	for _, node := range nodes {
		items = append(items, aggregate.NewCommentAggregate(toEntityComment(node), toEntityUser(node.Edges.User)))
	}
	return items
}

func toEntityInteractionLike(node *ent.InteractionLike) *entity.InteractionLike {
	if node == nil {
		return nil
	}
	return &entity.InteractionLike{
		ID:         node.ID,
		CreatedAt:  node.CreatedAt,
		UpdatedAt:  node.UpdatedAt,
		UserID:     node.UserID,
		ObjectType: enum.ParseObjectType(node.ObjectType.String()),
		ObjectID:   node.ObjectID,
		Status:     entity.ParseLikeStatusType(node.Status.String()),
		Version:    node.Version,
	}
}

func toEntityInteractionLikes(nodes []*ent.InteractionLike) []*entity.InteractionLike {
	items := make([]*entity.InteractionLike, 0, len(nodes))
	for _, node := range nodes {
		items = append(items, toEntityInteractionLike(node))
	}
	return items
}

func toEntityInteractionCounts(nodes []*ent.InteractionCount) []*entity.InteractionCount {
	items := make([]*entity.InteractionCount, 0, len(nodes))
	for _, node := range nodes {
		items = append(items, &entity.InteractionCount{
			ID:              node.ID,
			CreatedAt:       node.CreatedAt,
			UpdatedAt:       node.UpdatedAt,
			ObjectType:      enum.ParseObjectType(node.ObjectType.String()),
			ObjectID:        node.ObjectID,
			InteractionType: enum.ParseInteractionType(node.InteractionType.String()),
			Count:           int64(node.Count),
		})
	}
	return items
}

func toEntityArticleAggregates(nodes []*ent.Article) []*aggregate.ArticleAggregate {
	items := make([]*aggregate.ArticleAggregate, 0, len(nodes))
	for _, node := range nodes {
		items = append(items, toEntityArticleAggregate(node))
	}

	return items
}

func toEntityArticleAggregate(node *ent.Article) *aggregate.ArticleAggregate {
	author := node.Edges.User
	cate := node.Edges.Category
	return &aggregate.ArticleAggregate{
		Article: toEntityArticle(node),
		Author:  toEntityUser(author),
		Cate:    toEntityCategory(cate),
	}
}

// Package aggregate holds domain aggregates.
// An aggregate combines entities and value objects; only the root entity is externally referenceable.
package aggregate

import (
	"backend/core-server/internal/model/entity"
)

// ArticleAggregate 文章展示聚合根：文章本体 + 作者（user 表，通过 author_id 关联）。
// 分类信息可在展示层单独查询后补充。
type ArticleAggregate struct {
	Article *entity.Article
	Author  *entity.User
}

// NewArticleAggregate 新建聚合
func NewArticleAggregate(article *entity.Article, author *entity.User) *ArticleAggregate {
	return &ArticleAggregate{
		Article: article,
		Author:  author,
	}
}

// Package aggregate holds domain aggregates.
// An aggregate combines entities and value objects; only the root entity is externally referenceable.
package aggregate

import (
	"core-server/internal/model/entity"
)

// ArticleAggregate 文章展示聚合根：文章本体 + 作者（user 表，通过 author_id 关联）。
// 分类信息可在展示层单独查询后补充。
type ArticleAggregate struct {
	// Article 的基础信息
	Article *entity.Article

	// 作者的基础信息
	Author *entity.User

	// 分类
	Cate *entity.Category
}

func NewArticleAggregate(article *entity.Article, author *entity.User, Cate *entity.Category) *ArticleAggregate {
	return &ArticleAggregate{
		Article: article,
		Author:  author,
		Cate:    Cate,
	}
}

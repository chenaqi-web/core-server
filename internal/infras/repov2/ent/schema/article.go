package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Article holds the schema definition for the Article entity.
type Article struct {
	ent.Schema
}

// Annotations keeps Ent mapped to the existing table used by the SQLX repository.
func (Article) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Table: "blog_article"},
	}
}

// Fields of the Article.
func (Article) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable().Comment("软删除时间 为空表示未删除"),
		field.Time("published_at").Optional().Nillable().Comment("发布时间 为空表示未发布 还是草稿"),
		field.String("title").Comment("文章标题"),
		field.String("summary").Default("").Comment("文章概述"),
		field.String("content").Comment("文章正文"),
		field.String("cover_image").Default("").Comment("封面图片"),
		field.Uint64("author_id").Comment("作者ID"),
		field.Uint64("category_id").Comment("分类ID"),
		field.Bool("is_top").Default(false).Comment("是否置顶 默认为不置顶"),
		field.Bool("is_published").Default(true).Comment("是否发布 默认发布"),
		field.Int("visibility").Default(1).Comment("可见性 0为私密 1为公开 默认为公开 后续可加入仅关注可看与付费等等"),
		field.Int("view_count").Default(0).Comment("浏览量"),
		field.Int("like_count").Default(0).Comment("点赞量"),
		field.Int("comment_count").Default(0).Comment("评论量"),
	}
}

// Edges of the Article.
// article.go
func (Article) Edges() []ent.Edge {
	return []ent.Edge{
		// 文章属于一个作者
		edge.From("user", User.Type).
			Ref("articles").    // User 里的边叫 "articles"，这里表示的是反向边
			Field("author_id"). // 外键字段
			Unique().           // 唯一，一篇文章只有一个作者
			Required(),         // 必须设置
		// 文章属于一个分类
		edge.From("category", Category.Type).
			Ref("articles").      // Category 里的边叫 "articles",
			Field("category_id"). // 外键字段
			Unique().             // 唯一
			Required(),           // 必须设置

	}
}

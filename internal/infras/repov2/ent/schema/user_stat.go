package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// UserStat holds the denormalized counters for a user.
type UserStat struct {
	ent.Schema
}

func (UserStat) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "user_stat"}}
}

func (UserStat) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Immutable().Default(time.Now).Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
		field.Time("deleted_at").Optional().Nillable().Comment("删除时间"),

		field.Uint64("user_id").Unique(),
		field.Uint64("article_count").Optional().Comment("发帖/文章数量"),
		field.Uint64("followers_count").Default(0).Comment("粉丝数"),
		field.Uint64("following_count").Default(0).Comment("关注数"),
		field.Uint64("like_count").Default(0).Comment("点赞总数"),
		field.Uint64("receive_like_count").Default(0).Comment("收到的点赞总数"),
		field.Uint64("favor_count").Default(0).Comment("收藏的数量"),
		field.Uint64("receive_favor_count").Default(0).Comment("被收藏的总数"),
	}
}
func (UserStat) Edges() []ent.Edge {
	return []ent.Edge{edge.From("user", User.Type).Ref("stat").Field("user_id").Unique().Required()}
}

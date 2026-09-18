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
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable().Comment("软删除时间，为空表示未删除"),

		field.Uint64("user_id").Unique(),
		field.Uint64("followers_count").Default(0),
		field.Uint64("following_count").Default(0),
		field.Uint64("like_count").Default(0),
		field.Uint64("receive_like_count").Default(0),
		field.Uint64("view_count").Default(0),
		field.Uint64("receive_view_count").Default(0),
		field.Uint64("favor_count").Default(0),
		field.Uint64("receive_favor_count").Default(0),
	}
}

func (UserStat) Edges() []ent.Edge {
	return []ent.Edge{edge.From("user", User.Type).Ref("stat").Field("user_id").Unique().Required()}
}

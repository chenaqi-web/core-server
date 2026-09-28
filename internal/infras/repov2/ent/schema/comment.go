package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Comment stores both top-level comments and replies.
type Comment struct{ ent.Schema }

func (Comment) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "comment"}}
}

func (Comment) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id"),
		field.Time("created_at").Default(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
		field.Uint64("article_id"),
		field.Uint64("user_id"),
		field.Uint64("parent_id").Default(0),
		field.Uint64("root_id").Default(0),
		field.Uint64("reply_to_id").Default(0),
		field.String("content"),
		field.Uint32("like_count").Default(0),
		field.Uint32("child_count").Default(0),
	}
}

func (Comment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("article", Article.Type).Ref("comments").Field("article_id").Unique().Required(),
		edge.From("user", User.Type).Ref("comments").Field("user_id").Unique().Required(),
	}
}

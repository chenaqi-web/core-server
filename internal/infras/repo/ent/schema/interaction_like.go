package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InteractionLike records a user's current like state for an object.
type InteractionLike struct{ ent.Schema }

func (InteractionLike) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "interaction_like"}}
}

func (InteractionLike) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id"),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Uint64("user_id"),
		field.Enum("object_type").Values("article", "comment", "life"),
		field.Uint64("object_id"),
		field.Enum("status").Values("unknown", "thumb_up", "nothing").Default("nothing"),
		field.Int64("version").Default(0),
	}
}

func (InteractionLike) Indexes() []ent.Index {
	return []ent.Index{index.Fields("user_id", "object_type", "object_id").Unique().StorageKey("uk_like_user_object")}
}

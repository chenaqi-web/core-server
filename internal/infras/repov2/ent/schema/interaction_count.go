package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InteractionCount holds the schema definition for the InteractionCount entity.
type InteractionCount struct {
	ent.Schema
}

// Annotations keeps Ent mapped to the existing table used by the SQLX repository.
func (InteractionCount) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Table: "interaction_count"},
	}
}

// Fields of the InteractionCount.
func (InteractionCount) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id"),

		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),

		field.Uint64("object_id").Comment("对象ID 例如文章ID"),
		field.Enum("object_type").Values("article").Immutable().Comment("对象类型 例如文章"),
		field.Enum("interaction_type").Values("like", "view", "favor").Immutable().Comment("交互类型"),
		field.Int("count").Default(0).Comment("计数值"),
	}
}

// Edges of the InteractionCount.
func (InteractionCount) Edges() []ent.Edge {
	return nil
}

func (InteractionCount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("object_id", "object_type", "interaction_type").
			Unique().                // 创建唯一约束 不可重复
			StorageKey("uk_object"), // 指定索引名
	}
}

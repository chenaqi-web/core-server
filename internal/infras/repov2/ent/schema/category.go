package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Category holds the two-level category tree.
type Category struct {
	ent.Schema
}

func (Category) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "category"}}
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
		field.Uint64("parent_id").Default(0),
		field.String("name").MaxLen(64),
	}
}

func (Category) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_id", "name").Unique(),
	}
}

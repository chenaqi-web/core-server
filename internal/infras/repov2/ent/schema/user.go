package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// User holds the schema definition for the User entity.
type User struct{ ent.Schema }

func (User) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "user"}}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id"),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable().Comment("soft delete time"),
		field.String("name").Comment("username"),
		field.String("password").MaxLen(20).Comment("password"),
		field.String("phone").MaxLen(20).Default(""),
		field.String("avatar").MaxLen(255).Default(""),
		field.String("email").MaxLen(50).Default(""),
		field.Enum("role").Values("admin", "user").Default("user"),
		field.Enum("sex").Values("male", "female", "secret").Default("secret"),
		field.Time("birthday").Optional().Nillable(),
		field.String("signature").MaxLen(255).Default(""),
		field.Enum("status").Values("approved", "blocked").Default("approved"),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{edge.To("stat", UserStat.Type).Unique()}
}

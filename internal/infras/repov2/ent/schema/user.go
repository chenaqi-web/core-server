package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Annotations keeps Ent mapped to the existing table used by the SQLX repository.
func (User) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Table: "user"},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id"),

		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable().Comment("软删除时间，为空表示未删除"),

		// --- 基础信息 ---
		field.String("name").Comment("用户名"),
		field.String("password").MaxLen(255).Comment("密码"),
		field.String("phone").MaxLen(20).Default("").Comment("手机号"),
		field.String("avatar").MaxLen(500).Default("").Comment("头像"),
		field.String("email").MaxLen(100).Default("").Comment("邮箱"),
		field.String("role").MaxLen(20).Default("user").Comment("角色"),
		field.String("sex").MaxLen(6).Default("").Comment("性别"),
		field.Time("birthday").Optional().Nillable().Comment("生日"),
		field.String("status").MaxLen(20).Default("active").Comment("用户状态"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("stat", UserStat.Type).Unique(),

		edge.To("articles", Article.Type).Unique(),
	}
}

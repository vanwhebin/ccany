package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable(),
		field.String("username").
			Unique().
			Comment("Username"),
		field.String("email").
			Optional().
			Comment("Email address"),
		field.String("password_hash").
			Comment("Password hash"),
		field.String("salt").
			Comment("Password salt"),
		field.String("role").
			Default("user").
			Comment("User role: admin, user"),
		field.Bool("is_active").
			Default(true).
			Comment("Whether activated"),
		field.Time("last_login").
			Optional().
			Comment("Last login time"),
		field.String("session_token").
			Optional().
			Comment("Session token"),
		field.Time("session_expires").
			Optional().
			Comment("Session expiration time"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username").Unique(),
		index.Fields("email").Unique(),
		index.Fields("session_token"),
		index.Fields("created_at"),
	}
}

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AppConfig holds the schema definition for the AppConfig entity.
type AppConfig struct {
	ent.Schema
}

// Fields of the AppConfig.
func (AppConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable(),
		field.String("key").
			Unique().
			Comment("Configuration key name"),
		field.Text("value").
			Comment("Configuration value"),
		field.String("category").
			Default("general").
			Comment("Configuration category: api, model, server, performance"),
		field.String("type").
			Default("string").
			Comment("Value type: string, int, bool, json"),
		field.String("description").
			Optional().
			Comment("Configuration description"),
		field.Bool("is_encrypted").
			Default(false).
			Comment("Whether to encrypt storage"),
		field.Bool("is_required").
			Default(false).
			Comment("Whether required"),
		field.String("default_value").
			Optional().
			Comment("Default value"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the AppConfig.
func (AppConfig) Edges() []ent.Edge {
	return nil
}

// Indexes of the AppConfig.
func (AppConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("key").Unique(),
		index.Fields("category"),
		index.Fields("is_required"),
	}
}

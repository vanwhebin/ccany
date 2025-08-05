package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Channel holds the schema definition for the Channel entity.
type Channel struct {
	ent.Schema
}

// Fields of the Channel.
func (Channel) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable(),
		field.String("name").
			Comment("Channel display name"),
		field.String("provider").
			Comment("API provider: openai, anthropic, gemini"),
		field.String("base_url").
			Comment("API base URL"),
		field.String("api_key").
			Sensitive().
			Comment("API key for the provider"),
		field.String("custom_key").
			Unique().
			Comment("Custom key for users to access this channel"),
		field.Int("timeout").
			Default(30).
			Comment("Request timeout in seconds"),
		field.Int("max_retries").
			Default(3).
			Comment("Maximum retry attempts"),
		field.Bool("enabled").
			Default(true).
			Comment("Whether the channel is enabled"),
		field.Int("weight").
			Default(1).
			Comment("Load balancing weight"),
		field.Int("priority").
			Default(1).
			Comment("Channel priority (higher is better)"),
		field.JSON("models_mapping", map[string]string{}).
			Optional().
			Comment("Model name mappings"),
		field.JSON("capabilities", map[string]interface{}{}).
			Optional().
			Comment("Detected capabilities"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("last_used_at").
			Optional().
			Nillable().
			Comment("Last time this channel was used"),
		field.Int64("request_count").
			Default(0).
			Comment("Total number of requests"),
		field.Int64("error_count").
			Default(0).
			Comment("Total number of errors"),
		field.Float("success_rate").
			Default(1.0).
			Comment("Success rate (0.0 to 1.0)"),
		field.Int64("total_tokens").
			Default(0).
			Comment("Total tokens processed"),
		field.Float("avg_response_time").
			Default(0.0).
			Comment("Average response time in seconds"),
	}
}

// Edges of the Channel.
func (Channel) Edges() []ent.Edge {
	return nil
}

// Indexes of the Channel.
func (Channel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("custom_key").Unique(),
		index.Fields("provider"),
		index.Fields("enabled"),
		index.Fields("priority"),
		index.Fields("weight"),
		index.Fields("success_rate"),
		index.Fields("last_used_at"),
	}
}

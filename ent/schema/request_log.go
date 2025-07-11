package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RequestLog holds the schema definition for the RequestLog entity.
type RequestLog struct {
	ent.Schema
}

// Fields of the RequestLog.
func (RequestLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable(),
		field.String("claude_model").
			Comment("Original Claude model requested"),
		field.String("openai_model").
			Comment("Mapped OpenAI model used"),
		field.Text("request_body").
			Comment("Original Claude request body"),
		field.Text("response_body").
			Optional().
			Comment("Response body"),
		field.Int("status_code").
			Default(0).
			Comment("HTTP status code"),
		field.Bool("is_streaming").
			Default(false).
			Comment("Whether request was streaming"),
		field.Int("input_tokens").
			Default(0).
			Comment("Number of input tokens"),
		field.Int("output_tokens").
			Default(0).
			Comment("Number of output tokens"),
		field.Float("duration_ms").
			Default(0).
			Comment("Request duration in milliseconds"),
		field.String("error_message").
			Optional().
			Comment("Error message if request failed"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the RequestLog.
func (RequestLog) Edges() []ent.Edge {
	return nil
}

// Indexes of the RequestLog.
func (RequestLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
		index.Fields("claude_model"),
		index.Fields("status_code"),
	}
}

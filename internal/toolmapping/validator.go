package toolmapping

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// SchemaValidator validates tool inputs against their schemas
type SchemaValidator struct {
	definitions map[string]map[string]interface{}
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{
		definitions: make(map[string]map[string]interface{}),
	}
}

// ValidateToolInput validates input against a tool's schema
func (v *SchemaValidator) ValidateToolInput(toolName string, input interface{}) error {
	// Get tool definition
	def, exists := GetToolDefinitionByName(toolName)
	if !exists {
		// Try getting from all definitions
		allDefs := GetAllToolDefinitions()
		if toolDef, ok := allDefs[toolName]; ok {
			def = &toolDef
		} else {
			return fmt.Errorf("unknown tool: %s", toolName)
		}
	}

	// Validate against schema
	return v.validateAgainstSchema(def.InputSchema, input)
}

// ValidateVersionedToolCall validates input for versioned tools
func (v *SchemaValidator) ValidateVersionedToolCall(toolType string, input interface{}) error {
	// Get all definitions and find by type
	allDefs := GetAllToolDefinitions()

	var foundDef *map[string]interface{}
	for _, def := range allDefs {
		if def.Type == toolType || def.Name == toolType {
			foundDef = &def.InputSchema
			break
		}
	}

	if foundDef == nil {
		return fmt.Errorf("unknown tool type: %s", toolType)
	}

	return v.validateAgainstSchema(*foundDef, input)
}

// validateAgainstSchema validates input against a JSON schema
func (v *SchemaValidator) validateAgainstSchema(schema map[string]interface{}, input interface{}) error {
	// Convert input to map if needed
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		// Try JSON marshaling/unmarshaling to convert
		jsonBytes, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("invalid input format: %v", err)
		}
		if err := json.Unmarshal(jsonBytes, &inputMap); err != nil {
			return fmt.Errorf("invalid input format: %v", err)
		}
	}

	// Get schema properties
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid schema: missing properties")
	}

	// Get required fields
	requiredFields := []string{}
	if required, ok := schema["required"].([]string); ok {
		requiredFields = required
	} else if required, ok := schema["required"].([]interface{}); ok {
		for _, r := range required {
			if str, ok := r.(string); ok {
				requiredFields = append(requiredFields, str)
			}
		}
	}

	// Check required fields
	for _, field := range requiredFields {
		if _, exists := inputMap[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Validate each field
	for fieldName, value := range inputMap {
		fieldSchema, exists := properties[fieldName]
		if !exists {
			// Field not in schema - this might be okay for additional properties
			continue
		}

		if err := v.validateField(fieldName, fieldSchema, value); err != nil {
			return err
		}
	}

	return nil
}

// validateField validates a single field against its schema
func (v *SchemaValidator) validateField(fieldName string, fieldSchema interface{}, value interface{}) error {
	schemaMap, ok := fieldSchema.(map[string]interface{})
	if !ok {
		return nil // Skip validation if schema is not a map
	}

	// Get expected type
	expectedType, ok := schemaMap["type"].(string)
	if !ok {
		return nil // Skip validation if type is not specified
	}

	// Check type
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field %s: expected string, got %T", fieldName, value)
		}

		// Check enum values if specified
		if enum, ok := schemaMap["enum"].([]interface{}); ok {
			found := false
			valueStr := value.(string)
			for _, e := range enum {
				if enumStr, ok := e.(string); ok && enumStr == valueStr {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("field %s: value '%s' not in enum", fieldName, valueStr)
			}
		}

	case "integer", "number":
		// Accept both int and float64 for numbers
		switch value.(type) {
		case int, int32, int64, float32, float64:
			// Valid number type
		default:
			return fmt.Errorf("field %s: expected number, got %T", fieldName, value)
		}

	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field %s: expected boolean, got %T", fieldName, value)
		}

	case "array":
		if !isArray(value) {
			return fmt.Errorf("field %s: expected array, got %T", fieldName, value)
		}

	case "object":
		if !isObject(value) {
			return fmt.Errorf("field %s: expected object, got %T", fieldName, value)
		}
	}

	return nil
}

// isArray checks if a value is an array/slice
func isArray(value interface{}) bool {
	if value == nil {
		return false
	}

	valueType := reflect.TypeOf(value)
	return valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Array
}

// isObject checks if a value is an object/map
func isObject(value interface{}) bool {
	if value == nil {
		return false
	}

	valueType := reflect.TypeOf(value)
	return valueType.Kind() == reflect.Map || valueType.Kind() == reflect.Struct
}

// GetSchemaType gets the type from a schema
func GetSchemaType(schema map[string]interface{}) string {
	if typeStr, ok := schema["type"].(string); ok {
		return typeStr
	}
	return ""
}

// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mcpkit

// PropOption configures constraints or metadata on a JSON Schema property
// added via [ToolBuilder] methods.
type PropOption func(b *ToolBuilder, name string, prop map[string]any)

// Required marks the property as required in the parent object's JSON Schema.
func Required() PropOption {
	return func(b *ToolBuilder, name string, _ map[string]any) {
		b.required = append(b.required, name)
	}
}

// Min sets the numeric `minimum` constraint on a number/integer property.
func Min(val float64) PropOption {
	return func(_ *ToolBuilder, _ string, prop map[string]any) {
		prop["minimum"] = val
	}
}

// Max sets the numeric `maximum` constraint on a number/integer property.
func Max(val float64) PropOption {
	return func(_ *ToolBuilder, _ string, prop map[string]any) {
		prop["maximum"] = val
	}
}

// MinItems sets the `minItems` constraint on an array property.
func MinItems(n int) PropOption {
	return func(_ *ToolBuilder, _ string, prop map[string]any) {
		if n > 0 {
			prop["minItems"] = n
		}
	}
}

// MaxItems sets the `maxItems` constraint on an array property.
func MaxItems(n int) PropOption {
	return func(_ *ToolBuilder, _ string, prop map[string]any) {
		if n > 0 {
			prop["maxItems"] = n
		}
	}
}

// Default sets the `default` value metadata on a property.
func Default(val any) PropOption {
	return func(_ *ToolBuilder, _ string, prop map[string]any) {
		prop["default"] = val
	}
}

// ToolBuilder provides a fluent, zero-nesting builder for constructing MCP v1.0
// [ToolDefinition] schemas with functional [PropOption] modifiers.
type ToolBuilder struct {
	name        string
	description string
	properties  map[string]any
	required    []string
}

// NewTool creates a fluent [ToolBuilder] for an MCP tool with the given name and description.
func NewTool(name, description string) *ToolBuilder {
	return &ToolBuilder{
		name:        name,
		description: description,
		properties:  make(map[string]any),
	}
}

func (b *ToolBuilder) addProp(name, typ, description string, opts ...PropOption) map[string]any {
	prop := map[string]any{"type": typ}
	if description != "" {
		prop["description"] = description
	}
	for _, opt := range opts {
		if opt != nil {
			opt(b, name, prop)
		}
	}
	b.properties[name] = prop
	return prop
}

// Prop adds a custom JSON Schema property definition to the tool input schema.
func (b *ToolBuilder) Prop(name string, schema map[string]any, opts ...PropOption) *ToolBuilder {
	for _, opt := range opts {
		if opt != nil {
			opt(b, name, schema)
		}
	}
	b.properties[name] = schema
	return b
}

// String adds a string property to the tool input schema.
func (b *ToolBuilder) String(name, description string, opts ...PropOption) *ToolBuilder {
	b.addProp(name, "string", description, opts...)
	return b
}

// Enum adds a string enum property constrained to the specified values.
func (b *ToolBuilder) Enum(name, description string, values []string, opts ...PropOption) *ToolBuilder {
	prop := b.addProp(name, "string", description, opts...)
	prop["enum"] = values
	return b
}

// Int adds an integer property to the tool input schema.
func (b *ToolBuilder) Int(name, description string, opts ...PropOption) *ToolBuilder {
	b.addProp(name, "integer", description, opts...)
	return b
}

// Number adds a floating-point number property to the tool input schema.
func (b *ToolBuilder) Number(name, description string, opts ...PropOption) *ToolBuilder {
	b.addProp(name, "number", description, opts...)
	return b
}

// Bool adds a boolean property to the tool input schema.
func (b *ToolBuilder) Bool(name, description string, opts ...PropOption) *ToolBuilder {
	b.addProp(name, "boolean", description, opts...)
	return b
}

// StringArray adds an array-of-strings property to the tool input schema.
func (b *ToolBuilder) StringArray(name, description string, opts ...PropOption) *ToolBuilder {
	prop := b.addProp(name, "array", description, opts...)
	prop["items"] = map[string]any{"type": "string"}
	return b
}

// IntArray adds an array-of-integers property to the tool input schema.
func (b *ToolBuilder) IntArray(name, description string, opts ...PropOption) *ToolBuilder {
	prop := b.addProp(name, "array", description, opts...)
	prop["items"] = map[string]any{"type": "integer"}
	return b
}

// ObjectArray adds an array-of-objects property whose item properties are built via an inline [ToolBuilder] callback.
// If buildItem is nil, items defaults to a generic `{"type": "object"}` schema.
func (b *ToolBuilder) ObjectArray(
	name, description string,
	buildItem func(item *ToolBuilder),
	opts ...PropOption,
) *ToolBuilder {
	itemSchema := map[string]any{"type": "object"}
	if buildItem != nil {
		itemBuilder := NewTool("", "")
		buildItem(itemBuilder)
		itemSchema["properties"] = itemBuilder.properties
		if len(itemBuilder.required) > 0 {
			itemSchema["required"] = itemBuilder.required
		}
	}
	prop := b.addProp(name, "array", description, opts...)
	prop["items"] = itemSchema
	return b
}

// Object adds a nested object property whose fields are configured via an inline [ToolBuilder] callback.
func (b *ToolBuilder) Object(
	name, description string,
	buildObj func(obj *ToolBuilder),
	opts ...PropOption,
) *ToolBuilder {
	objBuilder := NewTool("", "")
	if buildObj != nil {
		buildObj(objBuilder)
	}
	prop := b.addProp(name, "object", description, opts...)
	prop["properties"] = objBuilder.properties
	if len(objBuilder.required) > 0 {
		prop["required"] = objBuilder.required
	}
	return b
}

// Build finalizes and returns the [ToolDefinition].
func (b *ToolBuilder) Build() ToolDefinition {
	schema := map[string]any{
		"type":       "object",
		"properties": b.properties,
	}
	if len(b.required) > 0 {
		schema["required"] = b.required
	}
	return ToolDefinition{
		Name:        b.name,
		Description: b.description,
		InputSchema: schema,
	}
}

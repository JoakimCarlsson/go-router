package openapi

import (
	"github.com/joakimcarlsson/go-router/metadata"
)

// SchemaFromMetadataSchema converts a metadata Schema to an OpenAPI Schema
func SchemaFromMetadataSchema(s metadata.Schema) Schema {
	return Schema{
		Type:                 s.Type,
		Ref:                  s.Ref,
		Format:               s.Format,
		Description:          s.Description,
		Example:              s.Example,
		Required:             s.Required,
		MinLength:            s.MinLength,
		MaxLength:            s.MaxLength,
		Minimum:              s.Minimum,
		Maximum:              s.Maximum,
		Enum:                 s.Enum,
		Nullable:             s.Nullable,
		TypeName:             s.TypeName,
		Properties:           convertProperties(s.Properties),
		Items:                convertItems(s.Items),
		AllOf:                convertSchemaSlice(s.AllOf),
		OneOf:                convertSchemaSlice(s.OneOf),
		AnyOf:                convertSchemaSlice(s.AnyOf),
		AdditionalProperties: convertAdditionalProperties(s.AdditionalProperties),
	}
}

func convertProperties(props map[string]metadata.Schema) map[string]Schema {
	if props == nil {
		return nil
	}
	result := make(map[string]Schema, len(props))
	for k, v := range props {
		result[k] = SchemaFromMetadataSchema(v)
	}
	return result
}

func convertItems(items *metadata.Schema) *Schema {
	if items == nil {
		return nil
	}
	schema := SchemaFromMetadataSchema(*items)
	return &schema
}

func convertSchemaSlice(schemas []metadata.Schema) []Schema {
	if schemas == nil {
		return nil
	}
	result := make([]Schema, len(schemas))
	for i, s := range schemas {
		result[i] = SchemaFromMetadataSchema(s)
	}
	return result
}

func convertAdditionalProperties(props *metadata.Schema) *Schema {
	if props == nil {
		return nil
	}
	schema := SchemaFromMetadataSchema(*props)
	return &schema
}

// ParameterFromMetadataParameter converts a metadata Parameter to an OpenAPI Parameter
func ParameterFromMetadataParameter(p metadata.Parameter) Parameter {
	return Parameter{
		Name:        p.Name,
		In:          p.In,
		Required:    p.Required,
		Description: p.Description,
		Schema:      SchemaFromMetadataSchema(p.Schema),
		Example:     p.Example,
	}
}

// ResponseFromMetadataResponse converts a metadata Response to an OpenAPI Response
func ResponseFromMetadataResponse(r metadata.Response) Response {
	content := make(map[string]MediaType)
	for k, v := range r.Content {
		schema := SchemaFromMetadataSchema(v.Schema)
		if schema.Type == "array" && schema.Items != nil {
			if schema.Items.Ref != "" {
				// For arrays with component references
				content[k] = MediaType{
					Schema: Schema{
						Type:  "array",
						Items: schema.Items,
					},
				}
			} else if schema.Items.Type == "object" && schema.Items.TypeName != "" {
				// For arrays of objects that should be referenced
				sanitizedName := metadata.SanitizeSchemaName(schema.Items.TypeName)
				content[k] = MediaType{
					Schema: Schema{
						Type: "array",
						Items: &Schema{
							Ref: "#/components/schemas/" + sanitizedName,
						},
					},
				}
			} else {
				// For arrays of primitive types
				content[k] = MediaType{
					Schema: schema,
				}
			}
		} else if schema.Ref != "" {
			// For direct references
			content[k] = MediaType{
				SchemaRef: &Reference{
					Ref: schema.Ref,
				},
			}
		} else {
			// For other cases
			content[k] = MediaType{
				Schema:  schema,
				Example: v.Example,
			}
		}
	}

	headers := make(map[string]Header)
	for k, v := range r.Headers {
		headers[k] = Header{
			Description: v.Description,
			Schema:      SchemaFromMetadataSchema(v.Schema),
		}
	}

	return Response{
		Description: r.Description,
		Content:     content,
		Headers:     headers,
	}
}

// RequestBodyFromMetadataRequestBody converts a metadata RequestBody to an OpenAPI RequestBody
func RequestBodyFromMetadataRequestBody(r *metadata.RequestBody) *RequestBody {
	if r == nil {
		return nil
	}

	content := make(map[string]MediaType)
	for k, v := range r.Content {
		schema := SchemaFromMetadataSchema(v.Schema)
		if schema.Type == "array" && schema.Items != nil {
			if schema.Items.Ref != "" {
				// For arrays with component references
				content[k] = MediaType{
					Schema: Schema{
						Type: "array",
						Items: &Schema{
							Ref: schema.Items.Ref,
						},
					},
				}
			} else if schema.Items.Type == "object" && schema.Items.TypeName != "" {
				// For arrays of objects that should be referenced
				sanitizedName := metadata.SanitizeSchemaName(schema.Items.TypeName)
				content[k] = MediaType{
					Schema: Schema{
						Type: "array",
						Items: &Schema{
							Ref: "#/components/schemas/" + sanitizedName,
						},
					},
				}
			} else {
				// For arrays of primitive types
				content[k] = MediaType{
					Schema: schema,
				}
			}
		} else if schema.Ref != "" {
			// For direct references
			content[k] = MediaType{
				SchemaRef: &Reference{
					Ref: schema.Ref,
				},
			}
		} else {
			content[k] = MediaType{
				Schema:  schema,
				Example: v.Example,
			}
		}
	}

	return &RequestBody{
		Description: r.Description,
		Required:    r.Required,
		Content:     content,
	}
}

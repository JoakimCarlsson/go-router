package openapi

import (
	"reflect"

	"github.com/joakimcarlsson/go-router/router/v2"
)

// MetadataFromOptions converts a slice of router.RouteOption to RouteMetadata.
func MetadataFromOptions(method, path string, opts []router.RouteOption) *RouteMetadata {
	m := &RouteMetadata{
		Method:     method,
		Path:       path,
		Parameters: make([]Parameter, 0),
		Tags:       make([]string, 0),
		Responses:  make(map[string]Response),
		Security:   make([]SecurityRequirement, 0),
		SSEEvents:  make([]SSEEventSchema, 0),
	}

	for _, opt := range opts {
		applyOption(m, opt)
	}

	return m
}

func applyOption(m *RouteMetadata, opt router.RouteOption) {
	switch o := opt.(type) {
	case OperationIDOption:
		m.OperationID = o.OperationID

	case SummaryOption:
		m.Summary = o.Summary

	case DescriptionOption:
		m.Description = o.Description

	case TagsOption:
		m.Tags = append(m.Tags, o.Tags...)

	case DeprecatedOption:
		m.Deprecated = true
		if o.Message != "" {
			if m.Description != "" {
				m.Description += "\n\n"
			}
			m.Description += "DEPRECATED: " + o.Message
		}

	case ExcludeFromDocsOption:
		m.ExcludeFromDocs = true

	case ParameterOption:
		m.Parameters = append(m.Parameters, Parameter{
			Name:        o.Name,
			In:          o.In,
			Required:    o.Required,
			Description: o.Description,
			Schema: Schema{
				Type:    o.Type,
				Example: o.Example,
			},
		})

	case ParameterWithSchemaOption:
		m.Parameters = append(m.Parameters, Parameter{
			Name:        o.Name,
			In:          o.In,
			Required:    o.Required,
			Description: o.Description,
			Schema:      o.Schema,
			Style:       o.Style,
			Explode:     o.Explode,
		})

	case RequestBodyOption:
		m.RequestBody = &RequestBody{
			Description: o.Description,
			Required:    o.Required,
			Content: map[string]MediaType{
				o.ContentType: {Schema: o.Schema},
			},
		}

	case JSONRequestBodyOption:
		schema := SchemaFromType(o.Type)
		m.RequestBody = &RequestBody{
			Description: o.Description,
			Required:    o.Required,
			Content: map[string]MediaType{
				ContentTypeJSON: {Schema: schema},
			},
		}

	case MultipartFormDataOption:
		properties := make(map[string]Schema)
		requiredFields := make([]string, 0)

		for fieldName, spec := range o.FormFields {
			switch spec.Type {
			case "file[]":
				properties[fieldName] = Schema{
					Type: "array",
					Items: &Schema{
						Type:        "string",
						Format:      "binary",
						Description: spec.Description,
					},
				}
			case "file":
				properties[fieldName] = Schema{
					Type:        "string",
					Format:      "binary",
					Description: spec.Description,
				}
			default:
				properties[fieldName] = Schema{
					Type:        "string",
					Description: spec.Description,
				}
			}

			if spec.Required {
				requiredFields = append(requiredFields, fieldName)
			}
		}

		schema := Schema{
			Type:       "object",
			Properties: properties,
		}
		if len(requiredFields) > 0 {
			schema.Required = requiredFields
		}

		m.RequestBody = &RequestBody{
			Description: o.Description,
			Required:    len(requiredFields) > 0,
			Content: map[string]MediaType{
				ContentTypeFormData: {Schema: schema},
			},
		}

	case MultipartFormStructOption:
		properties := make(map[string]Schema)
		requiredFields := make([]string, 0)
		t := o.Type

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			formTag := field.Tag.Get("form")
			if formTag == "" {
				continue
			}

			isFile := field.Tag.Get("file") == "true"
			isRequired := field.Tag.Get("required") == "true"
			fieldDesc := field.Tag.Get("description")
			if fieldDesc == "" {
				fieldDesc = formTag
			}

			isFileArray := isFile && field.Type.Kind() == reflect.Slice

			var schema Schema
			if isFile {
				if isFileArray {
					schema = Schema{
						Type: "array",
						Items: &Schema{
							Type:        "string",
							Format:      "binary",
							Description: fieldDesc,
						},
					}
				} else {
					schema = Schema{
						Type:        "string",
						Format:      "binary",
						Description: fieldDesc,
					}
				}
			} else {
				schema = Schema{
					Type:        "string",
					Description: fieldDesc,
				}
			}

			properties[formTag] = schema
			if isRequired {
				requiredFields = append(requiredFields, formTag)
			}
		}

		schema := Schema{
			Type:       "object",
			Properties: properties,
		}
		if len(requiredFields) > 0 {
			schema.Required = requiredFields
		}

		m.RequestBody = &RequestBody{
			Description: o.Description,
			Required:    len(requiredFields) > 0,
			Content: map[string]MediaType{
				ContentTypeFormData: {Schema: schema},
			},
		}

	case ResponseOption:
		EnsureResponsesMap(m)
		m.Responses[StatusCodeToString(o.StatusCode)] = Response{
			Description: o.Description,
		}

	case JSONResponseOption:
		schema := SchemaFromType(o.Type)
		EnsureResponsesMap(m)
		m.Responses[StatusCodeToString(o.StatusCode)] = Response{
			Description: o.Description,
			Content: map[string]MediaType{
				ContentTypeJSON: {Schema: schema},
			},
		}

	case SecurityOption:
		for _, req := range o.Requirements {
			secReq := make(SecurityRequirement)
			for k, v := range req {
				secReq[k] = v
			}
			m.Security = append(m.Security, secReq)
		}

	case SSEResponseOption:
		EnsureResponsesMap(m)
		m.Responses["200"] = Response{
			Description: o.Description,
			Content: map[string]MediaType{
				ContentTypeEventStream: {
					Schema: Schema{
						Type:        "string",
						Description: "Server-Sent Events stream",
					},
				},
			},
		}
		m.IsSSE = true

	case SSEEventOption:
		schema := SchemaFromType(o.Type)
		m.SSEEvents = append(m.SSEEvents, SSEEventSchema{
			EventName:   o.EventName,
			Description: o.Description,
			Schema:      schema,
		})

	case SSEEventsOption:
		EnsureResponsesMap(m)
		m.Responses["200"] = Response{
			Description: o.Description,
			Content: map[string]MediaType{
				ContentTypeEventStream: {
					Schema: Schema{
						Type:        "string",
						Description: "Server-Sent Events stream",
					},
				},
			},
		}
		m.IsSSE = true

		for _, event := range o.Events {
			m.SSEEvents = append(m.SSEEvents, SSEEventSchema{
				EventName:   event.Name,
				Description: event.Description,
				Schema:      event.Schema,
			})
		}
	}
}

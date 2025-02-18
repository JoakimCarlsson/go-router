package openapi

import (
	"reflect"
	"strings"
)

// Spec represents the OpenAPI 3.0.0 specification
type Spec struct {
	OpenAPI    string              `json:"openapi"`
	Info       Info                `json:"info"`
	Servers    []Server            `json:"servers,omitempty"`
	Paths      map[string]PathItem `json:"paths"`
	Components *Components         `json:"components,omitempty"`
}

type Info struct {
	Title          string  `json:"title"`
	Description    string  `json:"description,omitempty"`
	Version        string  `json:"version"`
	TermsOfService string  `json:"termsOfService,omitempty"`
	Contact        Contact `json:"contact,omitempty"`
}

type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
	Patch  *Operation `json:"patch,omitempty"`
}

type Operation struct {
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

type RequestBody struct {
	Description string               `json:"description"`
	Required    bool                 `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

type MediaType struct {
	Schema Schema `json:"schema"`
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"` // query, path, header, cookie
	Required    bool   `json:"required,omitempty"`
	Description string `json:"description,omitempty"`
	Schema      Schema `json:"schema"`
}

type Schema struct {
	Type        string            `json:"type,omitempty"`
	Format      string            `json:"format,omitempty"`
	Description string            `json:"description,omitempty"`
	Items       *Schema           `json:"items,omitempty"`
	Properties  map[string]Schema `json:"properties,omitempty"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

type Components struct {
	Schemas map[string]Schema `json:"schemas,omitempty"`
}

// SchemaFromType generates an OpenAPI schema from a Go type
func SchemaFromType(t reflect.Type) Schema {
	switch t.Kind() {
	case reflect.Ptr:
		return SchemaFromType(t.Elem())
	case reflect.Struct:
		return Schema{
			Type:       "object",
			Properties: getStructProperties(t),
		}
	case reflect.Slice, reflect.Array:
		return Schema{
			Type: "array",
			Items: &Schema{
				Type: getGoTypeSchema(t.Elem()),
			},
		}
	default:
		return Schema{
			Type: getGoTypeSchema(t),
		}
	}
}

func getStructProperties(t reflect.Type) map[string]Schema {
	props := make(map[string]Schema)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON tag name or field name
		name := field.Tag.Get("json")
		if idx := strings.Index(name, ","); idx != -1 {
			name = name[:idx]
		}
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}

		props[name] = SchemaFromType(field.Type)
	}
	return props
}

func getGoTypeSchema(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	default:
		return "object"
	}
}

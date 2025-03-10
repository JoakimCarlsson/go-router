package openapi

import (
	"encoding/json"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/joakimcarlsson/go-router/metadata"
)

// Spec represents the OpenAPI 3.0.0 specification
type Spec struct {
	OpenAPI      string              `json:"openapi"`
	Info         Info                `json:"info"`
	Servers      []Server            `json:"servers,omitempty"`
	Paths        map[string]PathItem `json:"paths"`
	Components   *Components         `json:"components,omitempty"`
	Tags         []Tag               `json:"tags,omitempty"`
	ExternalDocs map[string]string   `json:"externalDocs,omitempty"`
}

// Reference is a JSON reference to another component in the OpenAPI document
type Reference struct {
	Ref string `json:"$ref"`
}

// SchemaOrReference can be either a Schema object or a Reference to a schema
type SchemaOrReference struct {
	Schema    *Schema    `json:"-"`
	Reference *Reference `json:"-"`
}

// MarshalJSON implements the json.Marshaler interface for SchemaOrReference
func (s SchemaOrReference) MarshalJSON() ([]byte, error) {
	if s.Reference != nil {
		return json.Marshal(s.Reference)
	}
	if s.Schema != nil {
		return json.Marshal(s.Schema)
	}
	return json.Marshal(nil)
}

type Info struct {
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	Version        string   `json:"version"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
}

type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

type Server struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

type PathItem struct {
	Summary     string     `json:"summary,omitempty"`
	Description string     `json:"description,omitempty"`
	Get         *Operation `json:"get,omitempty"`
	Post        *Operation `json:"post,omitempty"`
	Put         *Operation `json:"put,omitempty"`
	Delete      *Operation `json:"delete,omitempty"`
	Patch       *Operation `json:"patch,omitempty"`
	Options     *Operation `json:"options,omitempty"`
	Head        *Operation `json:"head,omitempty"`
	Trace       *Operation `json:"trace,omitempty"`
}

type Operation struct {
	OperationID string                `json:"operationId,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []SecurityRequirement `json:"security,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty"`
}

type SecurityRequirement map[string][]string

type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

// MediaType represents a media type object in OpenAPI spec
type MediaType struct {
	Schema    Schema      `json:"schema,omitempty"`
	Example   interface{} `json:"example,omitempty"`
	SchemaRef *Reference  `json:"-"`
}

// MarshalJSON implements custom JSON marshaling for MediaType to handle schema references properly
func (m MediaType) MarshalJSON() ([]byte, error) {
	if m.SchemaRef != nil {
		return json.Marshal(struct {
			Schema  *Reference  `json:"schema"`
			Example interface{} `json:"example,omitempty"`
		}{
			Schema:  m.SchemaRef,
			Example: m.Example,
		})
	}

	// Otherwise marshal as normal
	return json.Marshal(struct {
		Schema  Schema      `json:"schema"`
		Example interface{} `json:"example,omitempty"`
	}{
		Schema:  m.Schema,
		Example: m.Example,
	})
}

type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // query, path, header, cookie
	Required    bool        `json:"required,omitempty"`
	Description string      `json:"description,omitempty"`
	Schema      Schema      `json:"schema"`
	Example     interface{} `json:"example,omitempty"`
}

// Schema represents an OpenAPI schema
type Schema struct {
	Type                 string            `json:"type,omitempty"`
	Ref                  string            `json:"$ref,omitempty"`
	Format               string            `json:"format,omitempty"`
	Description          string            `json:"description,omitempty"`
	Items                *Schema           `json:"items,omitempty"`
	Properties           map[string]Schema `json:"properties,omitempty"`
	Example              interface{}       `json:"example,omitempty"`
	Required             []string          `json:"required,omitempty"`
	MinLength            *int              `json:"minLength,omitempty"`
	MaxLength            *int              `json:"maxLength,omitempty"`
	Minimum              *float64          `json:"minimum,omitempty"`
	Maximum              *float64          `json:"maximum,omitempty"`
	Enum                 []interface{}     `json:"enum,omitempty"`
	AllOf                []Schema          `json:"allOf,omitempty"`
	OneOf                []Schema          `json:"oneOf,omitempty"`
	AnyOf                []Schema          `json:"anyOf,omitempty"`
	Nullable             bool              `json:"nullable,omitempty"`
	AdditionalProperties *Schema           `json:"additionalProperties,omitempty"`
	TypeName             string            `json:"-"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
	Headers     map[string]Header    `json:"headers,omitempty"`
}

type Header struct {
	Description string `json:"description,omitempty"`
	Schema      Schema `json:"schema"`
}

type Components struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type             string      `json:"type"`
	Scheme           string      `json:"scheme,omitempty"`
	Name             string      `json:"name,omitempty"`
	In               string      `json:"in,omitempty"`
	Description      string      `json:"description,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlows is the configuration container for the supported OAuth Flows
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow configuration details for a specific OAuth Flow
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// Tag represents a tag
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// SchemaFromType generates an OpenAPI schema from a Go type
func SchemaFromType(t reflect.Type) Schema {
	// Special handling for time.Time
	if t.String() == "time.Time" {
		return Schema{
			Type:     "string",
			Format:   "date-time",
			Example:  "2025-02-22T08:36:06.224266+01:00",
			TypeName: "time.Time",
		}
	}

	switch t.Kind() {
	case reflect.Ptr:
		return SchemaFromType(t.Elem())
	case reflect.Struct:
		properties, required := getStructProperties(t)

		// Register the type and get a collision-free name
		typeName := metadata.RegisterType(t)

		schema := Schema{
			Type:       "object",
			Properties: properties,
			TypeName:   typeName,
		}
		if len(required) > 0 {
			schema.Required = required
		}
		if example := generateExample(t); example != nil {
			schema.Example = example
		}
		return schema
	case reflect.Slice, reflect.Array:
		itemSchema := SchemaFromType(t.Elem())
		if itemSchema.Type == "object" && itemSchema.TypeName != "" {
			return Schema{
				Type: "array",
				Items: &Schema{
					Ref: "#/components/schemas/" + metadata.SanitizeSchemaName(itemSchema.TypeName),
				},
				TypeName: "[]" + itemSchema.TypeName,
			}
		}
		return Schema{
			Type:     "array",
			Items:    &itemSchema,
			TypeName: "[]" + itemSchema.TypeName,
		}
	default:
		schema := Schema{
			Type:     getGoTypeSchema(t),
			TypeName: t.Name(),
		}
		schema.Example = getExampleValue(t)
		return schema
	}
}

// sanitizeSchemaName converts a fully qualified type name to a valid schema name
// by removing invalid characters and normalizing the format
func sanitizeSchemaName(name string) string {
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

func getValidationRules(field reflect.StructField) (required bool, minLen, maxLen *int, min *float64) {
	tag := field.Tag.Get("validate")
	if tag == "" {
		return
	}

	rules := strings.Split(tag, ",")
	for _, rule := range rules {
		if rule == "required" {
			required = true
			continue
		}

		if strings.HasPrefix(rule, "min=") {
			val, err := strconv.Atoi(strings.TrimPrefix(rule, "min="))
			if err == nil {
				if field.Type.Kind() == reflect.String {
					minLen = &val
				} else {
					floatVal := float64(val)
					min = &floatVal
				}
			}
		}

		if strings.HasPrefix(rule, "max=") {
			val, err := strconv.Atoi(strings.TrimPrefix(rule, "max="))
			if err == nil && field.Type.Kind() == reflect.String {
				maxLen = &val
			}
		}
	}
	return
}

func getStructProperties(t reflect.Type) (map[string]Schema, []string) {
	properties := make(map[string]Schema)
	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

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

		isRequired, minLen, maxLen, min := getValidationRules(field)
		if isRequired {
			required = append(required, name)
		}

		schema := SchemaFromType(field.Type)
		schema.MinLength = minLen
		schema.MaxLength = maxLen
		schema.Minimum = min
		properties[name] = schema
	}

	return properties, required
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

func getExampleValue(t reflect.Type) interface{} {
	switch t.Kind() {
	case reflect.Bool:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return 42
	case reflect.Float32, reflect.Float64:
		return 3.14
	case reflect.String:
		return "example"
	default:
		return nil
	}
}

func generateExample(t reflect.Type) interface{} {
	if t.Kind() != reflect.Struct {
		return nil
	}

	example := make(map[string]interface{})
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

		// Generate example value for the field
		var value interface{}
		switch field.Type.Kind() {
		case reflect.Struct:
			if field.Type.String() == "time.Time" {
				value = "2025-02-22T08:36:06.224266+01:00"
			} else {
				value = generateExample(field.Type)
			}
		case reflect.Slice, reflect.Array:
			if elemExample := generateExample(field.Type.Elem()); elemExample != nil {
				value = []interface{}{elemExample}
			}
		default:
			value = getExampleValue(field.Type)
		}

		if value != nil {
			example[name] = value
		}
	}

	return example
}

// WriteJSON writes a JSON representation of the value to the writer
func WriteJSON(w io.Writer, value interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

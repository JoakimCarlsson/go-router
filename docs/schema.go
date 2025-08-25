package docs

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/joakimcarlsson/go-router/metadata"
)

// SchemaFromType generates a metadata Schema from a Go type
func SchemaFromType(t reflect.Type) metadata.Schema {
	return schemaFromTypeWithCycle(t, make(map[string]bool))
}

// schemaFromTypeWithCycle generates a metadata Schema from a Go type with circular reference detection
func schemaFromTypeWithCycle(t reflect.Type, visiting map[string]bool) metadata.Schema {
	// Check if there's a registered custom handler for this type
	typeName := t.String()
	if handler, exists := metadata.GetTypeHandler(typeName); exists {
		return handler(t)
	}

	if typeName == "time.Time" {
		return metadata.Schema{
			Type:     "string",
			Format:   "date-time",
			Example:  time.Now().Format(time.RFC3339),
			TypeName: "time.Time",
		}
	}

	if typeName == "uuid.UUID" {
		return metadata.Schema{
			Type:     "string",
			Format:   "uuid",
			Example:  "123e4567-e89b-12d3-a456-426614174000",
			TypeName: "UUID",
		}
	}

	switch t.Kind() {
	case reflect.Ptr:
		schema := schemaFromTypeWithCycle(t.Elem(), visiting)
		schema.Nullable = true
		return schema
	case reflect.Struct:
		typeID := t.String()
		if visiting[typeID] {
			registeredName := metadata.RegisterType(t)
			return metadata.Schema{
				Ref:      "#/components/schemas/" + registeredName,
				TypeName: registeredName,
			}
		}

		visiting[typeID] = true
		defer func() {
			delete(visiting, typeID)
		}()

		properties, required := getStructPropertiesWithCycle(t, visiting)

		// Register the type and get a collision-free name
		typeName := metadata.RegisterType(t)

		schema := metadata.Schema{
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
		elemType := t.Elem()
		itemSchema := schemaFromTypeWithCycle(elemType, visiting)

		if elemType.Kind() == reflect.Struct && elemType.Name() != "" {
			metadata.RegisterType(elemType)
		}

		return metadata.Schema{
			Type:     "array",
			Items:    &itemSchema,
			TypeName: "[]" + itemSchema.TypeName,
		}
	default:
		// For basic types, include default examples
		schema := metadata.Schema{
			Type:     getGoTypeSchema(t),
			TypeName: t.Name(),
		}

		// Set example only if a custom handler hasn't set one
		schema.Example = getExampleValue(t)
		return schema
	}
}

func getStructPropertiesWithCycle(t reflect.Type, visiting map[string]bool) (map[string]metadata.Schema, []string) {
	properties := make(map[string]metadata.Schema)
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

		if field.Type.Kind() == reflect.Ptr {
			schema := schemaFromTypeWithCycle(field.Type.Elem(), visiting)
			schema.Nullable = true
			schema.MinLength = minLen
			schema.MaxLength = maxLen
			schema.Minimum = min
			schema.Description = field.Tag.Get("description")
			properties[name] = schema
		} else {
			schema := schemaFromTypeWithCycle(field.Type, visiting)
			schema.MinLength = minLen
			schema.MaxLength = maxLen
			schema.Minimum = min
			schema.Description = field.Tag.Get("description")
			properties[name] = schema
		}
	}

	return properties, required
}

// getValidationRules returns validation rules for a field defined using struct tags with the `validate` key
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

// getExampleValue returns an appropriate example value for a Go type
// First checks if there's a custom type handler registered that provides an example
func getExampleValue(t reflect.Type) interface{} {
	typeName := t.String()
	if handler, exists := metadata.GetTypeHandler(typeName); exists {
		schema := handler(t)
		if schema.Example != nil {
			return schema.Example
		}
	}

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
	case reflect.Ptr:
		return getExampleValue(t.Elem())
	default:
		return nil
	}
}

func generateExample(t reflect.Type) interface{} {
	return generateExampleWithCycle(t, make(map[string]bool))
}

func generateExampleWithCycle(t reflect.Type, visiting map[string]bool) interface{} {
	if t.Kind() != reflect.Struct {
		return nil
	}

	typeID := t.String()
	if visiting[typeID] {
		return nil
	}

	visiting[typeID] = true
	defer func() {
		delete(visiting, typeID)
	}()

	example := make(map[string]interface{})
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

		var value interface{}
		switch field.Type.Kind() {
		case reflect.Struct:
			if field.Type.String() == "time.Time" {
				value = time.Now().Format(time.RFC3339)
			} else if field.Type.String() == "uuid.UUID" {
				value = "123e4567-e89b-12d3-a456-426614174000"
			} else {
				value = generateExampleWithCycle(field.Type, visiting)
			}
		case reflect.Ptr:
			// For pointer fields, generate an example of the underlying type
			elemType := field.Type.Elem()
			switch elemType.Kind() {
			case reflect.Struct:
				if elemType.String() == "time.Time" {
					value = time.Now().Format(time.RFC3339)
				} else if elemType.String() == "uuid.UUID" {
					value = "123e4567-e89b-12d3-a456-426614174000"
				} else {
					value = generateExampleWithCycle(elemType, visiting)
				}
			default:
				value = getExampleValue(elemType)
			}
		case reflect.Slice, reflect.Array:
			elemType := field.Type.Elem()
			if elemType.Kind() == reflect.Struct {
				if structExample := generateExampleWithCycle(elemType, visiting); structExample != nil {
					value = []interface{}{structExample}
				}
			} else if elemExample := getExampleValue(elemType); elemExample != nil {
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

// GetTypeFromGeneric extracts the reflect.Type from a generic type parameter T.
// This is a utility function to reduce the boilerplate of reflect.TypeOf((*T)(nil)).Elem().
func GetTypeFromGeneric[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

// IsArrayType checks if a reflect.Type is a slice or array.
// This is a utility function to reduce repeated slice/array checking.
func IsArrayType(t reflect.Type) bool {
	return t.Kind() == reflect.Slice || t.Kind() == reflect.Array
}

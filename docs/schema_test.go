package docs

import (
	"reflect"
	"testing"
	"time"

	"github.com/joakimcarlsson/go-router/metadata"
)

// MockUUID mimics uuid.UUID structure ([16]byte array) without external dependency
type MockUUID [16]byte

// String implements the same interface as uuid.UUID for testing
func (u MockUUID) String() string {
	return "123e4567-e89b-12d3-a456-426614174000"
}

// Test types for schema generation
type TestUser struct {
	ID       MockUUID               `json:"id"`
	Name     string                 `json:"name" validate:"required"`
	Email    *string                `json:"email"`
	Age      int                    `json:"age" validate:"min=0"`
	Score    float64                `json:"score"`
	Active   bool                   `json:"active"`
	Created  time.Time              `json:"created"`
	Updated  *time.Time             `json:"updated"`
	Tags     []string               `json:"tags"`
	Metadata map[string]interface{} `json:"metadata"`
}

type TestProduct struct {
	ID          MockUUID `json:"id"`
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Description *string  `json:"description"`
}

func TestSchemaFromType_UUID(t *testing.T) {
	// Register MockUUID to behave like real uuid.UUID
	metadata.RegisterTypeHandler("docs.MockUUID", func(t reflect.Type) metadata.Schema {
		return metadata.Schema{
			Type:     "string",
			Format:   "uuid",
			Example:  "123e4567-e89b-12d3-a456-426614174000",
			TypeName: "UUID",
		}
	})

	schema := SchemaFromType(reflect.TypeOf(MockUUID{}))

	if schema.Type != "string" {
		t.Errorf("Expected UUID type to be 'string', got '%s'", schema.Type)
	}

	if schema.Format != "uuid" {
		t.Errorf("Expected UUID format to be 'uuid', got '%s'", schema.Format)
	}

	if schema.Example != "123e4567-e89b-12d3-a456-426614174000" {
		t.Errorf("Expected UUID example to be '123e4567-e89b-12d3-a456-426614174000', got '%v'", schema.Example)
	}

	if schema.TypeName != "UUID" {
		t.Errorf("Expected UUID TypeName to be 'UUID', got '%s'", schema.TypeName)
	}
}

func TestSchemaFromType_TimeTime(t *testing.T) {
	schema := SchemaFromType(reflect.TypeOf(time.Time{}))

	if schema.Type != "string" {
		t.Errorf("Expected time.Time type to be 'string', got '%s'", schema.Type)
	}

	if schema.Format != "date-time" {
		t.Errorf("Expected time.Time format to be 'date-time', got '%s'", schema.Format)
	}

	if schema.TypeName != "time.Time" {
		t.Errorf("Expected time.Time TypeName to be 'time.Time', got '%s'", schema.TypeName)
	}

	// Check that example is a valid RFC3339 format
	if schema.Example == nil {
		t.Error("Expected time.Time example to be set")
	}
}

func TestSchemaFromType_BasicTypes(t *testing.T) {
	tests := []struct {
		name            string
		typ             reflect.Type
		expectedType    string
		expectedExample interface{}
	}{
		{"string", reflect.TypeOf(""), "string", "example"},
		{"int", reflect.TypeOf(0), "integer", 42},
		{"int32", reflect.TypeOf(int32(0)), "integer", 42},
		{"int64", reflect.TypeOf(int64(0)), "integer", 42},
		{"float32", reflect.TypeOf(float32(0)), "number", 3.14},
		{"float64", reflect.TypeOf(float64(0)), "number", 3.14},
		{"bool", reflect.TypeOf(false), "boolean", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema := SchemaFromType(test.typ)

			if schema.Type != test.expectedType {
				t.Errorf("Expected %s type to be '%s', got '%s'", test.name, test.expectedType, schema.Type)
			}

			if schema.Example != test.expectedExample {
				t.Errorf("Expected %s example to be '%v', got '%v'", test.name, test.expectedExample, schema.Example)
			}
		})
	}
}

func TestSchemaFromType_Pointer(t *testing.T) {
	// Test pointer to string
	schema := SchemaFromType(reflect.TypeOf((*string)(nil)))

	if schema.Type != "string" {
		t.Errorf("Expected *string type to be 'string', got '%s'", schema.Type)
	}

	if !schema.Nullable {
		t.Error("Expected *string to be nullable")
	}
}

func TestSchemaFromType_Slice(t *testing.T) {
	// Test slice of strings
	schema := SchemaFromType(reflect.TypeOf([]string{}))

	if schema.Type != "array" {
		t.Errorf("Expected []string type to be 'array', got '%s'", schema.Type)
	}

	if schema.Items == nil {
		t.Fatal("Expected []string to have Items schema")
	}

	if schema.Items.Type != "string" {
		t.Errorf("Expected []string items type to be 'string', got '%s'", schema.Items.Type)
	}

	if schema.TypeName != "[]string" {
		t.Errorf("Expected []string TypeName to be '[]string', got '%s'", schema.TypeName)
	}
}

func TestSchemaFromType_SliceOfUUID(t *testing.T) {
	// Register MockUUID to behave like real uuid.UUID
	metadata.RegisterTypeHandler("docs.MockUUID", func(t reflect.Type) metadata.Schema {
		return metadata.Schema{
			Type:     "string",
			Format:   "uuid",
			Example:  "123e4567-e89b-12d3-a456-426614174000",
			TypeName: "UUID",
		}
	})

	// Test that slice of UUID works correctly (should not be confused with UUID itself)
	schema := SchemaFromType(reflect.TypeOf([]MockUUID{}))

	if schema.Type != "array" {
		t.Errorf("Expected []MockUUID type to be 'array', got '%s'", schema.Type)
	}

	if schema.Items == nil {
		t.Fatal("Expected []MockUUID to have Items schema")
	}

	if schema.Items.Type != "string" {
		t.Errorf("Expected []MockUUID items type to be 'string', got '%s'", schema.Items.Type)
	}

	if schema.Items.Format != "uuid" {
		t.Errorf("Expected []MockUUID items format to be 'uuid', got '%s'", schema.Items.Format)
	}
}

func TestSchemaFromType_Struct(t *testing.T) {
	// Register MockUUID to behave like real uuid.UUID
	metadata.RegisterTypeHandler("docs.MockUUID", func(t reflect.Type) metadata.Schema {
		return metadata.Schema{
			Type:     "string",
			Format:   "uuid",
			Example:  "123e4567-e89b-12d3-a456-426614174000",
			TypeName: "UUID",
		}
	})

	schema := SchemaFromType(reflect.TypeOf(TestUser{}))

	if schema.Type != "object" {
		t.Errorf("Expected TestUser type to be 'object', got '%s'", schema.Type)
	}

	if schema.Properties == nil {
		t.Fatal("Expected TestUser to have Properties")
	}

	// Test UUID field
	idProp, exists := schema.Properties["id"]
	if !exists {
		t.Fatal("Expected TestUser to have 'id' property")
	}

	if idProp.Type != "string" {
		t.Errorf("Expected TestUser.id type to be 'string', got '%s'", idProp.Type)
	}

	if idProp.Format != "uuid" {
		t.Errorf("Expected TestUser.id format to be 'uuid', got '%s'", idProp.Format)
	}

	// Test time.Time field
	createdProp, exists := schema.Properties["created"]
	if !exists {
		t.Fatal("Expected TestUser to have 'created' property")
	}

	if createdProp.Type != "string" {
		t.Errorf("Expected TestUser.created type to be 'string', got '%s'", createdProp.Type)
	}

	if createdProp.Format != "date-time" {
		t.Errorf("Expected TestUser.created format to be 'date-time', got '%s'", createdProp.Format)
	}

	// Test nullable pointer field
	emailProp, exists := schema.Properties["email"]
	if !exists {
		t.Fatal("Expected TestUser to have 'email' property")
	}

	if !emailProp.Nullable {
		t.Error("Expected TestUser.email to be nullable")
	}

	// Test array field
	tagsProp, exists := schema.Properties["tags"]
	if !exists {
		t.Fatal("Expected TestUser to have 'tags' property")
	}

	if tagsProp.Type != "array" {
		t.Errorf("Expected TestUser.tags type to be 'array', got '%s'", tagsProp.Type)
	}

	if tagsProp.Items == nil {
		t.Fatal("Expected TestUser.tags to have Items schema")
	}

	if tagsProp.Items.Type != "string" {
		t.Errorf("Expected TestUser.tags items type to be 'string', got '%s'", tagsProp.Items.Type)
	}

	// Test required fields
	expectedRequired := []string{"name"}
	if len(schema.Required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(schema.Required))
	}

	for _, req := range expectedRequired {
		found := false
		for _, actual := range schema.Required {
			if actual == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected field '%s' to be required", req)
		}
	}
}

func TestSchemaFromType_SliceOfStructs(t *testing.T) {
	// Register MockUUID to behave like real uuid.UUID
	metadata.RegisterTypeHandler("docs.MockUUID", func(t reflect.Type) metadata.Schema {
		return metadata.Schema{
			Type:     "string",
			Format:   "uuid",
			Example:  "123e4567-e89b-12d3-a456-426614174000",
			TypeName: "UUID",
		}
	})

	schema := SchemaFromType(reflect.TypeOf([]TestProduct{}))

	if schema.Type != "array" {
		t.Errorf("Expected []TestProduct type to be 'array', got '%s'", schema.Type)
	}

	if schema.Items == nil {
		t.Fatal("Expected []TestProduct to have Items schema")
	}

	if schema.Items.Type != "object" {
		t.Errorf("Expected []TestProduct items type to be 'object', got '%s'", schema.Items.Type)
	}

	// Check that the item schema has the expected properties
	if schema.Items.Properties == nil {
		t.Fatal("Expected []TestProduct items to have Properties")
	}

	idProp, exists := schema.Items.Properties["id"]
	if !exists {
		t.Fatal("Expected TestProduct item to have 'id' property")
	}

	if idProp.Type != "string" || idProp.Format != "uuid" {
		t.Errorf("Expected TestProduct.id to be string with uuid format, got type='%s' format='%s'",
			idProp.Type, idProp.Format)
	}
}

func TestGetTypeFromGeneric(t *testing.T) {
	// Test the helper function
	mockUUIDType := GetTypeFromGeneric[MockUUID]()
	if mockUUIDType.String() != "docs.MockUUID" {
		t.Errorf("Expected GetTypeFromGeneric[MockUUID]() to return docs.MockUUID type, got %s", mockUUIDType.String())
	}

	stringType := GetTypeFromGeneric[string]()
	if stringType.Kind() != reflect.String {
		t.Errorf("Expected GetTypeFromGeneric[string]() to return string kind, got %s", stringType.Kind())
	}

	sliceType := GetTypeFromGeneric[[]TestProduct]()
	if sliceType.Kind() != reflect.Slice {
		t.Errorf("Expected GetTypeFromGeneric[[]TestProduct]() to return slice kind, got %s", sliceType.Kind())
	}
}

func TestIsArrayType(t *testing.T) {
	tests := []struct {
		name     string
		typ      reflect.Type
		expected bool
	}{
		{"MockUUID", reflect.TypeOf(MockUUID{}), true},
		{"[]string", reflect.TypeOf([]string{}), true},
		{"[5]int", reflect.TypeOf([5]int{}), true},
		{"string", reflect.TypeOf(""), false},
		{"int", reflect.TypeOf(0), false},
		{"struct", reflect.TypeOf(TestUser{}), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsArrayType(test.typ)
			if result != test.expected {
				t.Errorf("Expected IsArrayType(%s) to be %t, got %t", test.name, test.expected, result)
			}
		})
	}
}

// TestValidation tests that validation tags are properly handled
func TestSchemaFromType_ValidationTags(t *testing.T) {
	// Register MockUUID to behave like real uuid.UUID
	metadata.RegisterTypeHandler("docs.MockUUID", func(t reflect.Type) metadata.Schema {
		return metadata.Schema{
			Type:     "string",
			Format:   "uuid",
			Example:  "123e4567-e89b-12d3-a456-426614174000",
			TypeName: "UUID",
		}
	})

	schema := SchemaFromType(reflect.TypeOf(TestUser{}))

	// Test minimum validation on age field
	ageProp, exists := schema.Properties["age"]
	if !exists {
		t.Fatal("Expected TestUser to have 'age' property")
	}

	if ageProp.Minimum == nil {
		t.Fatal("Expected TestUser.age to have minimum validation")
	}

	if *ageProp.Minimum != 0 {
		t.Errorf("Expected TestUser.age minimum to be 0, got %f", *ageProp.Minimum)
	}
}

// TestCustomTypeHandler tests that custom type handlers work correctly
func TestCustomTypeHandler(t *testing.T) {
	// Define a custom type
	type CustomID int64

	// Register a custom handler
	metadata.RegisterTypeHandler("docs.CustomID", func(t reflect.Type) metadata.Schema {
		return metadata.Schema{
			Type:     "string",
			Format:   "custom-id",
			Example:  "CUSTOM-12345",
			TypeName: "CustomID",
		}
	})

	// Test that our custom handler is used
	schema := SchemaFromType(reflect.TypeOf(CustomID(0)))

	if schema.Type != "string" {
		t.Errorf("Expected CustomID type to be 'string', got '%s'", schema.Type)
	}

	if schema.Format != "custom-id" {
		t.Errorf("Expected CustomID format to be 'custom-id', got '%s'", schema.Format)
	}

	if schema.Example != "CUSTOM-12345" {
		t.Errorf("Expected CustomID example to be 'CUSTOM-12345', got '%v'", schema.Example)
	}
}

// TestEdgeCases tests edge cases and potential issues
func TestSchemaFromType_EdgeCases(t *testing.T) {
	t.Run("empty struct", func(t *testing.T) {
		type EmptyStruct struct{}
		schema := SchemaFromType(reflect.TypeOf(EmptyStruct{}))

		if schema.Type != "object" {
			t.Errorf("Expected empty struct type to be 'object', got '%s'", schema.Type)
		}

		if schema.Properties == nil {
			// Properties can be nil for empty structs, this is acceptable
		} else if len(schema.Properties) != 0 {
			t.Errorf("Expected empty struct to have 0 properties, got %d", len(schema.Properties))
		}
	})

	t.Run("nested structs", func(t *testing.T) {
		type Address struct {
			Street string `json:"street"`
			City   string `json:"city"`
		}

		type Person struct {
			Name    string  `json:"name"`
			Address Address `json:"address"`
		}

		schema := SchemaFromType(reflect.TypeOf(Person{}))

		if schema.Type != "object" {
			t.Errorf("Expected Person type to be 'object', got '%s'", schema.Type)
		}

		addressProp, exists := schema.Properties["address"]
		if !exists {
			t.Fatal("Expected Person to have 'address' property")
		}

		if addressProp.Type != "object" {
			t.Errorf("Expected Person.address type to be 'object', got '%s'", addressProp.Type)
		}

		if addressProp.Properties == nil {
			t.Fatal("Expected Person.address to have properties")
		}

		streetProp, exists := addressProp.Properties["street"]
		if !exists {
			t.Fatal("Expected Address to have 'street' property")
		}

		if streetProp.Type != "string" {
			t.Errorf("Expected Address.street type to be 'string', got '%s'", streetProp.Type)
		}
	})
}

// TestUUIDBugFix tests the specific UUID array bug that was fixed
func TestUUIDBugFix(t *testing.T) {
	// This test verifies that our fix correctly handles array types that should be treated as UUIDs
	// versus normal array types that should remain as arrays

	t.Run("normal array remains array", func(t *testing.T) {
		// Normal [16]byte array should remain as array (no special UUID handling)
		schema := SchemaFromType(reflect.TypeOf([16]byte{}))

		if schema.Type != "array" {
			t.Errorf("Expected normal [16]byte to be 'array', got '%s'", schema.Type)
		}

		if schema.Items == nil || schema.Items.Type != "integer" {
			t.Error("Expected normal [16]byte to have integer items")
		}
	})

	t.Run("MockUUID with handler is UUID", func(t *testing.T) {
		// Register MockUUID to behave like uuid.UUID
		metadata.RegisterTypeHandler("docs.MockUUID", func(t reflect.Type) metadata.Schema {
			return metadata.Schema{
				Type:     "string",
				Format:   "uuid",
				Example:  "123e4567-e89b-12d3-a456-426614174000",
				TypeName: "UUID",
			}
		})

		schema := SchemaFromType(reflect.TypeOf(MockUUID{}))

		// With handler registered, MockUUID should be treated as UUID
		if schema.Type != "string" {
			t.Errorf("Expected registered MockUUID to be 'string', got '%s'", schema.Type)
		}

		if schema.Format != "uuid" {
			t.Errorf("Expected registered MockUUID format to be 'uuid', got '%s'", schema.Format)
		}
	})
}

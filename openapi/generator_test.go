package openapi

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

// Test types for OpenAPI generation
type APIUser struct {
	ID      MockUUID   `json:"id"`
	Name    string     `json:"name"    validate:"required"`
	Email   *string    `json:"email"`
	Created time.Time  `json:"created"`
	Updated *time.Time `json:"updated"`
}

type APIProduct struct {
	ID    MockUUID `json:"id"`
	Name  string   `json:"name"`
	Price float64  `json:"price"`
}

// MockRouteInfo implements RouteInfo interface for testing
type MockRouteInfo struct {
	method      string
	path        string
	operationID string
	summary     string
	description string
	tags        []string
	parameters  []metadata.Parameter
	requestBody *metadata.RequestBody
	responses   map[string]metadata.Response
	security    []metadata.SecurityRequirement
	deprecated  bool
}

func (m *MockRouteInfo) Method() string { return m.method }

func (m *MockRouteInfo) Path() string { return m.path }

func (m *MockRouteInfo) OperationID() string { return m.operationID }

func (m *MockRouteInfo) Summary() string { return m.summary }

func (m *MockRouteInfo) Description() string { return m.description }

func (m *MockRouteInfo) Tags() []string { return m.tags }

func (m *MockRouteInfo) Parameters() []metadata.Parameter { return m.parameters }

func (m *MockRouteInfo) RequestBody() *metadata.RequestBody { return m.requestBody }

func (m *MockRouteInfo) Responses() map[string]metadata.Response { return m.responses }

func (m *MockRouteInfo) Security() []metadata.SecurityRequirement { return m.security }

func (m *MockRouteInfo) IsDeprecated() bool { return m.deprecated }

func setupMockUUID() {
	// Register MockUUID to behave like real uuid.UUID for all tests
	metadata.RegisterTypeHandler(
		"openapi.MockUUID",
		func(t reflect.Type) metadata.Schema {
			return metadata.Schema{
				Type:     "string",
				Format:   "uuid",
				Example:  "123e4567-e89b-12d3-a456-426614174000",
				TypeName: "UUID",
			}
		},
	)
}

func TestWithResponseType_UUID(t *testing.T) {
	setupMockUUID()

	// Test that WithResponseType correctly handles UUID fields
	var routeMetadata metadata.RouteMetadata

	// Apply the route option
	option := WithResponseType(200, "User details", APIUser{})
	option(&routeMetadata)

	// Check that responses were created
	if routeMetadata.Responses == nil {
		t.Fatal("Expected responses to be created")
	}

	response, exists := routeMetadata.Responses["200"]
	if !exists {
		t.Fatal("Expected 200 response to exist")
	}

	if response.Description != "User details" {
		t.Errorf(
			"Expected response description to be 'User details', got '%s'",
			response.Description,
		)
	}

	// Check JSON content
	jsonContent, exists := response.Content["application/json"]
	if !exists {
		t.Fatal("Expected JSON content to exist")
	}

	// The schema should use a reference for object types
	if jsonContent.SchemaRef == nil {
		t.Fatal("Expected schema reference for object type")
	}

	expectedRef := "#/components/schemas/APIUser"
	if jsonContent.SchemaRef.Ref != expectedRef {
		t.Errorf(
			"Expected schema reference to be '%s', got '%s'",
			expectedRef,
			jsonContent.SchemaRef.Ref,
		)
	}
}

func TestWithJSONResponseAdvanced_SliceOfUUID(t *testing.T) {
	setupMockUUID()

	// Test that arrays of UUID work correctly
	var routeMetadata metadata.RouteMetadata

	// Apply the route option for slice of APIUser
	option := WithJSONResponseAdvanced[[]APIUser](200, "List of users")
	option(&routeMetadata)

	// Check that responses were created
	if routeMetadata.Responses == nil {
		t.Fatal("Expected responses to be created")
	}

	response, exists := routeMetadata.Responses["200"]
	if !exists {
		t.Fatal("Expected 200 response to exist")
	}

	// Check JSON content
	jsonContent, exists := response.Content["application/json"]
	if !exists {
		t.Fatal("Expected JSON content to exist")
	}

	// For array types, the schema should be inline with items reference
	if jsonContent.Schema.Type != "array" {
		t.Errorf(
			"Expected schema type to be 'array', got '%s'",
			jsonContent.Schema.Type,
		)
	}

	if jsonContent.Schema.Items == nil {
		t.Fatal("Expected array schema to have items")
	}

	expectedRef := "#/components/schemas/APIUser"
	if jsonContent.Schema.Items.Ref != expectedRef {
		t.Errorf(
			"Expected items reference to be '%s', got '%s'",
			expectedRef,
			jsonContent.Schema.Items.Ref,
		)
	}
}

func TestWithRequestBody_UUID(t *testing.T) {
	setupMockUUID()

	// Test that request body with UUID works correctly
	var routeMetadata metadata.RouteMetadata

	// Apply the route option
	option := WithRequestBody("User data", true, APIUser{})
	option(&routeMetadata)

	// Check that request body was created
	if routeMetadata.RequestBody == nil {
		t.Fatal("Expected request body to be created")
	}

	if routeMetadata.RequestBody.Description != "User data" {
		t.Errorf(
			"Expected request body description to be 'User data', got '%s'",
			routeMetadata.RequestBody.Description,
		)
	}

	if !routeMetadata.RequestBody.Required {
		t.Error("Expected request body to be required")
	}

	// Check JSON content
	jsonContent, exists := routeMetadata.RequestBody.Content["application/json"]
	if !exists {
		t.Fatal("Expected JSON content to exist")
	}

	// Check that the schema has UUID field correctly typed
	if jsonContent.Schema.Type != "object" {
		t.Errorf(
			"Expected schema type to be 'object', got '%s'",
			jsonContent.Schema.Type,
		)
	}

	if jsonContent.Schema.Properties == nil {
		t.Fatal("Expected schema to have properties")
	}

	idProp, exists := jsonContent.Schema.Properties["id"]
	if !exists {
		t.Fatal("Expected schema to have 'id' property")
	}

	if idProp.Type != "string" {
		t.Errorf("Expected id type to be 'string', got '%s'", idProp.Type)
	}

	if idProp.Format != "uuid" {
		t.Errorf("Expected id format to be 'uuid', got '%s'", idProp.Format)
	}
}

func TestWithResponseExample_UUID(t *testing.T) {
	setupMockUUID()

	// Test that response examples work with UUID
	var routeMetadata metadata.RouteMetadata

	// Create a MockUUID with the expected value
	var mockID MockUUID
	copy(
		mockID[:],
		[]byte("123e4567-e89b-12d3"),
	) // Fill with some bytes for testing

	exampleUser := APIUser{
		ID:      mockID,
		Name:    "John Doe",
		Email:   nil,
		Created: time.Now(),
		Updated: nil,
	}

	// Apply the route option
	option := WithResponseExample(200, "User example", exampleUser)
	option(&routeMetadata)

	// Check that responses were created
	if routeMetadata.Responses == nil {
		t.Fatal("Expected responses to be created")
	}

	response, exists := routeMetadata.Responses["200"]
	if !exists {
		t.Fatal("Expected 200 response to exist")
	}

	// Check JSON content
	jsonContent, exists := response.Content["application/json"]
	if !exists {
		t.Fatal("Expected JSON content to exist")
	}

	// Check that example is set
	if jsonContent.Schema.Example == nil {
		t.Fatal("Expected schema to have example")
	}

	// Verify the example is the struct we provided
	exampleUser2, ok := jsonContent.Schema.Example.(APIUser)
	if !ok {
		t.Fatal("Expected example to be an APIUser struct")
	}

	// The ID should be the example UUID we set
	if exampleUser2.ID != exampleUser.ID {
		t.Errorf(
			"Expected example id to be '%v', got '%v'",
			exampleUser.ID,
			exampleUser2.ID,
		)
	}

	if exampleUser2.Name != exampleUser.Name {
		t.Errorf(
			"Expected example name to be '%s', got '%s'",
			exampleUser.Name,
			exampleUser2.Name,
		)
	}
}

func TestWithEmptyResponse(t *testing.T) {
	// Test simple response without content
	var routeMetadata metadata.RouteMetadata

	option := WithEmptyResponse(204, "No content")
	option(&routeMetadata)

	if routeMetadata.Responses == nil {
		t.Fatal("Expected responses to be created")
	}

	response, exists := routeMetadata.Responses["204"]
	if !exists {
		t.Fatal("Expected 204 response to exist")
	}

	if response.Description != "No content" {
		t.Errorf(
			"Expected response description to be 'No content', got '%s'",
			response.Description,
		)
	}

	// Empty response should have no content
	if len(response.Content) > 0 {
		t.Error("Expected empty response to have no content")
	}
}

func TestWithResponseSchema(t *testing.T) {
	// Test response with custom schema
	var routeMetadata metadata.RouteMetadata

	customSchema := metadata.Schema{
		Type:    "string",
		Format:  "uuid",
		Example: "123e4567-e89b-12d3-a456-426614174000",
	}

	option := WithResponseSchema(
		200,
		"Custom UUID",
		"application/json",
		customSchema,
	)
	option(&routeMetadata)

	if routeMetadata.Responses == nil {
		t.Fatal("Expected responses to be created")
	}

	response, exists := routeMetadata.Responses["200"]
	if !exists {
		t.Fatal("Expected 200 response to exist")
	}

	jsonContent, exists := response.Content["application/json"]
	if !exists {
		t.Fatal("Expected JSON content to exist")
	}

	if jsonContent.Schema.Type != "string" {
		t.Errorf(
			"Expected schema type to be 'string', got '%s'",
			jsonContent.Schema.Type,
		)
	}

	if jsonContent.Schema.Format != "uuid" {
		t.Errorf(
			"Expected schema format to be 'uuid', got '%s'",
			jsonContent.Schema.Format,
		)
	}
}

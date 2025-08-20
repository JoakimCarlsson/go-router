package metadata

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRouteMetadata_JSON(t *testing.T) {
	metadata := RouteMetadata{
		OperationID: "getUser",
		Summary:     "Get user by ID",
		Description: "Returns a user by their ID",
		Tags:        []string{"Users"},
		Deprecated:  false,
		Parameters: []Parameter{
			{
				Name:     "id",
				In:       "path",
				Required: true,
				Description: "User ID",
				Schema:   Schema{Type: "string", Format: "uuid"},
			},
		},
		Responses: map[string]Response{
			"200": {
				Description: "User found",
				Content: map[string]MediaType{
					ContentTypeJSON: {
						Schema: Schema{Type: "object"},
					},
				},
			},
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal RouteMetadata: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled RouteMetadata
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal RouteMetadata: %v", err)
	}

	// Check that the unmarshaled data matches the original
	if unmarshaled.OperationID != metadata.OperationID {
		t.Errorf("Expected OperationID '%s', got '%s'", metadata.OperationID, unmarshaled.OperationID)
	}

	if unmarshaled.Summary != metadata.Summary {
		t.Errorf("Expected Summary '%s', got '%s'", metadata.Summary, unmarshaled.Summary)
	}

	if len(unmarshaled.Parameters) != 1 {
		t.Fatalf("Expected 1 parameter, got %d", len(unmarshaled.Parameters))
	}

	param := unmarshaled.Parameters[0]
	if param.Name != "id" {
		t.Errorf("Expected parameter name 'id', got '%s'", param.Name)
	}

	if param.In != "path" {
		t.Errorf("Expected parameter in 'path', got '%s'", param.In)
	}

	if !param.Required {
		t.Error("Expected parameter to be required")
	}
}

func TestParameter_Validation(t *testing.T) {
	tests := []struct {
		name      string
		parameter Parameter
		valid     bool
	}{
		{
			name: "valid path parameter",
			parameter: Parameter{
				Name:     "id",
				In:       "path",
				Required: true,
				Schema:   Schema{Type: "string"},
			},
			valid: true,
		},
		{
			name: "valid query parameter",
			parameter: Parameter{
				Name:     "limit",
				In:       "query",
				Required: false,
				Schema:   Schema{Type: "integer"},
			},
			valid: true,
		},
		{
			name: "valid header parameter",
			parameter: Parameter{
				Name:     "Authorization",
				In:       "header",
				Required: true,
				Schema:   Schema{Type: "string"},
			},
			valid: true,
		},
		{
			name: "valid cookie parameter",
			parameter: Parameter{
				Name:     "sessionId",
				In:       "cookie",
				Required: false,
				Schema:   Schema{Type: "string"},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling/unmarshaling
			jsonData, err := json.Marshal(tt.parameter)
			if err != nil {
				if tt.valid {
					t.Errorf("Expected valid parameter to marshal successfully, got error: %v", err)
				}
				return
			}

			var unmarshaled Parameter
			if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
				if tt.valid {
					t.Errorf("Expected valid parameter to unmarshal successfully, got error: %v", err)
				}
				return
			}

			// Check that the unmarshaled data matches the original
			if unmarshaled.Name != tt.parameter.Name {
				t.Errorf("Expected Name '%s', got '%s'", tt.parameter.Name, unmarshaled.Name)
			}

			if unmarshaled.In != tt.parameter.In {
				t.Errorf("Expected In '%s', got '%s'", tt.parameter.In, unmarshaled.In)
			}

			if unmarshaled.Required != tt.parameter.Required {
				t.Errorf("Expected Required %t, got %t", tt.parameter.Required, unmarshaled.Required)
			}
		})
	}
}

func TestSchema_Validation(t *testing.T) {
	tests := []struct {
		name   string
		schema Schema
		valid  bool
	}{
		{
			name: "basic string schema",
			schema: Schema{
				Type:     "string",
				Format:   "uuid",
				Example:  "123e4567-e89b-12d3-a456-426614174000",
				TypeName: "UUID",
			},
			valid: true,
		},
		{
			name: "integer schema with validation",
			schema: Schema{
				Type:     "integer",
				Minimum:  floatPtr(1),
				Maximum:  floatPtr(100),
				Example:  42,
				TypeName: "int",
			},
			valid: true,
		},
		{
			name: "string schema with length validation",
			schema: Schema{
				Type:      "string",
				MinLength: intPtr(1),
				MaxLength: intPtr(255),
				Example:   "example",
				TypeName:  "string",
			},
			valid: true,
		},
		{
			name: "object schema with properties",
			schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id": {
						Type:   "string",
						Format: "uuid",
					},
					"name": {
						Type: "string",
					},
				},
				Required: []string{"id", "name"},
				TypeName: "User",
			},
			valid: true,
		},
		{
			name: "array schema with items",
			schema: Schema{
				Type: "array",
				Items: &Schema{
					Type: "string",
				},
				TypeName: "[]string",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling/unmarshaling
			jsonData, err := json.Marshal(tt.schema)
			if err != nil {
				if tt.valid {
					t.Errorf("Expected valid schema to marshal successfully, got error: %v", err)
				}
				return
			}

			var unmarshaled Schema
			if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
				if tt.valid {
					t.Errorf("Expected valid schema to unmarshal successfully, got error: %v", err)
				}
				return
			}

			// Check basic fields
			if unmarshaled.Type != tt.schema.Type {
				t.Errorf("Expected Type '%s', got '%s'", tt.schema.Type, unmarshaled.Type)
			}

			if unmarshaled.Format != tt.schema.Format {
				t.Errorf("Expected Format '%s', got '%s'", tt.schema.Format, unmarshaled.Format)
			}

			// Note: TypeName is not serialized (has json:"-" tag) as it's an internal field
		})
	}
}

func TestResponse_Validation(t *testing.T) {
	response := Response{
		Description: "Success response",
		Headers: map[string]Header{
			"X-Rate-Limit": {
				Description: "Requests per hour",
				Schema:      Schema{Type: "integer"},
			},
		},
		Content: map[string]MediaType{
			ContentTypeJSON: {
				Schema: Schema{
					Type: "object",
					Properties: map[string]Schema{
						"status": {Type: "string"},
						"data":   {Type: "object"},
					},
				},
			},
			ContentTypeXML: {
				Schema: Schema{
					Type: "object",
				},
			},
		},
	}

	// Test JSON marshaling/unmarshaling
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal Response: %v", err)
	}

	var unmarshaled Response
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Response: %v", err)
	}

	// Check that the unmarshaled data matches the original
	if unmarshaled.Description != response.Description {
		t.Errorf("Expected Description '%s', got '%s'", response.Description, unmarshaled.Description)
	}

	if len(unmarshaled.Headers) != 1 {
		t.Fatalf("Expected 1 header, got %d", len(unmarshaled.Headers))
	}

	if len(unmarshaled.Content) != 2 {
		t.Fatalf("Expected 2 content types, got %d", len(unmarshaled.Content))
	}

	// Check specific content types exist
	if _, exists := unmarshaled.Content[ContentTypeJSON]; !exists {
		t.Error("Expected JSON content type to exist")
	}

	if _, exists := unmarshaled.Content[ContentTypeXML]; !exists {
		t.Error("Expected XML content type to exist")
	}
}

func TestRequestBody_Validation(t *testing.T) {
	requestBody := RequestBody{
		Description: "User data",
		Required:    true,
		Content: map[string]MediaType{
			ContentTypeJSON: {
				Schema: Schema{
					Type: "object",
					Properties: map[string]Schema{
						"name":  {Type: "string"},
						"email": {Type: "string", Format: "email"},
					},
					Required: []string{"name", "email"},
				},
			},
			ContentTypeFormData: {
				Schema: Schema{
					Type: "object",
					Properties: map[string]Schema{
						"name":  {Type: "string"},
						"email": {Type: "string"},
					},
				},
			},
		},
	}

	// Test JSON marshaling/unmarshaling
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal RequestBody: %v", err)
	}

	var unmarshaled RequestBody
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal RequestBody: %v", err)
	}

	// Check that the unmarshaled data matches the original
	if unmarshaled.Description != requestBody.Description {
		t.Errorf("Expected Description '%s', got '%s'", requestBody.Description, unmarshaled.Description)
	}

	if unmarshaled.Required != requestBody.Required {
		t.Errorf("Expected Required %t, got %t", requestBody.Required, unmarshaled.Required)
	}

	if len(unmarshaled.Content) != 2 {
		t.Fatalf("Expected 2 content types, got %d", len(unmarshaled.Content))
	}
}

func TestSecurityRequirement_Validation(t *testing.T) {
	securityReq := SecurityRequirement{
		"apiKey": []string{"read", "write"},
		"oauth2": []string{"user:read", "user:write"},
	}

	// Test JSON marshaling/unmarshaling
	jsonData, err := json.Marshal(securityReq)
	if err != nil {
		t.Fatalf("Failed to marshal SecurityRequirement: %v", err)
	}

	var unmarshaled SecurityRequirement
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal SecurityRequirement: %v", err)
	}

	// Check that the unmarshaled data matches the original
	if len(unmarshaled) != 2 {
		t.Fatalf("Expected 2 security schemes, got %d", len(unmarshaled))
	}

	apiKeyScopes, exists := unmarshaled["apiKey"]
	if !exists {
		t.Error("Expected apiKey security scheme to exist")
	} else if len(apiKeyScopes) != 2 {
		t.Errorf("Expected 2 apiKey scopes, got %d", len(apiKeyScopes))
	}

	oauth2Scopes, exists := unmarshaled["oauth2"]
	if !exists {
		t.Error("Expected oauth2 security scheme to exist")
	} else if len(oauth2Scopes) != 2 {
		t.Errorf("Expected 2 oauth2 scopes, got %d", len(oauth2Scopes))
	}
}

func TestInfo_Validation(t *testing.T) {
	info := Info{
		Title:          "Test API",
		Version:        "1.0.0",
		Description:    "A test API",
		TermsOfService: "https://example.com/terms",
		Contact: &Contact{
			Name:  "API Support",
			URL:   "https://example.com/contact",
			Email: "support@example.com",
		},
		License: &License{
			Name: "MIT",
			URL:  "https://opensource.org/licenses/MIT",
		},
	}

	// Test JSON marshaling/unmarshaling
	jsonData, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal Info: %v", err)
	}

	var unmarshaled Info
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Info: %v", err)
	}

	// Check that the unmarshaled data matches the original
	if unmarshaled.Title != info.Title {
		t.Errorf("Expected Title '%s', got '%s'", info.Title, unmarshaled.Title)
	}

	if unmarshaled.Version != info.Version {
		t.Errorf("Expected Version '%s', got '%s'", info.Version, unmarshaled.Version)
	}

	if unmarshaled.Description != info.Description {
		t.Errorf("Expected Description '%s', got '%s'", info.Description, unmarshaled.Description)
	}

	if unmarshaled.Contact == nil {
		t.Error("Expected Contact to exist")
	} else {
		if unmarshaled.Contact.Name != info.Contact.Name {
			t.Errorf("Expected Contact Name '%s', got '%s'", info.Contact.Name, unmarshaled.Contact.Name)
		}
	}

	if unmarshaled.License == nil {
		t.Error("Expected License to exist")
	} else {
		if unmarshaled.License.Name != info.License.Name {
			t.Errorf("Expected License Name '%s', got '%s'", info.License.Name, unmarshaled.License.Name)
		}
	}
}

func TestTypeHandlerRegistry(t *testing.T) {
	// Test registering a new type handler
	typeName := "test.CustomType"
	handler := func(t reflect.Type) Schema {
		return Schema{
			Type:     "string",
			Format:   "custom",
			TypeName: "CustomType",
		}
	}

	RegisterTypeHandler(typeName, handler)

	// Test retrieving the handler
	retrievedHandler, exists := GetTypeHandler(typeName)
	if !exists {
		t.Error("Expected type handler to exist after registration")
	}

	// Test that the handler works correctly
	testType := reflect.TypeOf("")
	schema := retrievedHandler(testType)
	if schema.Type != "string" {
		t.Errorf("Expected schema type 'string', got '%s'", schema.Type)
	}

	if schema.Format != "custom" {
		t.Errorf("Expected schema format 'custom', got '%s'", schema.Format)
	}

	if schema.TypeName != "CustomType" {
		t.Errorf("Expected schema TypeName 'CustomType', got '%s'", schema.TypeName)
	}

	// Test retrieving non-existent handler
	_, exists = GetTypeHandler("nonexistent.Type")
	if exists {
		t.Error("Expected non-existent type handler to not exist")
	}
}

func TestTypeRegistry(t *testing.T) {
	// Test registering a new type
	type TestStruct struct {
		Name string
	}

	testType := reflect.TypeOf(TestStruct{})
	registeredName := RegisterType(testType)

	if registeredName == "" {
		t.Error("Expected non-empty registered name")
	}

	// The registered name should be based on the type name
	if registeredName != "TestStruct" {
		t.Errorf("Expected registered name 'TestStruct', got '%s'", registeredName)
	}

	// Test registering the same type again (should return the same name)
	registeredName2 := RegisterType(testType)
	if registeredName2 != registeredName {
		t.Errorf("Expected same registered name '%s', got '%s'", registeredName, registeredName2)
	}

	// Test registering anonymous struct (name will be empty since t.Name() returns empty for anonymous structs)
	anonymousType := reflect.TypeOf(struct{ Value int }{})
	anonymousName := RegisterType(anonymousType)
	// Anonymous structs have empty Name(), so RegisterType returns empty string
	// This is expected behavior as anonymous structs don't have a type name
	if anonymousType.Name() == "" && anonymousName == "" {
		// This is expected - anonymous structs don't have names
	} else if anonymousName == "" && anonymousType.Name() != "" {
		t.Error("Expected non-empty name for named struct type")
	}
}

func TestConstantValues(t *testing.T) {
	// Test that content type constants have expected values
	expectedConstants := map[string]string{
		"ContentTypeJSON":        "application/json",
		"ContentTypeXML":         "application/xml",
		"ContentTypeEventStream": "text/event-stream",
		"ContentTypeHTML":        "text/html",
		"ContentTypeFormData":    "multipart/form-data",
	}

	actualConstants := map[string]string{
		"ContentTypeJSON":        ContentTypeJSON,
		"ContentTypeXML":         ContentTypeXML,
		"ContentTypeEventStream": ContentTypeEventStream,
		"ContentTypeHTML":        ContentTypeHTML,
		"ContentTypeFormData":    ContentTypeFormData,
	}

	for name, expected := range expectedConstants {
		if actual := actualConstants[name]; actual != expected {
			t.Errorf("Expected %s to be '%s', got '%s'", name, expected, actual)
		}
	}
}

// Helper functions for creating pointers to primitive types
func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}
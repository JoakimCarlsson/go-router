package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/metadata"
	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swagger"
)

// Custom types that need special OpenAPI handling

// EmailAddress is a custom string type that should be validated as an email
type EmailAddress string

// PhoneNumber is a custom string type that should be validated as a phone number
type PhoneNumber string

// CustomID is a custom integer type
type CustomID int

// Money represents a monetary value with currency
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// GeoPoint represents a geographic point with latitude and longitude
type GeoPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// User represents a user in the system with custom field types
type User struct {
	ID        CustomID    `json:"id"`
	Email     EmailAddress `json:"email"`
	Phone     PhoneNumber  `json:"phone,omitempty"`
	CreatedAt time.Time    `json:"createdAt"`
	Location  *GeoPoint    `json:"location,omitempty"`
}

// Product represents a product in the catalog
type Product struct {
	ID          CustomID    `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Price       Money       `json:"price"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

// In-memory store for demo purposes
var (
	users    = map[CustomID]User{}
	products = map[CustomID]Product{}
	nextID   = CustomID(1001)
)

func init() {
	// Initialize with some sample data
	users[1001] = User{
		ID:        1001,
		Email:     "john.doe@example.com",
		Phone:     "+1-555-123-4567",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		Location: &GeoPoint{
			Latitude:  37.7749,
			Longitude: -122.4194,
		},
	}

	products[1001] = Product{
		ID:          1001,
		Name:        "Ergonomic Keyboard",
		Description: "A comfortable keyboard for long coding sessions",
		Price: Money{
			Amount:   129.99,
			Currency: "USD",
		},
		CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}
}

func main() {
	// Register custom type handlers before creating any routes
	registerTypeHandlers()
	
	// Create a new router
	r := router.New()
	
	// Setup the OpenAPI generator with info about the API
	gen := openapi.NewGenerator()
	gen.SetInfo("Custom Types API", "API demonstrating custom type handlers", "1.0.0")
	
	// User routes
	r.GET("/users", getAllUsers).
		With(
			openapi.WithDescription("Get all users"),
			openapi.WithTags("Users"),
			openapi.WithResponseType(http.StatusOK, "List of users", []User{}),
		)

	r.GET("/users/{id}", getUser).
		With(
			openapi.WithDescription("Get a user by ID"),
			openapi.WithTags("Users"),
			openapi.WithPathParam("id", "integer", "User ID", true),
			openapi.WithResponseType(http.StatusOK, "User data", User{}),
		)

	r.POST("/users", createUser).
		With(
			openapi.WithDescription("Create a new user"),
			openapi.WithTags("Users"),
			openapi.WithJSONRequestBody(User{}, "User data", true),
			openapi.WithResponseType(http.StatusCreated, "Created user", User{}),
		)

	// Product routes
	r.GET("/products", getAllProducts).
		With(
			openapi.WithDescription("Get all products"),
			openapi.WithTags("Products"),
			openapi.WithResponseType(http.StatusOK, "List of products", []Product{}),
		)

	r.GET("/products/{id}", getProduct).
		With(
			openapi.WithDescription("Get a product by ID"),
			openapi.WithTags("Products"),
			openapi.WithPathParam("id", "integer", "Product ID", true),
			openapi.WithResponseType(http.StatusOK, "Product data", Product{}),
		)

	r.POST("/products", createProduct).
		With(
			openapi.WithDescription("Create a new product"),
			openapi.WithTags("Products"),
			openapi.WithJSONRequestBody(Product{}, "Product data", true),
			openapi.WithResponseType(http.StatusCreated, "Created product", Product{}),
		)

	// Debug route to see the schema for a specific type
	r.GET("/debug/schema/{type}", debugSchema).
		With(
			openapi.WithDescription("Get OpenAPI schema for a specific type"),
			openapi.WithTags("Debug"),
			openapi.WithPathParam("type", "string", "Type name (email, phone, money, geopoint)", true),
			openapi.WithResponseType(http.StatusOK, "Schema information", map[string]interface{}{}),
		)

	// Generate OpenAPI documentation
	spec := gen.Generate(r.Routes)
	
	// Serve the Swagger UI
	r.Mount("/swagger", swagger.Handler(spec))
	
	// Serve the API
	fmt.Println("Starting server at :8080")
	fmt.Println("Access Swagger UI at http://localhost:8080/swagger")
	http.ListenAndServe(":8080", r)
}

// Register all custom type handlers
func registerTypeHandlers() {
	// Register handler for EmailAddress type
	metadata.RegisterTypeHandler("main.EmailAddress", func(t reflect.Type) metadata.Schema {
		return metadata.Schema{
			Type:        "string",
			Format:      "email",
			Example:     "user@example.com",
			Description: "Email address in standard format",
			TypeName:    "EmailAddress",
		}
	})
	
	// Register handler for PhoneNumber type
	metadata.RegisterTypeHandler("main.PhoneNumber", func(t reflect.Type) metadata.Schema {
		return metadata.Schema{
			Type:        "string",
			Format:      "phone",
			Pattern:     `^\+[1-9]\d{1,14}$`,
			Example:     "+1-555-123-4567",
			Description: "Phone number with optional country code",
			TypeName:    "PhoneNumber",
		}
	})
	
	// Register handler for CustomID type
	metadata.RegisterTypeHandler("main.CustomID", func(t reflect.Type) metadata.Schema {
		return metadata.Schema{
			Type:        "integer",
			Format:      "int64",
			Example:     1001,
			Description: "A unique identifier for this resource",
			Minimum:     func() *float64 { min := 1000.0; return &min }(),
			TypeName:    "CustomID",
		}
	})

	// Register handler for Money type
	metadata.RegisterTypeHandler("main.Money", func(t reflect.Type) metadata.Schema {
		// For struct types, we could either create a custom schema
		// or let the default struct handling work and just add some metadata
		return metadata.Schema{
			Type:        "object",
			Description: "A monetary value with currency",
			Properties: map[string]metadata.Schema{
				"amount": {
					Type:        "number",
					Format:      "double",
					Example:     99.99,
					Description: "The monetary amount",
				},
				"currency": {
					Type:        "string",
					Example:     "USD",
					Description: "The currency code (ISO 4217)",
					MinLength:   func() *int { min := 3; return &min }(),
					MaxLength:   func() *int { max := 3; return &max }(),
				},
			},
			Required:  []string{"amount", "currency"},
			Example:   map[string]interface{}{"amount": 99.99, "currency": "USD"},
			TypeName:  "Money",
		}
	})

	// GeoPoint is handled by the default struct handling,
	// but we could register a handler if we needed special behavior
}

// Handler functions

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	allUsers := make([]User, 0, len(users))
	for _, user := range users {
		allUsers = append(allUsers, user)
	}
	router.JSON(w, http.StatusOK, allUsers)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	idParam := router.PathParam(r, "id")
	id := CustomID(atoi(idParam))
	
	user, exists := users[id]
	if !exists {
		router.JSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}
	
	router.JSON(w, http.StatusOK, user)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		router.JSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	
	// Assign ID and creation time
	user.ID = nextID
	nextID++
	user.CreatedAt = time.Now()
	
	// Save to store
	users[user.ID] = user
	
	router.JSON(w, http.StatusCreated, user)
}

func getAllProducts(w http.ResponseWriter, r *http.Request) {
	allProducts := make([]Product, 0, len(products))
	for _, product := range products {
		allProducts = append(allProducts, product)
	}
	router.JSON(w, http.StatusOK, allProducts)
}

func getProduct(w http.ResponseWriter, r *http.Request) {
	idParam := router.PathParam(r, "id")
	id := CustomID(atoi(idParam))
	
	product, exists := products[id]
	if !exists {
		router.JSON(w, http.StatusNotFound, map[string]string{"error": "Product not found"})
		return
	}
	
	router.JSON(w, http.StatusOK, product)
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	var product Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		router.JSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	
	// Assign ID and timestamps
	product.ID = nextID
	nextID++
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	
	// Save to store
	products[product.ID] = product
	
	router.JSON(w, http.StatusCreated, product)
}

func debugSchema(w http.ResponseWriter, r *http.Request) {
	typeParam := router.PathParam(r, "type")
	
	var schema metadata.Schema
	
	switch typeParam {
	case "email":
		schema = docs.SchemaFromType(reflect.TypeOf(EmailAddress("")))
	case "phone":
		schema = docs.SchemaFromType(reflect.TypeOf(PhoneNumber("")))
	case "money":
		schema = docs.SchemaFromType(reflect.TypeOf(Money{}))
	case "geopoint":
		schema = docs.SchemaFromType(reflect.TypeOf(GeoPoint{}))
	case "customid":
		schema = docs.SchemaFromType(reflect.TypeOf(CustomID(0)))
	default:
		router.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "Unknown type. Use one of: email, phone, money, geopoint, customid",
		})
		return
	}
	
	router.JSON(w, http.StatusOK, schema)
}

// Helper function to convert string to int
func atoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
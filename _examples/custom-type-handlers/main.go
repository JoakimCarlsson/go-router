package main

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/integration"
	"github.com/joakimcarlsson/go-router/metadata"
	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
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
	ID        CustomID     `json:"id"`
	Email     EmailAddress `json:"email"`
	Phone     PhoneNumber  `json:"phone,omitempty"`
	CreatedAt time.Time    `json:"createdAt"`
	Location  *GeoPoint    `json:"location,omitempty"`
}

// Product represents a product in the catalog
type Product struct {
	ID          CustomID  `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       Money     `json:"price"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
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
	generator := openapi.NewGenerator(metadata.Info{
		Title:       "Product Catalog API",
		Version:     "1.0.0",
		Description: "A sample product catalog API built with go-router",
	})

	// User routes
	r.GET("/users", getAllUsers,
		docs.WithTags("Users"),
		docs.WithSummary("Get all users"),
		docs.WithDescription("Retrieves a list of all users in the system"),
		docs.WithJSONResponse[[]User](http.StatusOK, "List of users"),
	)

	r.GET("/users/{id}", getUser,
		docs.WithTags("Users"),
		docs.WithSummary("Get user by ID"),
		docs.WithDescription("Retrieves a single user by their ID"),
		docs.WithPathParam("id", "integer", true, "User ID", 1001),
		docs.WithJSONResponse[User](http.StatusOK, "User data"),
		docs.WithResponse(http.StatusNotFound, "User not found"),
	)

	r.POST("/users", createUser,
		docs.WithTags("Users"),
		docs.WithSummary("Create a new user"),
		docs.WithDescription("Creates a new user in the system"),
		docs.WithJSONRequestBody[User](true, "User data to create"),
		docs.WithJSONResponse[User](http.StatusCreated, "Created user"),
		docs.WithResponse(http.StatusBadRequest, "Invalid request data"),
	)

	r.GET("/products", getAllProducts,
		docs.WithTags("Products"),
		docs.WithSummary("Get all products"),
		docs.WithDescription("Retrieves a list of all products in the catalog"),
		docs.WithJSONResponse[[]Product](http.StatusOK, "List of products"),
	)

	r.GET("/products/{id}", getProduct,
		docs.WithTags("Products"),
		docs.WithSummary("Get product by ID"),
		docs.WithDescription("Retrieves a single product by its ID"),
		docs.WithPathParam("id", "integer", true, "Product ID", 1001),
		docs.WithJSONResponse[Product](http.StatusOK, "Product data"),
		docs.WithResponse(http.StatusNotFound, "Product not found"),
	)

	r.POST("/products", createProduct,
		docs.WithTags("Products"),
		docs.WithSummary("Create a new product"),
		docs.WithDescription("Creates a new product in the catalog"),
		docs.WithJSONRequestBody[Product](true, "Product data to create"),
		docs.WithJSONResponse[Product](http.StatusCreated, "Created product"),
		docs.WithResponse(http.StatusBadRequest, "Invalid request data"),
	)

	// Serve the Swagger UI
	swaggerUI := integration.NewSwaggerUIIntegration(r, generator)
	swaggerUI.SetupRoutes(r, "/openapi.json", "/docs")

	// Serve the API
	fmt.Println("Starting server at :8080")
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
			Required: []string{"amount", "currency"},
			Example:  map[string]interface{}{"amount": 99.99, "currency": "USD"},
			TypeName: "Money",
		}
	})

	// GeoPoint is handled by the default struct handling,
	// but we could register a handler if we needed special behavior
}

// Handler functions

// getAllUsers handles GET /users
func getAllUsers(c *router.Context) {
	allUsers := make([]User, 0, len(users))
	for _, user := range users {
		allUsers = append(allUsers, user)
	}
	c.JSON(http.StatusOK, allUsers)
}

// getUser handles GET /users/{id}
func getUser(c *router.Context) {
	idParam := c.Param("id")
	id := CustomID(atoi(idParam))

	user, exists := users[id]
	if !exists {
		c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// createUser handles POST /users
func createUser(c *router.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Assign ID and creation time
	user.ID = nextID
	nextID++
	user.CreatedAt = time.Now()

	// Save to store
	users[user.ID] = user

	c.JSON(http.StatusCreated, user)
}

// getAllProducts handles GET /products
func getAllProducts(c *router.Context) {
	allProducts := make([]Product, 0, len(products))
	for _, product := range products {
		allProducts = append(allProducts, product)
	}
	c.JSON(http.StatusOK, allProducts)
}

// getProduct handles GET /products/{id}
func getProduct(c *router.Context) {
	idParam := c.Param("id")
	id := CustomID(atoi(idParam))

	product, exists := products[id]
	if !exists {
		c.JSON(http.StatusNotFound, map[string]string{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// createProduct handles POST /products
func createProduct(c *router.Context) {
	var product Product
	if err := c.BindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Assign ID and timestamps
	product.ID = nextID
	nextID++
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	// Save to store
	products[product.ID] = product

	c.JSON(http.StatusCreated, product)
}

// Helper function to convert string to int
func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

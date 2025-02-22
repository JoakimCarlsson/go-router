package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
)

type Todo struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type PaginatedResponse struct {
	Items      []Todo `json:"items"`
	TotalCount int    `json:"totalCount"`
	Skip       int    `json:"skip"`
	Take       int    `json:"take"`
}

func main() {
	r := router.New()

	info := openapi.Info{
		Title:       "Todo API",
		Description: "API for managing todos in the system",
		Version:     "1.0.0",
		Contact: openapi.Contact{
			Name:  "API Support",
			Email: "support@example.com",
		},
	}
	generator := openapi.NewGenerator(info)

	generator.WithSecurityScheme("bearerAuth", openapi.SecurityScheme{
		Type:        "http",
		Scheme:      "bearer",
		Description: "JWT Bearer token authentication",
	})

	r.Group("/v1", func(v1 *router.Router) {
		v1.Group("/todos", func(todos *router.Router) {
			todos.WithTags("Todos").
				WithSecurity(map[string][]string{"bearerAuth": {}})

			todos.GET("", listTodos,
				openapi.WithSummary("List all todos"),
				openapi.WithDescription("Returns a paginated list of todos"),
				openapi.WithParameter("skip", "query", "integer", false, "Number of items to skip"),
				openapi.WithParameter("take", "query", "integer", false, "Number of items to take"),
				openapi.WithResponseType("200", "Successfully retrieved todos", PaginatedResponse{}),
			)

			todos.GET("/{id}", getTodo,
				openapi.WithSummary("Get a todo by ID"),
				openapi.WithDescription("Returns a single todo by its ID"),
				openapi.WithParameter("id", "path", "integer", true, "Todo ID"),
				openapi.WithResponseType("200", "Todo found", Todo{}),
				openapi.WithResponseType("404", "Todo not found", ErrorResponse{}),
			)

			todos.POST("", createTodo,
				openapi.WithSummary("Create a new todo"),
				openapi.WithDescription("Creates a new todo in the system"),
				openapi.WithRequestBody("Todo information to create", true, Todo{}),
				openapi.WithResponseType("201", "Todo created", Todo{}),
				openapi.WithEmptyResponse("400", "Invalid request"),
			)

			todos.PUT("/{id}", updateTodo,
				openapi.WithSummary("Update a todo"),
				openapi.WithDescription("Updates an existing todo"),
				openapi.WithParameter("id", "path", "integer", true, "Todo ID"),
				openapi.WithRequestBody("Todo information to update", true, Todo{}),
				openapi.WithResponseType("200", "Todo updated", Todo{}),
				openapi.WithResponseType("404", "Todo not found", ErrorResponse{}),
			)

			todos.DELETE("/{id}", deleteTodo,
				openapi.WithSummary("Delete a todo"),
				openapi.WithDescription("Deletes a todo by its ID"),
				openapi.WithParameter("id", "path", "integer", true, "Todo ID"),
				openapi.WithEmptyResponse("204", "Todo deleted"),
				openapi.WithResponseType("404", "Todo not found", ErrorResponse{}),
			)
		})
	})

	r.GET("/openapi.json", r.ServeOpenAPI(generator))

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("OpenAPI documentation available at http://localhost:8080/openapi.json")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func listTodos(c *router.Context) {
	skip, _ := strconv.Atoi(c.QueryDefault("skip", "0"))
	take, _ := strconv.Atoi(c.QueryDefault("take", "10"))

	todos := []Todo{
		{ID: 1, Title: "Learn Go", Description: "Study Go programming language", Completed: true, CreatedAt: time.Now()},
		{ID: 2, Title: "Build API", Description: "Create REST API with go-router", Completed: false, CreatedAt: time.Now()},
		{ID: 3, Title: "Write Tests", Description: "Add unit tests for the API", Completed: false, CreatedAt: time.Now()},
	}

	totalCount := len(todos)

	end := skip + take
	if end > len(todos) {
		end = len(todos)
	}
	if skip >= len(todos) {
		skip = 0
		end = 0
	}

	paginatedTodos := todos[skip:end]

	response := PaginatedResponse{
		Items:      paginatedTodos,
		TotalCount: totalCount,
		Skip:       skip,
		Take:       take,
	}

	c.JSON(200, response)
}

func getTodo(c *router.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, ErrorResponse{Error: "invalid todo ID"})
		return
	}

	todo := Todo{
		ID:          id,
		Title:       "Sample Todo",
		Description: "This is a sample todo",
		Completed:   false,
		CreatedAt:   time.Now(),
	}

	c.JSON(200, todo)
}

func createTodo(c *router.Context) {
	var newTodo Todo
	if err := c.BindJSON(&newTodo); err != nil {
		c.JSON(400, ErrorResponse{Error: "invalid request body"})
		return
	}

	newTodo.ID = 1
	newTodo.CreatedAt = time.Now()

	c.JSON(201, newTodo)
}

func updateTodo(c *router.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, ErrorResponse{Error: "invalid todo ID"})
		return
	}

	var updatedTodo Todo
	if err := c.BindJSON(&updatedTodo); err != nil {
		c.JSON(400, ErrorResponse{Error: "invalid request body"})
		return
	}

	updatedTodo.ID = id

	c.JSON(200, updatedTodo)
}

func deleteTodo(c *router.Context) {
	_, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, ErrorResponse{Error: "invalid todo ID"})
		return
	}
	c.Status(204)
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joakimcarlsson/go-router/router"
)

// Task represents a simple task
type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"createdAt"`
}

// Global tasks storage
var tasks = []Task{
	{
		ID:        "1",
		Title:     "Learn Go",
		Completed: true,
		CreatedAt: time.Now().Add(-48 * time.Hour),
	},
	{
		ID:        "2",
		Title:     "Build a REST API",
		Completed: false,
		CreatedAt: time.Now().Add(-24 * time.Hour),
	},
	{
		ID:        "3",
		Title:     "Deploy to production",
		Completed: false,
		CreatedAt: time.Now(),
	},
}

func main() {
	r := router.New()

	// Add a simple recovery middleware
	r.Use(recoveryMiddleware)

	// Define routes without any documentation
	r.GET("/", home)
	r.GET("/tasks", listTasks)
	r.GET("/tasks/{id}", getTask)
	r.POST("/tasks", createTask)
	r.PUT("/tasks/{id}", updateTask)
	r.DELETE("/tasks/{id}", deleteTask)

	// Group routes by version
	r.Group("/v2", func(r *router.Router) {
		r.GET("/tasks", listTasksV2)
	})

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// Simple home handler
func home(c *router.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"message": "Welcome to the minimal API example",
		"version": "1.0",
		"docs":    "No documentation available - this is a minimal example",
	})
}

// Handler to list all tasks
func listTasks(c *router.Context) {
	c.JSON(http.StatusOK, tasks)
}

// Enhanced version of listTasks with filtering
func listTasksV2(c *router.Context) {
	// Check for completed filter query parameter
	completed := c.QueryBoolDefault("completed", false)
	showAll := c.QueryBoolDefault("all", false)

	if showAll {
		c.JSON(http.StatusOK, tasks)
		return
	}

	// Filter tasks
	filtered := make([]Task, 0)
	for _, task := range tasks {
		if task.Completed == completed {
			filtered = append(filtered, task)
		}
	}

	c.JSON(http.StatusOK, filtered)
}

// Handler to get a single task
func getTask(c *router.Context) {
	id := c.Param("id")

	for _, task := range tasks {
		if task.ID == id {
			c.JSON(http.StatusOK, task)
			return
		}
	}

	c.JSON(http.StatusNotFound, map[string]string{
		"error": fmt.Sprintf("Task with ID %s not found", id),
	})
}

// Handler to create a new task
func createTask(c *router.Context) {
	var newTask struct {
		Title string `json:"title"`
	}

	if err := c.BindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	if newTask.Title == "" {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Title is required",
		})
		return
	}

	task := Task{
		ID:        fmt.Sprintf("%d", len(tasks)+1),
		Title:     newTask.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}

	tasks = append(tasks, task)

	c.JSON(http.StatusCreated, task)
}

// Handler to update a task
func updateTask(c *router.Context) {
	id := c.Param("id")

	var update struct {
		Title     *string `json:"title"`
		Completed *bool   `json:"completed"`
	}

	if err := c.BindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	for i, task := range tasks {
		if task.ID == id {
			if update.Title != nil {
				tasks[i].Title = *update.Title
			}

			if update.Completed != nil {
				tasks[i].Completed = *update.Completed
			}

			c.JSON(http.StatusOK, tasks[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, map[string]string{
		"error": fmt.Sprintf("Task with ID %s not found", id),
	})
}

// Handler to delete a task
func deleteTask(c *router.Context) {
	id := c.Param("id")

	for i, task := range tasks {
		if task.ID == id {
			// Remove task from slice
			tasks = append(tasks[:i], tasks[i+1:]...)
			c.Status(http.StatusNoContent)
			return
		}
	}

	c.JSON(http.StatusNotFound, map[string]string{
		"error": fmt.Sprintf("Task with ID %s not found", id),
	})
}

// Simple recovery middleware
func recoveryMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(c *router.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered: %v", r)
				c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Internal server error",
				})
			}
		}()

		next(c)
	}
}

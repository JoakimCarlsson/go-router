package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joakimcarlsson/go-router/router"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"createdAt"`
}

var tasks = []Task{
	{ID: "1", Title: "Learn Go", Completed: true, CreatedAt: time.Now().Add(-48 * time.Hour)},
	{ID: "2", Title: "Build a REST API", Completed: false, CreatedAt: time.Now().Add(-24 * time.Hour)},
	{ID: "3", Title: "Deploy to production", Completed: false, CreatedAt: time.Now()},
}

func main() {
	r := router.New()

	r.Use(recoveryMiddleware)

	r.GET("/", home)
	r.GET("/tasks", listTasks)
	r.GET("/tasks/{id}", getTask)
	r.POST("/tasks", createTask)
	r.PUT("/tasks/{id}", updateTask)
	r.DELETE("/tasks/{id}", deleteTask)

	r.Group("/v2", func(r *router.Router) {
		r.GET("/tasks", listTasksV2)
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}

func home(c *router.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"message": "Welcome to the minimal API example",
		"version": "1.0",
	})
}

func listTasks(c *router.Context) {
	c.JSON(http.StatusOK, tasks)
}

func listTasksV2(c *router.Context) {
	completed := c.QueryBoolDefault("completed", false)
	showAll := c.QueryBoolDefault("all", false)

	if showAll {
		c.JSON(http.StatusOK, tasks)
		return
	}

	filtered := make([]Task, 0)
	for _, task := range tasks {
		if task.Completed == completed {
			filtered = append(filtered, task)
		}
	}
	c.JSON(http.StatusOK, filtered)
}

func getTask(c *router.Context) {
	id := c.Param("id")
	for _, task := range tasks {
		if task.ID == id {
			c.JSON(http.StatusOK, task)
			return
		}
	}
	c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Task %s not found", id)})
}

func createTask(c *router.Context) {
	var newTask struct {
		Title string `json:"title"`
	}
	if err := c.BindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	if newTask.Title == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Title is required"})
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

func updateTask(c *router.Context) {
	id := c.Param("id")
	var update struct {
		Title     *string `json:"title"`
		Completed *bool   `json:"completed"`
	}
	if err := c.BindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
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
	c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Task %s not found", id)})
}

func deleteTask(c *router.Context) {
	id := c.Param("id")
	for i, task := range tasks {
		if task.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			c.Status(http.StatusNoContent)
			return
		}
	}
	c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Task %s not found", id)})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

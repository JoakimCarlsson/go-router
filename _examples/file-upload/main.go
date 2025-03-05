package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/integration"
	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
)

// FileInfo represents metadata about an uploaded file
type FileInfo struct {
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
	StoredAt    string `json:"storedAt,omitempty"`
}

// UploadResponse represents the response from the upload endpoint
type UploadResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Files   []FileInfo `json:"files,omitempty"`
}

var uploadDir = "./uploads"

// setupRoutes configures all the routes for the application
func setupRoutes(r *router.Router) {
	// Configure max memory for multipart forms (10 MB in this example)
	r.WithMultipartConfig(10 << 20)

	// Create upload directory if it doesn't exist
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	// Single file upload endpoint
	r.POST("/upload/file", uploadSingleFile,
		docs.WithSummary("Upload a single file"),
		docs.WithDescription("Upload a single file with optional metadata"),
		docs.WithMultipartFormData(true, "File to upload", map[string]string{
			"file": "The file to upload",
		}),
		docs.WithJSONResponse[UploadResponse](http.StatusCreated, "File uploaded successfully"),
		docs.WithResponse(http.StatusBadRequest, "Invalid request"),
		docs.WithResponse(http.StatusInternalServerError, "Server error"),
	)

	// Multiple file upload endpoint
	r.POST("/upload/files", uploadMultipleFiles,
		docs.WithSummary("Upload multiple files"),
		docs.WithDescription("Upload multiple files in a single request"),
		docs.WithMultipartFormData(true, "Files to upload", map[string]string{
			"files[]": "Multiple files to upload",
		}),
		docs.WithJSONResponse[UploadResponse](http.StatusCreated, "Files uploaded successfully"),
		docs.WithResponse(http.StatusBadRequest, "Invalid request"),
		docs.WithResponse(http.StatusInternalServerError, "Server error"),
	)

	// List uploaded files endpoint
	r.GET("/files", listFiles,
		docs.WithSummary("List uploaded files"),
		docs.WithDescription("Get a list of all files that have been uploaded"),
		docs.WithResponse(http.StatusOK, "List of files"),
		docs.WithResponse(http.StatusInternalServerError, "Server error"),
	)
}

func main() {
	r := router.New()

	generator := openapi.NewGenerator(openapi.Info{
		Title:       "File Upload API",
		Version:     "1.0.0",
		Description: "API for uploading and managing files using multipart form data",
	})

	generator.WithServer("http://localhost:8080", "Development server")

	swaggerUI := integration.NewSwaggerUIIntegration(r, generator)
	swaggerUI.SetupRoutes(r, "/openapi.json", "/docs")

	setupRoutes(r)

	port := "8080"
	log.Printf("Server starting on port %s...", port)
	log.Printf("API documentation available at http://localhost:%s/docs", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// uploadSingleFile handles uploading a single file
func uploadSingleFile(c *router.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: "No file provided or invalid request: " + err.Error(),
		})
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", file.Size, ext)
	dst := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "Failed to save the file: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, UploadResponse{
		Success: true,
		Message: "File uploaded successfully",
		Files: []FileInfo{
			{
				Filename:    file.Filename,
				Size:        file.Size,
				ContentType: file.Header.Get("Content-Type"),
				StoredAt:    dst,
			},
		},
	})
}

// uploadMultipleFiles handles uploading multiple files
func uploadMultipleFiles(c *router.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: "Invalid multipart form: " + err.Error(),
		})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: "No files provided",
		})
		return
	}

	var fileInfos []FileInfo
	for _, file := range files {
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%d_%s%s", file.Size, strings.ReplaceAll(file.Filename, " ", "_"), ext)
		dst := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, UploadResponse{
				Success: false,
				Message: "Failed to save file: " + err.Error(),
				Files:   fileInfos,
			})
			return
		}

		fileInfos = append(fileInfos, FileInfo{
			Filename:    file.Filename,
			Size:        file.Size,
			ContentType: file.Header.Get("Content-Type"),
			StoredAt:    dst,
		})
	}

	c.JSON(http.StatusCreated, UploadResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully uploaded %d files", len(fileInfos)),
		Files:   fileInfos,
	})
}

// listFiles returns a list of all uploaded files
func listFiles(c *router.Context) {
	files, err := os.ReadDir(uploadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to read uploads directory",
		})
		return
	}

	var fileList []string
	for _, file := range files {
		if !file.IsDir() {
			fileList = append(fileList, file.Name())
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"files": fileList,
		"count": len(fileList),
	})
}

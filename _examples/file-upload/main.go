package main

import (
	"fmt"
	"log"
	"mime/multipart"
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
	Description string `json:"description,omitempty"`
}

// UploadResponse represents the response from the upload endpoint
type UploadResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Files   []FileInfo `json:"files,omitempty"`
}

// SingleUpload represents a single file upload with metadata
type SingleUpload struct {
	File        *multipart.FileHeader `form:"file" file:"true"`
	Name        string                `form:"name"`
	Description string                `form:"description"`
}

// MultiUpload represents a multiple file upload with metadata
type MultiUpload struct {
	Files    []*multipart.FileHeader `form:"files" file:"true"`
	Category string                  `form:"category"`
	Tags     string                  `form:"tags"`
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
		docs.WithDescription("Upload a single file with metadata"),
		docs.WithMultipartFormData("File to upload with metadata", map[string]docs.FormFieldSpec{
			"file": {
				Type:        "file",
				Description: "The file to upload",
				Required:    true,
			},
			"name": {
				Type:        "string",
				Description: "Name of the file",
				Required:    false,
			},
			"description": {
				Type:        "string",
				Description: "Description of the file",
				Required:    false,
			},
		}),
		docs.WithJSONResponse[UploadResponse](http.StatusCreated, "File uploaded successfully"),
		docs.WithResponse(http.StatusBadRequest, "Invalid request"),
		docs.WithResponse(http.StatusInternalServerError, "Server error"),
	)

	// Multiple file upload endpoint
	r.POST("/upload/files", uploadMultipleFiles,
		docs.WithSummary("Upload multiple files"),
		docs.WithDescription("Upload multiple files in a single request"),
		docs.WithMultipartFormData("Files to upload", map[string]docs.FormFieldSpec{
			"files": {
				Type:        "file[]",
				Description: "Multiple files to upload",
				Required:    true,
			},
			"category": {
				Type:        "string",
				Description: "Category for all files",
				Required:    false,
			},
			"tags": {
				Type:        "string",
				Description: "Comma-separated tags for the files",
				Required:    false,
			},
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

// uploadSingleFile handles uploading a single file with metadata
func uploadSingleFile(c *router.Context) {
	var upload SingleUpload
	if err := c.BindForm(&upload); err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: "Invalid form data: " + err.Error(),
		})
		return
	}

	if upload.File == nil {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: "No file provided",
		})
		return
	}

	// Create filename
	ext := filepath.Ext(upload.File.Filename)
	filename := upload.File.Filename
	if upload.Name != "" {
		filename = upload.Name + ext
	}

	// Make sure filename is safe for filesystem
	safeName := strings.ReplaceAll(filename, " ", "_")
	dst := filepath.Join(uploadDir, safeName)

	// Save the file
	if err := c.SaveUploadedFile(upload.File, dst); err != nil {
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
				Filename:    upload.File.Filename,
				Size:        upload.File.Size,
				ContentType: upload.File.Header.Get("Content-Type"),
				StoredAt:    dst,
				Description: upload.Description,
			},
		},
	})
}

// uploadMultipleFiles handles uploading multiple files with metadata
func uploadMultipleFiles(c *router.Context) {
	var upload MultiUpload
	if err := c.BindForm(&upload); err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: "Invalid form data: " + err.Error(),
		})
		return
	}

	if len(upload.Files) == 0 {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: "No files provided",
		})
		return
	}

	categoryMsg := ""
	if upload.Category != "" {
		categoryMsg = fmt.Sprintf(" in category '%s'", upload.Category)
	}
	tagsMsg := ""
	if upload.Tags != "" {
		tagsMsg = fmt.Sprintf(" with tags: %s", upload.Tags)
	}

	var fileInfos []FileInfo
	for _, file := range upload.Files {
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%s_%s%s",
			strings.ReplaceAll(upload.Category, " ", "_"),
			strings.ReplaceAll(file.Filename, " ", "_"),
			ext)
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
			Description: fmt.Sprintf("Category: %s, Tags: %s", upload.Category, upload.Tags),
		})
	}

	c.JSON(http.StatusCreated, UploadResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully uploaded %d files%s%s", len(fileInfos), categoryMsg, tagsMsg),
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

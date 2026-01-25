package main

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swaggerui"
)

type FileInfo struct {
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
	StoredAt    string `json:"storedAt,omitempty"`
	Description string `json:"description,omitempty"`
}

type UploadResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Files   []FileInfo `json:"files,omitempty"`
}

type SingleUpload struct {
	File        *multipart.FileHeader `form:"file" file:"true" required:"true" description:"The file to upload"`
	Name        string                `form:"name" description:"Optional name for the file"`
	Description string                `form:"description" description:"Description of the uploaded file"`
}

type MultiUpload struct {
	Files    []*multipart.FileHeader `form:"files" file:"true" required:"true" description:"Multiple files to upload"`
	Category string                  `form:"category" description:"Category for all files"`
	Tags     string                  `form:"tags" description:"Comma-separated tags for the files"`
}

var uploadDir = "./uploads"

func main() {
	r := router.New()
	r.WithMultipartConfig(10 << 20)

	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	r.POST("/upload/file", uploadSingleFile,
		openapi.WithSummary("Upload a single file"),
		openapi.WithMultipartFormStruct[SingleUpload]("File to upload with metadata"),
		openapi.WithJSONResponse[UploadResponse](http.StatusCreated, "File uploaded successfully"),
	)

	r.POST("/upload/files", uploadMultipleFiles,
		openapi.WithSummary("Upload multiple files"),
		openapi.WithMultipartFormStruct[MultiUpload]("Files to upload"),
		openapi.WithJSONResponse[UploadResponse](http.StatusCreated, "Files uploaded successfully"),
	)

	r.GET("/files", listFiles,
		openapi.WithSummary("List uploaded files"),
	)

	generator := openapi.NewGenerator(openapi.Info{
		Title:       "File Upload API",
		Version:     "1.0.0",
		Description: "API for uploading and managing files",
	})

	setup := swaggerui.NewSetup(r, generator)
	setup.RegisterRoutes(r, "/openapi.json", "/docs")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func uploadSingleFile(c *router.Context) {
	var upload SingleUpload
	if err := c.BindForm(&upload); err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{Success: false, Message: "Invalid form data: " + err.Error()})
		return
	}
	if upload.File == nil {
		c.JSON(http.StatusBadRequest, UploadResponse{Success: false, Message: "No file provided"})
		return
	}

	ext := filepath.Ext(upload.File.Filename)
	filename := upload.File.Filename
	if upload.Name != "" {
		filename = upload.Name + ext
	}
	dst := filepath.Join(uploadDir, strings.ReplaceAll(filename, " ", "_"))

	if err := c.SaveUploadedFile(upload.File, dst); err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{Success: false, Message: "Failed to save file: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, UploadResponse{
		Success: true,
		Message: "File uploaded successfully",
		Files: []FileInfo{{
			Filename:    upload.File.Filename,
			Size:        upload.File.Size,
			ContentType: upload.File.Header.Get("Content-Type"),
			StoredAt:    dst,
			Description: upload.Description,
		}},
	})
}

func uploadMultipleFiles(c *router.Context) {
	var upload MultiUpload
	if err := c.BindForm(&upload); err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{Success: false, Message: "Invalid form data: " + err.Error()})
		return
	}
	if len(upload.Files) == 0 {
		c.JSON(http.StatusBadRequest, UploadResponse{Success: false, Message: "No files provided"})
		return
	}

	var fileInfos []FileInfo
	for _, file := range upload.Files {
		filename := fmt.Sprintf("%s_%s", strings.ReplaceAll(upload.Category, " ", "_"), strings.ReplaceAll(file.Filename, " ", "_"))
		dst := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, UploadResponse{Success: false, Message: "Failed to save file: " + err.Error(), Files: fileInfos})
			return
		}

		fileInfos = append(fileInfos, FileInfo{
			Filename:    file.Filename,
			Size:        file.Size,
			ContentType: file.Header.Get("Content-Type"),
			StoredAt:    dst,
		})
	}

	c.JSON(http.StatusCreated, UploadResponse{Success: true, Message: fmt.Sprintf("Uploaded %d files", len(fileInfos)), Files: fileInfos})
}

func listFiles(c *router.Context) {
	files, err := os.ReadDir(uploadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read uploads directory"})
		return
	}

	var fileList []string
	for _, file := range files {
		if !file.IsDir() {
			fileList = append(fileList, file.Name())
		}
	}
	c.JSON(http.StatusOK, map[string]interface{}{"files": fileList, "count": len(fileList)})
}

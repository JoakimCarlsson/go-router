package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joakimcarlsson/go-router/router"
)

//go:embed static/*
var embeddedFiles embed.FS

func main() {
	r := router.New()

	r.GET("/favicon.ico", func(c *router.Context) {
		c.File("./static/favicon.ico")
	})

	r.GET("/robots.txt", func(c *router.Context) {
		c.File("./static/robots.txt")
	})

	fileServer := http.FileServer(http.Dir("./static"))
	r.GET("/static/{filepath...}", func(c *router.Context) {
		fp := c.Param("filepath")
		c.Request.URL.Path = "/" + fp
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	embeddedFS, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	embeddedServer := http.FileServer(http.FS(embeddedFS))
	r.GET("/embedded/{filepath...}", func(c *router.Context) {
		fp := c.Param("filepath")
		c.Request.URL.Path = "/" + fp
		embeddedServer.ServeHTTP(c.Writer, c.Request)
	})

	r.GET("/spa/{filepath...}", spaHandler)

	r.GET("/api/data", func(c *router.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"message": "API endpoint",
			"items":   []string{"item1", "item2", "item3"},
		})
	})

	r.GET("/", func(c *router.Context) {
		c.File("./static/index.html")
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}

func spaHandler(c *router.Context) {
	fp := c.Param("filepath")
	staticDir := "./static"
	fullPath := filepath.Join(staticDir, fp)

	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(staticDir)) {
		c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		c.File(filepath.Join(staticDir, "index.html"))
		return
	}

	c.File(fullPath)
}

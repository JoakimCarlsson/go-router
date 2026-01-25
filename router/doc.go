/*
Package router provides a high-performance HTTP router for Go with support for
path parameters, middleware, route groups, and Server-Sent Events.

# Quick Start

	r := router.New()

	r.GET("/", func(c *router.Context) {
		c.JSON(200, map[string]string{"message": "Hello, World!"})
	})

	r.GET("/users/{id}", func(c *router.Context) {
		id := c.Param("id")
		c.JSON(200, map[string]string{"user_id": id})
	})

	http.ListenAndServe(":8080", r)

# Features

  - RESTful HTTP routing with path parameters (/users/{id})
  - Standard middleware compatibility (func(http.Handler) http.Handler)
  - Route groups for organization and shared middleware
  - JSON, XML, and plain text responses
  - Request body binding (JSON, XML, form data)
  - File upload support with multipart forms
  - Server-Sent Events (SSE) for real-time communication
  - High performance with object pooling

# Middleware

The router uses standard HTTP middleware, making it compatible with the Go ecosystem:

	r := router.New()

	// Use any standard HTTP middleware
	r.Use(cors.Default())

	// Custom middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, req)
			log.Printf("%s %s %v", req.Method, req.URL.Path, time.Since(start))
		})
	})

# Route Groups

Organize routes with prefixes and shared middleware:

	r.Group("/api/v1", func(api *router.Router) {
		api.Use(authMiddleware)

		api.GET("/users", listUsers)
		api.POST("/users", createUser)
		api.GET("/users/{id}", getUser)
	})

# Context Methods

The Context provides methods for handling requests and responses:

Request:
  - Param(key) - Get path parameter
  - Query(key) - Get query parameter
  - QueryDefault(key, default) - Get query parameter with default
  - BindJSON(obj) - Parse JSON request body
  - BindXML(obj) - Parse XML request body
  - BindForm(obj) - Parse form data

Response:
  - JSON(code, obj) - Send JSON response
  - XML(code, obj) - Send XML response
  - String(code, text) - Send plain text response
  - Status(code) - Set status code only
  - File(path) - Serve a file
  - Redirect(code, url) - HTTP redirect

Headers:
  - SetHeader(key, value) - Set response header
  - GetHeader(key) - Get request header

# Server-Sent Events

Built-in SSE support for real-time communication:

	r.GET("/events", func(c *router.Context) {
		c.InitSSE()

		for i := 0; i < 10; i++ {
			c.SSEJson("message", map[string]int{"count": i}, "")
			time.Sleep(time.Second)
		}
	})

# Handler Conversion

Convert between router handlers and standard HTTP handlers:

	// Standard http.Handler to router.HandlerFunc
	fileServer := http.FileServer(http.Dir("./static"))
	r.GET("/static/*", router.FromHTTPHandler(fileServer))

	// router.HandlerFunc to standard http.HandlerFunc
	handler := router.ToHTTPHandlerFunc(myRouterHandler)
	http.Handle("/api", handler)

# File Uploads

Handle multipart form uploads:

	type Upload struct {
		File *multipart.FileHeader `form:"file" file:"true"`
		Name string                `form:"name"`
	}

	r.POST("/upload", func(c *router.Context) {
		var upload Upload
		if err := c.BindForm(&upload); err != nil {
			c.JSON(400, map[string]string{"error": err.Error()})
			return
		}
		c.SaveUploadedFile(upload.File, "./uploads/"+upload.File.Filename)
		c.JSON(201, map[string]string{"status": "uploaded"})
	})

	// Configure max upload size (default 32MB)
	r.WithMultipartConfig(64 << 20) // 64MB

For OpenAPI documentation support, see the openapi package.
For Swagger UI integration, see the swaggerui package.

For more information: https://github.com/JoakimCarlsson/go-router
*/
package router

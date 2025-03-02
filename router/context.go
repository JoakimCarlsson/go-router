package router

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Context represents the context of an HTTP request, including the request and response writer.
// It provides methods for accessing request data, setting response data, and managing context values.
// Context objects are pooled to reduce allocation overhead.
type Context struct {
	// Writer is the http.ResponseWriter for the current request
	Writer http.ResponseWriter
	// Request is the *http.Request instance for the current request
	Request *http.Request
	ctx     context.Context
	// StartTime records when the context was created for tracking request duration
	StartTime time.Time
	// StatusCode holds the HTTP status code that will be or has been sent
	StatusCode int
	// params contains URL parameters extracted from the request path
	params map[string]string
	// store provides a per-request key/value store
	store map[string]interface{}
	mu    sync.RWMutex
}

// Context pool to minimize allocations
var contextPool = sync.Pool{
	New: func() interface{} {
		return &Context{
			params: make(map[string]string),
			store:  make(map[string]interface{}),
		}
	},
}

// acquireContext retrieves a Context from the pool and initializes it with the given response writer and request.
// This is called by the router for each incoming request.
func acquireContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx := contextPool.Get().(*Context)
	ctx.Writer = w
	ctx.Request = r
	ctx.ctx = r.Context()
	ctx.StartTime = time.Now()
	ctx.StatusCode = http.StatusOK
	return ctx
}

// releaseContext returns a Context to the pool and clears its data.
// This is called after a request has been processed to allow the context to be reused.
func releaseContext(ctx *Context) {
	ctx.Writer = nil
	ctx.Request = nil
	clearStringMap(ctx.params)
	clearInterfaceMap(ctx.store)
	contextPool.Put(ctx)
}

// Query returns the query parameters of the request.
// Returns the same structure as http.Request.URL.Query().
func (c *Context) Query() url.Values {
	return c.Request.URL.Query()
}

// QueryDefault returns the value of the query parameter with the given key,
// or the default value if the parameter is not present.
func (c *Context) QueryDefault(key, defaultValue string) string {
	if values, exists := c.Query()[key]; exists && len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

// QueryInt returns the integer value of the query parameter with the given key.
// Returns an error if the parameter is not present or cannot be converted to an integer.
func (c *Context) QueryInt(key string) (int, error) {
	return strconv.Atoi(c.Query().Get(key))
}

// QueryIntDefault returns the integer value of the query parameter with the given key,
// or the default value if the parameter is not present or cannot be converted to an integer.
func (c *Context) QueryIntDefault(key string, defaultValue int) int {
	if value, err := strconv.Atoi(c.Query().Get(key)); err == nil {
		return value
	}
	return defaultValue
}

// QueryBool returns the boolean value of the query parameter with the given key.
// Returns an error if the parameter is not present or cannot be converted to a boolean.
func (c *Context) QueryBool(key string) (bool, error) {
	return strconv.ParseBool(c.Query().Get(key))
}

// QueryBoolDefault returns the boolean value of the query parameter with the given key,
// or the default value if the parameter is not present or cannot be converted to a boolean.
func (c *Context) QueryBoolDefault(key string, defaultValue bool) bool {
	if value, err := strconv.ParseBool(c.Query().Get(key)); err == nil {
		return value
	}
	return defaultValue
}

// ParamInt returns the integer value of the path parameter with the given key.
// Returns an error if the parameter is not present or cannot be converted to an integer.
func (c *Context) ParamInt(key string) (int, error) {
	return strconv.Atoi(c.Param(key))
}

// ParamIntDefault returns the integer value of the path parameter with the given key,
// or the default value if the parameter is not present or cannot be converted to an integer.
func (c *Context) ParamIntDefault(key string, defaultValue int) int {
	if value, err := strconv.Atoi(c.Param(key)); err == nil {
		return value
	}
	return defaultValue
}

// ParamBool returns the boolean value of the path parameter with the given key.
// Returns an error if the parameter is not present or cannot be converted to a boolean.
func (c *Context) ParamBool(key string) (bool, error) {
	return strconv.ParseBool(c.Param(key))
}

// ParamBoolDefault returns the boolean value of the path parameter with the given key,
// or the default value if the parameter is not present or cannot be converted to a boolean.
func (c *Context) ParamBoolDefault(key string, defaultValue bool) bool {
	if value, err := strconv.ParseBool(c.Param(key)); err == nil {
		return value
	}
	return defaultValue
}

// Param returns the value of the path parameter with the given key.
// Uses Go 1.22's PathValue for path parameter extraction.
func (c *Context) Param(key string) string {
	if c.Request != nil {
		return c.Request.PathValue(key)
	}
	return ""
}

// JSON writes the given object as a JSON response with the given status code.
// It sets the Content-Type header to "application/json; charset=utf-8".
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json; charset=utf-8")
	c.Status(code)
	if err := json.NewEncoder(c.Writer).Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// XML sends an XML response with the given status code and object.
// It sets the Content-Type header to "application/xml; charset=utf-8".
func (c *Context) XML(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/xml; charset=utf-8")
	c.Status(code)
	if err := xml.NewEncoder(c.Writer).Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// Data sends a raw data response with the specified content type.
func (c *Context) Data(code int, contentType string, data []byte) {
	c.SetHeader("Content-Type", contentType)
	c.Status(code)
	c.Writer.Write(data)
}

// File serves a file response using http.ServeFile.
func (c *Context) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}

// Redirect performs an HTTP redirect to the specified location.
func (c *Context) Redirect(code int, location string) {
	http.Redirect(c.Writer, c.Request, location, code)
}

// Error sends an error response with the given status code and message.
func (c *Context) Error(code int, message string) {
	http.Error(c.Writer, message, code)
}

// Status sets the HTTP status code for the response.
// This method writes the status code to the response writer.
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// GetHeader returns the value of the request header with the given key.
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets the value of the response header with the given key.
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// BindJSON binds the request body to the given target object.
// Returns an error if the binding fails.
func (c *Context) BindJSON(target interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(target)
}

// BindXML binds XML request body to a struct.
// Returns an error if the binding fails.
func (c *Context) BindXML(obj interface{}) error {
	return xml.NewDecoder(c.Request.Body).Decode(obj)
}

// Set stores a key-value pair in the context.
// This can be used to pass data between middleware and handlers.
func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	c.store[key] = value
	c.mu.Unlock()
}

// Get retrieves a value from the context by key.
// Returns the value and a boolean indicating whether the key was found.
func (c *Context) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	value, exists := c.store[key]
	c.mu.RUnlock()
	return value, exists
}

// GetString retrieves a string value from the context.
// Returns the value and a boolean indicating whether the key was found
// and the value was of type string.
func (c *Context) GetString(key interface{}) (string, bool) {
	if val := c.ctx.Value(key); val != nil {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetInt retrieves an int value from the context.
// Returns the value and a boolean indicating whether the key was found
// and the value was of type int.
func (c *Context) GetInt(key interface{}) (int, bool) {
	if val := c.ctx.Value(key); val != nil {
		if i, ok := val.(int); ok {
			return i, true
		}
	}
	return 0, false
}

// Context returns the underlying context.Context.
func (c *Context) Context() context.Context {
	return c.ctx
}

// clearStringMap clears a string map by removing all entries.
// Used internally for context pooling.
func clearStringMap(m map[string]string) {
	for k := range m {
		delete(m, k)
	}
}

// clearInterfaceMap clears an interface map by removing all entries.
// Used internally for context pooling.
func clearInterfaceMap(m map[string]interface{}) {
	for k := range m {
		delete(m, k)
	}
}

// Negotiate performs content negotiation and returns the most appropriate content type
// based on the Accept header and the offered content types.
// If no matching content type is found, it returns the first offered type or "application/json" by default.
func (c *Context) Negotiate(offered ...string) string {
	accept := c.GetHeader("Accept")
	if accept == "" {
		if len(offered) > 0 {
			return offered[0]
		}
		return "application/json"
	}

	accepts := strings.Split(accept, ",")
	for _, accepted := range accepts {
		mediaType := strings.Split(strings.TrimSpace(accepted), ";")[0]
		for _, offer := range offered {
			if mediaType == offer || mediaType == "*/*" {
				return offer
			}
		}
	}

	return offered[0]
}

// Respond sends a response with content negotiation.
// It chooses between JSON and XML based on the Accept header.
func (c *Context) Respond(code int, obj interface{}) {
	switch c.Negotiate("application/json", "application/xml") {
	case "application/xml":
		c.XML(code, obj)
	default:
		c.JSON(code, obj)
	}
}

// GetDuration returns a duration from context.
// Returns the value and a boolean indicating whether the key was found
// and the value was of type time.Duration.
func (c *Context) GetDuration(key interface{}) (time.Duration, bool) {
	if val := c.ctx.Value(key); val != nil {
		if d, ok := val.(time.Duration); ok {
			return d, true
		}
	}
	return 0, false
}

// Deadline returns the context deadline and ok flag.
// Implements context.Context interface.
func (c *Context) Deadline() (time.Time, bool) {
	return c.ctx.Deadline()
}

// Done returns the context's Done channel.
// Implements context.Context interface.
func (c *Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

// Err returns the context's error.
// Implements context.Context interface.
func (c *Context) Err() error {
	return c.ctx.Err()
}

// Value returns the context's value for key.
// Implements context.Context interface.
func (c *Context) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

// Elapsed returns the time elapsed since the context was created.
// Useful for measuring request processing time.
func (c *Context) Elapsed() time.Duration {
	return time.Since(c.StartTime)
}

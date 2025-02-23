package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// Context represents the context of an HTTP request, including the request and response writer
// and provides methods for query parameters, headers, and JSON handling

type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	ctx        context.Context
	StartTime  time.Time
	StatusCode int
}

var contextPool = sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}

// acquireContext retrieves a Context from the pool and initializes it with the given response writer and request
func acquireContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx := contextPool.Get().(*Context)
	ctx.Writer = w
	ctx.Request = r
	ctx.ctx = r.Context()
	ctx.StartTime = time.Now()
	ctx.StatusCode = http.StatusOK
	return ctx
}

// releaseContext returns a Context to the pool
func releaseContext(ctx *Context) {
	contextPool.Put(ctx)
}

// Query returns the query parameters of the request
func (c *Context) Query() url.Values {
	return c.Request.URL.Query()
}

// QueryDefault returns the value of the query parameter with the given key, or the default value if the parameter is not present
func (c *Context) QueryDefault(key, defaultValue string) string {
	if values, exists := c.Query()[key]; exists && len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

// QueryInt returns the integer value of the query parameter with the given key
func (c *Context) QueryInt(key string) (int, error) {
	return strconv.Atoi(c.Query().Get(key))
}

// QueryIntDefault returns the integer value of the query parameter with the given key, or the default value if the parameter is not present or cannot be converted to an integer
func (c *Context) QueryIntDefault(key string, defaultValue int) int {
	if value, err := strconv.Atoi(c.Query().Get(key)); err == nil {
		return value
	}
	return defaultValue
}

// QueryBool returns the boolean value of the query parameter with the given key
func (c *Context) QueryBool(key string) (bool, error) {
	return strconv.ParseBool(c.Query().Get(key))
}

// QueryBoolDefault returns the boolean value of the query parameter with the given key, or the default value if the parameter is not present or cannot be converted to a boolean
func (c *Context) QueryBoolDefault(key string, defaultValue bool) bool {
	if value, err := strconv.ParseBool(c.Query().Get(key)); err == nil {
		return value
	}
	return defaultValue
}

// ParamInt returns the integer value of the path parameter with the given key
func (c *Context) ParamInt(key string) (int, error) {
	return strconv.Atoi(c.Param(key))
}

// ParamIntDefault returns the integer value of the path parameter with the given key, or the default value if the parameter is not present or cannot be converted to an integer
func (c *Context) ParamIntDefault(key string, defaultValue int) int {
	if value, err := strconv.Atoi(c.Param(key)); err == nil {
		return value
	}
	return defaultValue
}

// ParamBool returns the boolean value of the path parameter with the given key
func (c *Context) ParamBool(key string) (bool, error) {
	return strconv.ParseBool(c.Param(key))
}

// ParamBoolDefault returns the boolean value of the path parameter with the given key, or the default value if the parameter is not present or cannot be converted to a boolean
func (c *Context) ParamBoolDefault(key string, defaultValue bool) bool {
	if value, err := strconv.ParseBool(c.Param(key)); err == nil {
		return value
	}
	return defaultValue
}

// Param returns the value of the path parameter with the given key
func (c *Context) Param(key string) string {
	return c.Request.PathValue(key)
}

// JSON writes the given object as a JSON response with the given status code
func (c *Context) JSON(code int, obj interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.StatusCode = code
	c.Writer.WriteHeader(code)

	if err := json.NewEncoder(c.Writer).Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// Status sets the status code for the response
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// GetHeader returns the value of the request header with the given key
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets the value of the response header with the given key
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// BindJSON binds the request body to the given target object
func (c *Context) BindJSON(target interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(target)
}

// Set stores a value in the context
func (c *Context) Set(key interface{}, value interface{}) {
	c.ctx = context.WithValue(c.ctx, key, value)
	c.Request = c.Request.WithContext(c.ctx)
}

// Get retrieves a value from the context
func (c *Context) Get(key interface{}) interface{} {
	return c.ctx.Value(key)
}

// GetString retrieves a string value from the context
func (c *Context) GetString(key interface{}) (string, bool) {
	if val := c.ctx.Value(key); val != nil {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetInt retrieves an int value from the context
func (c *Context) GetInt(key interface{}) (int, bool) {
	if val := c.ctx.Value(key); val != nil {
		if i, ok := val.(int); ok {
			return i, true
		}
	}
	return 0, false
}

// Context returns the underlying context.Context
func (c *Context) Context() context.Context {
	return c.ctx
}
y
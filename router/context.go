package router

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
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
	// store provides a per-request key/value store
	store map[string]interface{}
	mu    sync.RWMutex
	// maxMultipartMemory specifies the maximum memory used for parsing multipart forms
	maxMultipartMemory int64
}

// Context pool to minimize allocations
var contextPool = sync.Pool{
	New: func() interface{} {
		return &Context{
			store: make(map[string]interface{}),
		}
	},
}

// EncoderContainer holds both a buffer and an encoder
type EncoderContainer struct {
	Buffer  *bytes.Buffer
	Encoder interface{}
}

// Encoder pools to minimize allocations
var (
	jsonEncoderPool = sync.Pool{
		New: func() interface{} {
			buf := bytes.Buffer{}
			return &EncoderContainer{
				Buffer:  &buf,
				Encoder: json.NewEncoder(&buf),
			}
		},
	}
	xmlEncoderPool = sync.Pool{
		New: func() interface{} {
			buf := bytes.Buffer{}
			return &EncoderContainer{
				Buffer:  &buf,
				Encoder: xml.NewEncoder(&buf),
			}
		},
	}
)

// acquireContext retrieves a Context from the pool and initializes it with the given response writer and request.
// This is called by the router for each incoming request.
func acquireContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx := contextPool.Get().(*Context)
	ctx.Writer = w
	ctx.Request = r
	ctx.ctx = r.Context()
	ctx.StartTime = time.Now()
	ctx.StatusCode = http.StatusOK
	ctx.maxMultipartMemory = 32 << 20 // 32 MB
	return ctx
}

// releaseContext returns a Context to the pool and clears its data.
// This is called after a request has been processed to allow the context to be reused.
func releaseContext(ctx *Context) {
	ctx.Writer = nil
	ctx.Request = nil
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
	container := jsonEncoderPool.Get().(*EncoderContainer)
	container.Buffer.Reset()
	encoder := container.Encoder.(*json.Encoder)

	if err := encoder.Encode(obj); err != nil {
		jsonEncoderPool.Put(container)
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	c.SetHeader("Content-Type", "application/json; charset=utf-8")
	c.Status(code)
	c.Writer.Write(container.Buffer.Bytes())
	jsonEncoderPool.Put(container)
}

// XML sends an XML response with the given status code and object.
// It sets the Content-Type header to "application/xml; charset=utf-8".
func (c *Context) XML(code int, obj interface{}) {
	container := xmlEncoderPool.Get().(*EncoderContainer)
	container.Buffer.Reset()
	encoder := container.Encoder.(*xml.Encoder)

	if err := encoder.Encode(obj); err != nil {
		xmlEncoderPool.Put(container)
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	c.SetHeader("Content-Type", "application/xml; charset=utf-8")
	c.Status(code)
	c.Writer.Write(container.Buffer.Bytes())
	xmlEncoderPool.Put(container)
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

// BindForm binds form data (including multipart form data) to a struct.
// It uses struct tags to map form fields to struct fields:
//   - `form:"name"` tag for form field mapping
//   - `file:"true"` tag to indicate a field should be bound to a file upload
//
// Example:
//
//	type Upload struct {
//	    File   *multipart.FileHeader `form:"file" file:"true"`
//	    Name   string                `form:"name"`
//	    Description string           `form:"description"`
//	}
//
//	var upload Upload
//	if err := c.BindForm(&upload); err != nil {
//	    // handle error
//	}
func (c *Context) BindForm(obj interface{}) error {
	if c.Request.Form == nil {
		// Try to parse multipart first, fall back to regular form
		err := c.Request.ParseMultipartForm(c.maxMultipartMemory)
		if err != nil {
			if err := c.Request.ParseForm(); err != nil {
				return err
			}
		}
	}

	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr || objValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("binding element must be a pointer to a struct")
	}

	objValue = objValue.Elem()
	objType := objValue.Type()

	for i := 0; i < objValue.NumField(); i++ {
		field := objValue.Field(i)
		fieldType := objType.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Check for form tag
		formTag := fieldType.Tag.Get("form")
		if formTag == "" {
			// Try json tag as fallback
			formTag = strings.Split(fieldType.Tag.Get("json"), ",")[0]
			if formTag == "" || formTag == "-" {
				continue
			}
		}

		// Check if this is a file field
		if fieldType.Tag.Get("file") == "true" {
			if field.Type() == reflect.TypeOf((*multipart.FileHeader)(nil)) {
				if fh, err := c.FormFile(formTag); err == nil {
					field.Set(reflect.ValueOf(fh))
				}
			} else if field.Type() == reflect.TypeOf([]*multipart.FileHeader{}) {
				if form, err := c.MultipartForm(); err == nil {
					if files := form.File[formTag]; len(files) > 0 {
						field.Set(reflect.ValueOf(files))
					}
				}
			}
			continue
		}

		// Handle regular form fields
		if values := c.Request.Form[formTag]; len(values) > 0 {
			setValue(field, values)
		}
	}

	return nil
}

// setValue sets the appropriate value to the struct field based on its type
func setValue(field reflect.Value, values []string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(values[0])
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, err := strconv.ParseInt(values[0], 10, 64); err == nil {
			field.SetInt(val)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val, err := strconv.ParseUint(values[0], 10, 64); err == nil {
			field.SetUint(val)
		}
	case reflect.Float32, reflect.Float64:
		if val, err := strconv.ParseFloat(values[0], 64); err == nil {
			field.SetFloat(val)
		}
	case reflect.Bool:
		if val, err := strconv.ParseBool(values[0]); err == nil {
			field.SetBool(val)
		}
	case reflect.Slice:
		// Handle slices of supported types
		if field.Type().Elem().Kind() == reflect.String {
			field.Set(reflect.ValueOf(values))
		}
	}
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

// FormFile returns the first file for the provided form field.
// It wraps the http.Request's FormFile function and returns the file header.
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if c.Request.MultipartForm == nil {
		if err := c.Request.ParseMultipartForm(c.maxMultipartMemory); err != nil {
			return nil, err
		}
	}
	_, fh, err := c.Request.FormFile(name)
	if err != nil {
		return nil, err
	}
	return fh, nil
}

// MultipartForm returns the parsed multipart form data.
// It calls ParseMultipartForm on the request if it hasn't been called already.
func (c *Context) MultipartForm() (*multipart.Form, error) {
	if c.Request.MultipartForm == nil {
		if err := c.Request.ParseMultipartForm(c.maxMultipartMemory); err != nil {
			return nil, err
		}
	}
	return c.Request.MultipartForm, nil
}

// FormValue returns the first value for the named component of the form data.
// It tries the URL query parameters first, then the POST or PUT form data.
func (c *Context) FormValue(name string) string {
	return c.Request.FormValue(name)
}

// SaveUploadedFile saves the uploaded file with given file header to specified destination path.
// It creates the destination file and copies the content from the uploaded file.
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

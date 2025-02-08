package router

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	StartTime  time.Time
	StatusCode int
}

var contextPool = sync.Pool{
	New: func() interface{} { return new(Context) },
}

func acquireContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx := contextPool.Get().(*Context)
	ctx.Writer = w
	ctx.Request = r
	ctx.StartTime = time.Now()
	ctx.StatusCode = http.StatusOK
	return ctx
}

func releaseContext(ctx *Context) {
	contextPool.Put(ctx)
}

func (c *Context) Query() url.Values {
	return c.Request.URL.Query()
}

func (c *Context) QueryDefault(key, defaultValue string) string {
	if values, exists := c.Query()[key]; exists && len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

func (c *Context) QueryInt(key string) (int, error) {
	return strconv.Atoi(c.Query().Get(key))
}

func (c *Context) Param(key string) string {
	return c.Request.PathValue(key)
}

func (c *Context) JSON(code int, obj interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.StatusCode = code
	c.Writer.WriteHeader(code)

	if err := json.NewEncoder(c.Writer).Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

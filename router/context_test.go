package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContext_Query(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?name=test&age=25", nil)
	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	queries := ctx.Query()
	if queries.Get("name") != "test" {
		t.Errorf("expected name=test, got %s", queries.Get("name"))
	}
	if queries.Get("age") != "25" {
		t.Errorf("expected age=25, got %s", queries.Get("age"))
	}
}

func TestContext_QueryDefault(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?name=test", nil)
	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	tests := []struct {
		key          string
		defaultValue string
		expected     string
	}{
		{"name", "default", "test"},
		{"notexist", "default", "default"},
	}

	for _, tt := range tests {
		result := ctx.QueryDefault(tt.key, tt.defaultValue)
		if result != tt.expected {
			t.Errorf("QueryDefault(%s, %s) = %s; want %s",
				tt.key, tt.defaultValue, result, tt.expected)
		}
	}
}

func TestContext_QueryInt(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?age=25&invalid=abc", nil)
	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	age, err := ctx.QueryInt("age")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if age != 25 {
		t.Errorf("expected age=25, got %d", age)
	}

	_, err = ctx.QueryInt("invalid")
	if err == nil {
		t.Error("expected error for invalid integer")
	}
}

func TestContext_JSON(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := testStruct{Name: "test", Age: 25}
	ctx.JSON(http.StatusOK, data)

	if ctx.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, ctx.StatusCode)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	var result testStruct
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Name != data.Name || result.Age != data.Age {
		t.Errorf("expected %+v, got %+v", data, result)
	}
}

func TestContext_BindJSON(t *testing.T) {
	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	input := `{"name":"test","age":25}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", strings.NewReader(input))
	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	var result testStruct
	err := ctx.BindJSON(&result)
	if err != nil {
		t.Fatalf("BindJSON failed: %v", err)
	}

	expected := testStruct{Name: "test", Age: 25}
	if result != expected {
		t.Errorf("expected %+v, got %+v", expected, result)
	}
}

func BenchmarkContext_QueryDefault(b *testing.B) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?name=test", nil)
	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.QueryDefault("name", "default")
	}
}

func BenchmarkContext_JSON(b *testing.B) {
	data := map[string]string{"test": "value"}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		ctx.JSON(http.StatusOK, data)
	}
}

func BenchmarkContext_Query(b *testing.B) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?name=test&age=25&city=stockholm&country=sweden", nil)
	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Query()
	}
}

func BenchmarkContext_QueryInt(b *testing.B) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?age=25", nil)
	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.QueryInt("age")
	}
}

func BenchmarkContext_BindJSONSmall(b *testing.B) {
	input := `{"name":"test","age":25}`
	type small struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", strings.NewReader(input))
		ctx := acquireContext(w, req)
		var result small
		ctx.BindJSON(&result)
		releaseContext(ctx)
	}
}

func BenchmarkContext_BindJSONLarge(b *testing.B) {
	// Create a larger JSON payload
	type large struct {
		ID        string                 `json:"id"`
		Name      string                 `json:"name"`
		Email     string                 `json:"email"`
		Age       int                    `json:"age"`
		Addresses []string               `json:"addresses"`
		Metadata  map[string]interface{} `json:"metadata"`
		Tags      []string               `json:"tags"`
		Active    bool                   `json:"active"`
	}

	testData := large{
		ID:        "123456789",
		Name:      "test user",
		Email:     "test@example.com",
		Age:       25,
		Addresses: []string{"addr1", "addr2", "addr3", "addr4", "addr5"},
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
			"key4": []interface{}{1, 2, 3, 4, 5},
		},
		Tags:   []string{"tag1", "tag2", "tag3", "tag4", "tag5"},
		Active: true,
	}

	input, _ := json.Marshal(testData)
	inputStr := string(input)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", strings.NewReader(inputStr))
		ctx := acquireContext(w, req)
		var result large
		ctx.BindJSON(&result)
		releaseContext(ctx)
	}
}

func BenchmarkContext_JSONSmall(b *testing.B) {
	data := map[string]string{"test": "value"}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		ctx.JSON(http.StatusOK, data)
	}
}

func BenchmarkContext_JSONLarge(b *testing.B) {
	data := map[string]interface{}{
		"id":    "123456789",
		"items": make([]map[string]interface{}, 100),
		"metadata": map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
			"nested": map[string]interface{}{
				"a": 1,
				"b": "2",
				"c": []interface{}{1, 2, 3, 4, 5},
			},
		},
	}

	// Fill items array
	for i := 0; i < 100; i++ {
		data["items"].([]map[string]interface{})[i] = map[string]interface{}{
			"index":  i,
			"value":  fmt.Sprintf("value-%d", i),
			"active": i%2 == 0,
		}
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		ctx.JSON(http.StatusOK, data)
	}
}

func BenchmarkContext_GetSetHeader(b *testing.B) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Test", "test-value")
	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.GetHeader("X-Test")
		ctx.SetHeader("X-Response", "response-value")
	}
}

func BenchmarkContext_Pool(b *testing.B) {
	b.ReportAllocs()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := acquireContext(w, req)
		releaseContext(ctx)
	}
}

func BenchmarkContext_HeaderOperations(b *testing.B) {
	b.ReportAllocs()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	headers := []string{"X-Request-ID", "X-Real-IP", "User-Agent", "Accept", "Authorization"}
	values := []string{"123", "1.2.3.4", "benchmark-client", "application/json", "Bearer token"}

	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j, h := range headers {
			ctx.SetHeader(h, values[j])
			_ = ctx.GetHeader(h)
		}
	}
}

func BenchmarkContext_QueryParamsLarge(b *testing.B) {
	b.ReportAllocs()
	// Create a URL with many query parameters
	params := make([]string, 20)
	for i := 0; i < 20; i++ {
		params[i] = fmt.Sprintf("param%d=value%d", i, i)
	}
	url := "/?" + strings.Join(params, "&")
	
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", url, nil)
	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queries := ctx.Query()
		for j := 0; j < 20; j++ {
			_ = queries.Get(fmt.Sprintf("param%d", j))
		}
	}
}

func BenchmarkContext_ConcurrentOperations(b *testing.B) {
	r := New()
	r.GET("/test", func(c *Context) {
		c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		for pb.Next() {
			ctx := acquireContext(w, req)
			ctx.Query()
			ctx.SetHeader("X-Test", "value")
			ctx.GetHeader("X-Test")
			ctx.JSON(http.StatusOK, map[string]string{"test": "value"})
			releaseContext(ctx)
		}
	})
}

func BenchmarkContext_NestedJSON(b *testing.B) {
	b.ReportAllocs()
	type deeply struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}
	
	type nested struct {
		ID        string   `json:"id"`
		Names     []string `json:"names"`
		Deep      deeply   `json:"deep"`
		DeepSlice []deeply `json:"deepSlice"`
	}

	data := nested{
		ID:    "test",
		Names: []string{"name1", "name2", "name3"},
		Deep: deeply{
			Field1: "value1",
			Field2: 42,
		},
		DeepSlice: make([]deeply, 5),
	}

	for i := 0; i < 5; i++ {
		data.DeepSlice[i] = deeply{
			Field1: fmt.Sprintf("value%d", i),
			Field2: i,
		}
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := acquireContext(w, req)
	defer releaseContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		ctx.JSON(http.StatusOK, data)
	}
}

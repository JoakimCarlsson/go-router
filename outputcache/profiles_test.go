package outputcache

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/joakimcarlsson/go-router/router"
)

func TestProfiles_AddGet(t *testing.T) {
	profiles := NewProfiles()

	profiles.Add("aggressive", time.Hour, VaryByPath())
	profiles.Add("short", 30*time.Second)

	aggressive := profiles.Get("aggressive")
	if aggressive == nil {
		t.Fatal("Expected to find aggressive profile")
	}
	if aggressive.Duration != time.Hour {
		t.Errorf("Expected duration %v, got %v", time.Hour, aggressive.Duration)
	}
	if len(aggressive.Options) != 1 {
		t.Errorf("Expected 1 option, got %d", len(aggressive.Options))
	}

	short := profiles.Get("short")
	if short == nil {
		t.Fatal("Expected to find short profile")
	}
	if short.Duration != 30*time.Second {
		t.Errorf("Expected duration %v, got %v", 30*time.Second, short.Duration)
	}

	notFound := profiles.Get("nonexistent")
	if notFound != nil {
		t.Error("Expected nil for nonexistent profile")
	}
}

func TestProfiles_Remove(t *testing.T) {
	profiles := NewProfiles()

	profiles.Add("test", time.Minute)

	if profiles.Get("test") == nil {
		t.Fatal("Expected to find test profile")
	}

	profiles.Remove("test")

	if profiles.Get("test") != nil {
		t.Error("Expected test profile to be removed")
	}
}

func TestProfiles_List(t *testing.T) {
	profiles := NewProfiles()

	profiles.Add("profile1", time.Minute)
	profiles.Add("profile2", time.Hour)
	profiles.Add("profile3", 30*time.Second)

	names := profiles.List()
	if len(names) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(names))
	}

	expected := map[string]bool{
		"profile1": true,
		"profile2": true,
		"profile3": true,
	}

	for _, name := range names {
		if !expected[name] {
			t.Errorf("Unexpected profile name: %s", name)
		}
	}
}

func TestCache_WithProfile(t *testing.T) {
	profiles := NewProfiles()
	profiles.Add("aggressive", time.Hour, VaryByPath())

	r := router.New()
	cache := New(Config{
		DefaultDuration: time.Minute,
		Profiles:        profiles,
	})

	r.Use(WithRouter(r))
	r.Use(cache.Middleware())

	callCount := 0
	r.GET("/products/{id}", func(c *router.Context) {
		callCount++
		c.JSON(200, map[string]string{"id": c.Param("id")})
	}).WithProfile("aggressive")

	req1 := httptest.NewRequest("GET", "/products/123", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	if callCount != 1 {
		t.Errorf("Expected handler to be called once, was called %d times", callCount)
	}

	req2 := httptest.NewRequest("GET", "/products/123", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if callCount != 1 {
		t.Errorf("Expected handler to still be called once (cache hit), was called %d times", callCount)
	}

	if rec2.Header().Get("X-Cache") != "HIT" {
		t.Error("Expected cache hit")
	}
}

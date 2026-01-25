package outputcache

import (
	"sync"
	"time"
)

// ProfileConfig defines a reusable cache configuration.
type ProfileConfig struct {
	Duration time.Duration
	Options  []interface{}
}

// Profiles manages named cache profiles for reusable configurations.
type Profiles struct {
	mu       sync.RWMutex
	profiles map[string]ProfileConfig
}

// NewProfiles creates a new Profiles instance.
func NewProfiles() *Profiles {
	return &Profiles{
		profiles: make(map[string]ProfileConfig),
	}
}

// Add registers a new cache profile with the given name.
func (p *Profiles) Add(name string, duration time.Duration, opts ...interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.profiles[name] = ProfileConfig{
		Duration: duration,
		Options:  opts,
	}
}

// Get retrieves a cache profile by name.
// Returns nil if the profile doesn't exist.
func (p *Profiles) Get(name string) *ProfileConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if profile, ok := p.profiles[name]; ok {
		return &profile
	}
	return nil
}

// Remove deletes a cache profile by name.
func (p *Profiles) Remove(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.profiles, name)
}

// List returns all registered profile names.
func (p *Profiles) List() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	names := make([]string, 0, len(p.profiles))
	for name := range p.profiles {
		names = append(names, name)
	}
	return names
}

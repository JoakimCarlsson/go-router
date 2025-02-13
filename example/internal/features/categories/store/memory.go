package store

import (
	"sync"

	"github.com/joakimcarlsson/go-router/example/internal/features/categories/models"
)

type InMemoryStore struct {
	categories map[string]*models.Category
	mu         sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		categories: make(map[string]*models.Category),
	}
}

func (s *InMemoryStore) Create(category *models.Category) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.categories[category.ID] = category
	return nil
}

func (s *InMemoryStore) Get(id string) (*models.Category, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	category, exists := s.categories[id]
	if !exists {
		return nil, ErrNotFound
	}
	return category, nil
}

func (s *InMemoryStore) List(limit, offset int) ([]models.Category, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.Category
	for _, category := range s.categories {
		result = append(result, *category)
	}

	if offset >= len(result) {
		return []models.Category{}, nil
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

func (s *InMemoryStore) Update(category *models.Category) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.categories[category.ID]; !exists {
		return ErrNotFound
	}

	s.categories[category.ID] = category
	return nil
}

func (s *InMemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.categories[id]; !exists {
		return ErrNotFound
	}

	delete(s.categories, id)
	return nil
}

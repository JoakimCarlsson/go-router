package store

import (
	"sync"

	"github.com/joakimcarlsson/go-router/example/internal/features/todos/models"
)

type InMemoryStore struct {
	todos map[string]*models.Todo
	mu    sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		todos: make(map[string]*models.Todo),
	}
}

func (s *InMemoryStore) Create(todo *models.Todo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.todos[todo.ID] = todo
	return nil
}

func (s *InMemoryStore) Get(id string) (*models.Todo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todo, exists := s.todos[id]
	if !exists {
		return nil, ErrNotFound
	}
	return todo, nil
}

func (s *InMemoryStore) List(limit, offset int, done *bool) ([]models.Todo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.Todo
	for _, todo := range s.todos {
		if done != nil && todo.Completed != *done {
			continue
		}
		result = append(result, *todo)
	}

	if offset >= len(result) {
		return []models.Todo{}, nil
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

func (s *InMemoryStore) Update(todo *models.Todo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.todos[todo.ID]; !exists {
		return ErrNotFound
	}

	s.todos[todo.ID] = todo
	return nil
}

func (s *InMemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.todos[id]; !exists {
		return ErrNotFound
	}

	delete(s.todos, id)
	return nil
}

package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var ErrNotFound = errors.New("not found")

// Store defines the interface for task persistence.
// In-memory for tests/local dev, PostgreSQL for production.
type Store interface {
	All() ([]*Task, error)
	Get(id string) (*Task, error)
	Create(title, description, status string) (*Task, error)
	Update(id, title, description, status string) (*Task, error)
	Delete(id string) error
	Close() error
	Ping() error
}

// MemoryStore implements Store with an in-memory map (used in tests)
type MemoryStore struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		tasks: make(map[string]*Task),
	}
}

func (s *MemoryStore) All() ([]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		result = append(result, t)
	}
	return result, nil
}

func (s *MemoryStore) Get(id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

func (s *MemoryStore) Create(title, description, status string) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	t := &Task{
		ID:          generateID(),
		Title:       title,
		Description: description,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.tasks[t.ID] = t
	return t, nil
}

func (s *MemoryStore) Update(id, title, description, status string) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, ErrNotFound
	}
	t.Title = title
	t.Description = description
	t.Status = status
	t.UpdatedAt = time.Now().UTC()
	return t, nil
}

func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(s.tasks, id)
	return nil
}

func (s *MemoryStore) Close() error { return nil }
func (s *MemoryStore) Ping() error  { return nil }

func generateID() string {
	return fmt.Sprintf("%d-%04x", time.Now().UnixNano(), rand.Intn(0xFFFF))
}

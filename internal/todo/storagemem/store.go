package storagemem

import (
	"context"
	"strings"
	"sync"
	"time"
	"todo-api/internal/todo"
)

type InMemoryStore struct {
	mu     sync.RWMutex
	items  map[int64]todo.Todo
	lastID int64
}

// Ping implements todo.Repository.
func (s *InMemoryStore) Ping(ctx context.Context) error {
	return nil
}

var _ todo.Repository = (*InMemoryStore)(nil)

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		items:  make(map[int64]todo.Todo),
		lastID: 0}
}

func (s *InMemoryStore) Create(ctx context.Context, t todo.Todo) (todo.Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastID++
	t.ID = s.lastID
	if strings.TrimSpace(t.Status) == "" {
		t.Status = todo.StatusPending
	} else {
		t.Status = strings.ToLower(t.Status)
	}
	curTime := time.Now().UTC()
	t.CreatedAt = curTime
	t.UpdatedAt = curTime
	s.items[s.lastID] = t
	return t, nil
}

func (s *InMemoryStore) Get(ctx context.Context, id int64) (t todo.Todo, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.items[id]
	if !ok {
		return todo.Todo{}, todo.ErrNotFound
	}
	return t, nil
}

func (s *InMemoryStore) Remove(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.items[id]
	if !ok {
		return todo.ErrNotFound
	}
	delete(s.items, id)
	return nil
}

func (s *InMemoryStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.items)
}

func (s *InMemoryStore) Snapshot() map[int64]todo.Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cp := make(map[int64]todo.Todo, len(s.items))
	for k, v := range s.items {
		cp[k] = v
	}
	return cp
}

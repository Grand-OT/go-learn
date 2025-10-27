package todo

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

var ErrNotFound = errors.New("not found")

func Ping(ctx context.Context) error {
	return nil
}

type InMemoryStore struct {
	mu     sync.RWMutex
	items  map[int64]Todo
	lastID int64
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		items:  make(map[int64]Todo),
		lastID: 0}
}

type Repository interface {
	Create(ctx context.Context, t Todo) (Todo, error)
	Get(ctx context.Context, id int64) (Todo, error)
	Remove(ctx context.Context, id int64) error
}

func (s *InMemoryStore) Create(ctx context.Context, t Todo) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastID++
	t.ID = s.lastID
	if strings.TrimSpace(t.Status) == "" {
		t.Status = StatusPending
	} else {
		t.Status = strings.ToLower(t.Status)
	}
	curTime := time.Now().UTC()
	t.CreatedAt = curTime
	t.UpdatedAt = curTime
	s.items[s.lastID] = t
	return t, nil
}

func (s *InMemoryStore) Get(ctx context.Context, id int64) (t Todo, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.items[id]
	if !ok {
		return Todo{}, ErrNotFound
	}
	return t, nil
}

func (s *InMemoryStore) Remove(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.items[id]
	if !ok {
		return ErrNotFound
	}
	delete(s.items, id)
	return nil
}

func (s *InMemoryStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.items)
}

func (s *InMemoryStore) Snapshot() map[int64]Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cp := make(map[int64]Todo, len(s.items))
	for k, v := range s.items {
		cp[k] = v
	}
	return cp
}

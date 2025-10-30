package storagepg

import (
	"context"
	"database/sql"
	"strings"
	"time"
	"todo-api/internal/todo"
)

type PostgresStore struct {
	db *sql.DB
}

func New(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// Ping implements todo.Repository.
func (p *PostgresStore) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Create implements todo.Repository.
func (p *PostgresStore) Create(ctx context.Context, t todo.Todo) (todo.Todo, error) {
	if strings.TrimSpace(t.Status) == "" {
		t.Status = todo.StatusPending
	} else {
		t.Status = strings.ToLower(t.Status)
	}
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now

	err := p.db.QueryRowContext(ctx, `
	INSERT INTO todos (title, description, status, created_at, updated_at)
	VALUES($1, $2, $3, $4, $5)
	RETURNING id
	`, t.Title, t.Description, t.Status, t.CreatedAt, t.UpdatedAt).Scan(&t.ID)
	if err != nil {
		return todo.Todo{}, err
	}
	return t, nil
}

// Get implements todo.Repository.
func (p *PostgresStore) Get(ctx context.Context, id int64) (todo.Todo, error) {
	panic("unimplemented")
}

// Remove implements todo.Repository.
func (p *PostgresStore) Remove(ctx context.Context, id int64) error {
	panic("unimplemented")
}

var _ todo.Repository = (*PostgresStore)(nil)

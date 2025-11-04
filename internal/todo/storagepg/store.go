package storagepg

import (
	"context"
	"database/sql"
	"errors"
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
	if id <= 0 {
		return todo.Todo{}, todo.ErrNotFound
	}

	res := todo.Todo{}
	err := p.db.QueryRowContext(ctx, `
	SELECT id, title, description, status, created_at, updated_at 
	FROM todos 
	WHERE id = $1
	`, id).Scan(
		&res.ID,
		&res.Title,
		&res.Description,
		&res.Status,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return todo.Todo{}, todo.ErrNotFound
		}
		return todo.Todo{}, err
	}

	return res, nil
}

// Remove implements todo.Repository.
func (p *PostgresStore) Remove(ctx context.Context, id int64) error {
	if id <= 0 {
		return todo.ErrNotFound
	}

	res, err := p.db.ExecContext(ctx, `
	DELETE FROM todos 
	WHERE id = $1
	`, id)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return todo.ErrNotFound
	}
	return nil
}

var _ todo.Repository = (*PostgresStore)(nil)

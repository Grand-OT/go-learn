package storagepg

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"
	"todo-api/internal/todo"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *PostgresStore) {
	t.Helper() // to indicate4 that it is a helper function

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return db, mock, New(db)
}

func TestPing_OK(t *testing.T) {
	db, mock, store := newMock(t)
	defer db.Close()

	mock.ExpectPing()

	if err := store.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectation: %v", err)
	}
}

func TestCreate_OK(t *testing.T) {
	db, mock, store := newMock(t)
	defer db.Close()

	q := regexp.QuoteMeta(`
	INSERT INTO todos (title, description, status, created_at, updated_at)
	VALUES($1, $2, $3, $4, $5)
	RETURNING id
	`)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(42)
	desc := "Desc"
	mock.ExpectQuery(q).
		WithArgs(
			"My title",
			desc,
			todo.StatusPending,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(rows)

	in := todo.Todo{
		Title:       "My title",
		Description: &desc,
		Status:      "   ",
	}

	got, err := store.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if got.ID != 42 {
		t.Fatalf("Create() id = %d, want 42", got.ID)
	}

	if got.Status != todo.StatusPending {
		t.Fatalf("Create() status = %v, want %v", got.Status, todo.StatusPending)
	}

	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Fatalf("Create() timestamp not set")
	}

	if got.CreatedAt.Sub(got.UpdatedAt) < 0 || got.CreatedAt.Sub(got.UpdatedAt) > time.Second {
		t.Fatalf("Create() timestamps look inconsistant: created_at=%v, updated_at=%v", got.CreatedAt, got.UpdatedAt)
	}
}

func TestCreate_DBError(t *testing.T) {
	db, mock, store := newMock(t)
	defer db.Close()

	q := regexp.QuoteMeta(`
	INSERT INTO todos (title, description, status, created_at, updated_at)
	VALUES($1, $2, $3, $4, $5)
	RETURNING id
	`)

	desc := "D"
	mock.ExpectQuery(q).
		WithArgs("T", desc, "Done", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("db failed!"))

	_, err := store.Create(context.Background(), todo.Todo{
		Title:       "T",
		Description: &desc,
		Status:      "Done",
	})

	if err == nil {
		t.Fatalf("Create() expected error, got nil")
	}
}

func TestGet_OK(t *testing.T) {
	db, mock, store := newMock(t)
	defer db.Close()

	q := regexp.QuoteMeta(`
		SELECT id, title, description, status, created_at, updated_at
		FROM todos
		WHERE id = $1
	`)

	now := time.Now().UTC()

	rows := sqlmock.NewRows([]string{"id", "title", "description", "status", "created_at", "updated_at"}).
		AddRow(7, "T", "D", "pending", now, now)

	mock.ExpectQuery(q).WithArgs(7).WillReturnRows(rows)

	got, err := store.Get(context.Background(), 7)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.ID != 7 || got.Title != "T" || *got.Description != "D" || got.Status != "pending" {
		t.Fatalf("Get() unexpected todo: %v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGet_NotFound(t *testing.T) {
	db, mock, store := newMock(t)
	defer db.Close()

	q := regexp.QuoteMeta(`
		SELECT id, title, description, status, created_at, updated_at
		FROM todos
		WHERE id = $1
	`)

	mock.ExpectQuery(q).WithArgs(999).WillReturnError(sql.ErrNoRows)

	_, err := store.Get(context.Background(), 999)
	if !errors.Is(err, todo.ErrNotFound) {
		t.Fatalf("Get() err = %v, want %v", err, todo.ErrNotFound)
	}
}

func TestGet_InvalidID(t *testing.T) {
	db, _, store := newMock(t)
	defer db.Close()
	_, err := store.Get(context.Background(), 0)
	if !errors.Is(err, todo.ErrNotFound) {
		t.Fatalf("Get() with id=0 should return %v, got %v", todo.ErrNotFound, err)
	}
}

func TestRemove_OK(t *testing.T) {
	db, mock, store := newMock(t)
	defer db.Close()

	q := regexp.QuoteMeta(`
	DELETE FROM todos 
	WHERE id = $1
	`)

	mock.ExpectExec(q).
		WithArgs(10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := store.Remove(context.Background(), 10)
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("Remove() unmet expectation %v", err)
	}
}

func TestRemove_NotFound(t *testing.T) {
	db, mock, store := newMock(t)
	defer db.Close()

	q := regexp.QuoteMeta(`
	DELETE FROM todos
	WHERE id = $1
	`)

	mock.ExpectExec(q).
		WithArgs(10).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := store.Remove(context.Background(), 10)
	if !errors.Is(err, todo.ErrNotFound) {
		t.Fatalf("Remove() err want %v, got %v", todo.ErrNotFound, err)
	}
}

func TestRemove_InvalidID(t *testing.T) {
	db, _, store := newMock(t)
	defer db.Close()
	err := store.Remove(context.Background(), 0)
	if !errors.Is(err, todo.ErrNotFound) {
		t.Fatalf("Remove() with id <= 0 should return err = %v, got %v", todo.ErrNotFound, err)
	}
}

func TestRemove_DBError(t *testing.T) {
	db, mock, store := newMock(t)
	defer db.Close()

	q := regexp.QuoteMeta(`
	DELETE FROM todos 
	WHERE id = $1
	`)

	mock.ExpectExec(q).
		WithArgs(10).
		WillReturnError(errors.New("db down"))

	err := store.Remove(context.Background(), 10)
	if err == nil {
		t.Fatalf("Remove() want error, got nil")
	}
}

package storagemem

import (
	"context"
	"errors"
	"sync"
	"testing"
	"todo-api/internal/todo"
)

func TestCreation(t *testing.T) {
	store := NewInMemoryStore()
	if store == nil {
		t.Errorf("Creation InMemoryStorage: expected storage, created nil")
	}
}

func TestCreationTodo(t *testing.T) {

	const testName = "Creation new todo in storage"

	store := NewInMemoryStore()

	todo := todo.Todo{Title: "title"}

	todoOut, err := store.Create(context.Background(), todo)
	if err != nil {
		t.Errorf("%s: expected no err, obtained %s", testName, err)
	}

	if todoOut.ID != store.lastID {
		t.Errorf("%s: ID mismatch: got %d, want %d",
			testName, todoOut.ID, store.lastID)
	}

	snapshot := store.Snapshot()
	if todoOut != snapshot[todoOut.ID] {
		t.Errorf("%s: item mismatch. Stored is %v, returned is %v",
			testName, snapshot[todoOut.ID], todoOut)
	}
}

func TestCreateGet_OK(t *testing.T) {

	const testName = "Creation new todo in storage"

	store := NewInMemoryStore()

	todo := todo.Todo{Title: "title"}

	todoOut, err := store.Create(context.Background(), todo)
	if err != nil {
		t.Errorf("%s: expected no err, obtained %s", testName, err)
	}

	if todoOut.ID != store.lastID {
		t.Errorf("%s: ID mismatch: got %d, want %d",
			testName, todoOut.ID, store.lastID)
	}

	todoGet, err := store.Get(context.Background(), todoOut.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %q", err)
	}

	if todoGet.ID != todoOut.ID {
		t.Fatalf("got %d, %d; want same", todoGet.ID, todoOut.ID)
	}

	if todoOut.Title != todoGet.Title || todoOut.Description != todoGet.Description {
		t.Fatalf("got %v, %v; want same", todoOut, todoGet)
	}
}

func TestCreateGetAnother_NotFound(t *testing.T) {

	const testName = "Creation new todo in storage"

	store := NewInMemoryStore()

	todoStr := todo.Todo{Title: "title"}

	todoOut, err := store.Create(context.Background(), todoStr)
	if err != nil {
		t.Errorf("%s: expected no err, obtained %s", testName, err)
	}

	if todoOut.ID != store.lastID {
		t.Errorf("%s: ID mismatch: got %d, want %d",
			testName, todoOut.ID, store.lastID)
	}

	_, err = store.Get(context.Background(), todoOut.ID+1)
	if err == nil {
		t.Fatalf("expected error: %q", todo.ErrNotFound)
	}
}

func TestCreateDeleteAnotherGet_NotFound(t *testing.T) {

	const testName = "Creation new todo in storage"

	store := NewInMemoryStore()

	todoStr := todo.Todo{Title: "title"}

	todoOut, err := store.Create(context.Background(), todoStr)
	if err != nil {
		t.Errorf("%s: expected no err, obtained %s", testName, err)
	}

	if todoOut.ID != store.lastID {
		t.Errorf("%s: ID mismatch: got %d, want %d",
			testName, todoOut.ID, store.lastID)
	}

	err = store.Remove(context.Background(), todoOut.ID+1)
	if errors.Is(err, todo.ErrNotFound) {
		t.Fatalf("expected error: %q", todo.ErrNotFound)
	}

	todoGet, err := store.Get(context.Background(), todoOut.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %q", err)
	}

	if todoGet.ID != todoOut.ID {
		t.Fatalf("got %d, %d; want same", todoGet.ID, todoOut.ID)
	}
}

func TestCreateDeleteGet_NotFound(t *testing.T) {

	const testName = "Creation new todo in storage"

	store := NewInMemoryStore()

	todoStr := todo.Todo{Title: "title"}

	todoOut, err := store.Create(context.Background(), todoStr)
	if err != nil {
		t.Errorf("%s: expected no err, obtained %s", testName, err)
	}

	if todoOut.ID != store.lastID {
		t.Errorf("%s: ID mismatch: got %d, want %d",
			testName, todoOut.ID, store.lastID)
	}

	err = store.Remove(context.Background(), todoOut.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %q", err)
	}

	_, err = store.Get(context.Background(), todoOut.ID)
	if !errors.Is(err, todo.ErrNotFound) {
		t.Fatalf("expected error: %q", todo.ErrNotFound)
	}
}

func TestCreateDeleteGetAnother_OK(t *testing.T) {

	const testName = "Creation new todo in storage"

	store := NewInMemoryStore()

	todoStr := todo.Todo{Title: "title"}

	todoOut, err := store.Create(context.Background(), todoStr)
	if err != nil {
		t.Errorf("%s: expected no err, obtained %s", testName, err)
	}

	if todoOut.ID != store.lastID {
		t.Errorf("%s: ID mismatch: got %d, want %d",
			testName, todoOut.ID, store.lastID)
	}

	err = store.Remove(context.Background(), todoOut.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %q", err)
	}

	_, err = store.Get(context.Background(), todoOut.ID)
	if !errors.Is(err, todo.ErrNotFound) {
		t.Fatalf("expected error: %q", todo.ErrNotFound)
	}
}

func TestConcurrent(t *testing.T) {
	const testName = "Concurent writing"

	store := NewInMemoryStore()

	wg := sync.WaitGroup{}
	threadNum := 100
	for i := 0; i < threadNum; i++ {

		wg.Add(1)
		go func() {
			todoStr := todo.Todo{}
			store.Create(context.Background(), todoStr)
			wg.Done()
		}()
	}
	wg.Wait()

	if store.Len() != threadNum {
		t.Errorf("%s: expected len of storage is %d, real %d", testName, threadNum, store.Len())
	}
}

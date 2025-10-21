package todo

import (
	"context"
	"sync"
	"testing"
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

	todo := Todo{Title: "title"}

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

func TestConcurrent(t *testing.T) {
	const testName = "Concurent writing"

	store := NewInMemoryStore()

	wg := sync.WaitGroup{}
	threadNum := 100
	for i := 0; i < threadNum; i++ {

		wg.Add(1)
		go func() {
			todo := Todo{}
			store.Create(context.Background(), todo)
			wg.Done()
		}()
	}
	wg.Wait()

	if store.Len() != threadNum {
		t.Errorf("%s: expected len of storage is %d, real %d", testName, threadNum, store.Len())
	}
}

package todo_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"todo-api/internal/todo"
)

func TestCreateTodo_OK(t *testing.T) {
	store := todo.NewInMemoryStore()
	h := todo.NewHandler(store)

	body := bytes.NewBufferString(`{"title":"Buy milk","description":"2L"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", body)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status=%d, want %d; body=%s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type=%q, want application/json", ct)
	}

	resp := todo.TodoDTO{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal("invalid json: ", err)
	}
	if resp.Title != "Buy milk" {
		t.Fatalf("title=%q, want %q", resp.Title, "Buy milk")
	}
}

func TestCreateTodo_EmptyDescr_OK(t *testing.T) {
	store := todo.NewInMemoryStore()
	h := todo.NewHandler(store)

	body := bytes.NewBufferString(`{"title":"Buy milk"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", body)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status=%d, want %d; body=%s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type=%q, want application/json", ct)
	}

	resp := todo.TodoDTO{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal("invalid json: ", err)
	}
	if resp.Title != "Buy milk" {
		t.Fatalf("title=%q, want %q", resp.Title, "Buy milk")
	}
}

func TestCreate_UnknownField_400(t *testing.T) {
	store := todo.NewInMemoryStore()
	h := todo.NewHandler(store)

	body := bytes.NewBufferString(`{"title":"Buy milk","description":"2L","unknown_field":"boom"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", body)

	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Status: got %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestCreateTodo_Validation_422(t *testing.T) {
	store := todo.NewInMemoryStore()
	h := todo.NewHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos",
		strings.NewReader(`{"title":""}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d, want 422", rr.Code)
	}
}

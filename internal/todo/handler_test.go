package todo_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"todo-api/internal/pkg"
	"todo-api/internal/todo"
	"todo-api/internal/todo/storagemem"
)

func TestCreateTodo_OK(t *testing.T) {
	store := storagemem.NewInMemoryStore()
	h := todo.NewHandler(store)

	body := bytes.NewBufferString(`{"title":"Buy milk","description":"2L"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", body)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

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
	store := storagemem.NewInMemoryStore()
	h := todo.NewHandler(store)

	body := bytes.NewBufferString(`{"title":"Buy milk"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", body)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

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
	store := storagemem.NewInMemoryStore()
	h := todo.NewHandler(store)

	body := bytes.NewBufferString(`{"title":"Buy milk","description":"2L","unknown_field":"boom"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", body)

	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Status: got %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestCreateTodo_Validation_422(t *testing.T) {
	store := storagemem.NewInMemoryStore()
	h := todo.NewHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos",
		strings.NewReader(`{"title":""}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d, want 422", rr.Code)
	}
}

func TestGetTodo_NoItem_404(t *testing.T) {
	store := storagemem.NewInMemoryStore()
	h := todo.NewHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/todos/123", nil)
	params := map[string]string{"id": "123"}
	scope := &pkg.Scope{}
	scope.Params = params
	req = pkg.WithScope(req, scope)

	rr := httptest.NewRecorder()
	h.GetByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d, want 404", rr.Code)
	}
}

func TestGetTodo_SetGet_OK(t *testing.T) {

	store := storagemem.NewInMemoryStore()
	h := todo.NewHandler(store)

	body := bytes.NewBufferString(`{"title":"Buy milk","description":"2L"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", body)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

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

	idStr := strconv.FormatInt(resp.ID, 10)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/todos/"+idStr, nil)
	params := map[string]string{"id": idStr}
	scope := &pkg.Scope{}
	scope.Params = params
	req = pkg.WithScope(req, scope)

	rr = httptest.NewRecorder()
	h.GetByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", rr.Code)
	}

	resp = todo.TodoDTO{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal("invalid json: ", err)
	}
	if resp.Title != "Buy milk" {
		t.Fatalf("title=%q, want %q", resp.Title, "Buy milk")
	}
}

package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"todo-api/internal/pkg"
)

func getID(w http.ResponseWriter, r *http.Request) {
	s := pkg.ScopeFrom(r)
	id := ""

	if s != nil && s.Params != nil {
		id = s.Params["id"]
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func ok(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestRouter_GroupAndEmptyPattern_RootOfGroup(t *testing.T) {
	r := &Router{}
	r.Group("/api/v1", func(api *Router) {
		api.Group("/todos", func(tr *Router) {
			tr.Handle(http.MethodGet, "", http.HandlerFunc(ok))        // GET /api/v1/todos
			tr.Handle(http.MethodGet, "/:id", http.HandlerFunc(getID)) // GET /api/v1/todos/:id
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/todos", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/todos/123", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec2.Code)
	}

	var body map[string]string
	_ = json.Unmarshal(rec2.Body.Bytes(), &body)
	if body["id"] != "123" {
		t.Fatalf("want 123, got %s", body["id"])
	}
}

func TestRouter_405_AllowHeader(t *testing.T) {
	r := &Router{}
	r.Handle(http.MethodGet, "/api/v1/ping", http.HandlerFunc(ok))
	r.Handle(http.MethodDelete, "/api/v1/ping", http.HandlerFunc(ok))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ping", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected 405, got %d", rec.Code)
	}

	allow := rec.Header().Get("Allow")

	if !(allow == "GET, DELETE" || allow == "DELETE, GET") {
		t.Fatalf("want Allow to contain GET, DELETE: got %q", allow)
	}
}

func TestRouter_404(t *testing.T) {
	r := &Router{}
	r.Handle(http.MethodGet, "/api/v1/ping", http.HandlerFunc(ok))

	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestRouter_400_BadEncoding(t *testing.T) {
	r := &Router{}
	r.Handle(http.MethodGet, "/api/v1/x/:id", http.HandlerFunc(ok))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/x/foo%2Fbar", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

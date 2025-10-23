package todo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"todo-api/internal/pkg"
)

type TodoCreateRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
}

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var in TodoCreateRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
	if err := dec.Decode(&in); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	if err := validate(&in); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	t := Todo{
		Title:       in.Title,
		Description: in.Description,
	}

	out, err := h.repo.Create(r.Context(), t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ToDTO(out))
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	scope := pkg.ScopeFrom(r)
	if scope == nil {
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	idStr, ok := scope.Params["id"]
	if !ok {
		http.Error(w, "Id required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Incorrect id", http.StatusBadRequest)
		return
	}
	if id < 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	todo, err := h.repo.Get(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Storeage error", http.StatusInternalServerError)
		return
	}

	buf := bytes.Buffer{}
	err = json.NewEncoder(&buf).Encode(ToDTO(todo))

	if err != nil {
		http.Error(w, "Encoding error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func (h *Handler) RemoveById(w http.ResponseWriter, r *http.Request) {
	scope := pkg.ScopeFrom(r)
	if scope == nil || scope.Params == nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	idStr, ok := scope.Params["id"]
	if !ok {
		http.Error(w, "Id required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Id required", http.StatusBadRequest)
		return
	}

	if err := h.repo.Remove(r.Context(), id); errors.Is(err, ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func HelloMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to my website")
}

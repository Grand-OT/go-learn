package todo

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.create(w, r)
	default:
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HelloMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to my website")
}

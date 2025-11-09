package todo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	httpx "todo-api/internal/http"
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

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var in TodoCreateRequest

	if r.Header.Get("Content-Type") != "application/json" {
		httpx.WriteError(w,
			http.StatusUnsupportedMediaType,
			"unsupported_media_type",
			"expect application/json")
	}

	if err := dec.Decode(&in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "unable to process json")
		return
	}

	if dec.Decode(&struct{}{}) != io.EOF {
		httpx.WriteError(w,
			http.StatusBadRequest,
			"invalid_json",
			"multiple json values")
	}

	if err := validate(&in); err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid_content", "unable to handle data")
		return
	}

	t := Todo{
		Title:       in.Title,
		Description: in.Description,
	}

	out, err := h.repo.Create(r.Context(), t)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "storage_error", "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ToDTO(out))
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	scope := pkg.ScopeFrom(r)
	if scope == nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_parameters", "invalid request parameters")
		return
	}

	idStr, ok := scope.Params["id"]
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "unable to get id")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid todo id")
		return
	}
	if id <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "positive id required")
		return
	}

	todo, err := h.repo.Get(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "todo_not_found", "todo not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "storage_error", "internal server error")
		return
	}

	buf := bytes.Buffer{}
	err = json.NewEncoder(&buf).Encode(ToDTO(todo))

	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "encoding_error", "internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func (h *Handler) RemoveById(w http.ResponseWriter, r *http.Request) {
	scope := pkg.ScopeFrom(r)
	if scope == nil || scope.Params == nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_parameters", "invalid_request_parameters")
		return
	}

	idStr, ok := scope.Params["id"]
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "id_required", "unable to get id")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid todo id")
		return
	}

	if err := h.repo.Remove(r.Context(), id); errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "todo_not_found", "todo not found")
		return
	} else if err != nil {
		httpx.WriteError(w,
			http.StatusInternalServerError,
			"storage_error",
			http.StatusText(http.StatusInternalServerError))
	}

	w.WriteHeader(http.StatusOK)
}

func HelloMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to my website")
}

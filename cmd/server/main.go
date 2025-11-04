package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"todo-api/internal/config"
	"todo-api/internal/http/middleware"
	"todo-api/internal/http/router"
	"todo-api/internal/todo"
	"todo-api/internal/todo/storagemem"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func healthz(w http.ResponseWriter, t *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store") // не кэшировать

	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

type ReadyHandler struct {
	Repo todo.Repository
}

func (h ReadyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := h.Repo.Ping(ctx); err != nil {
		http.Error(w, "storage not ready", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

func main() {
	cfg := config.Load()

	fs := http.FileServer(http.Dir(cfg.StaticDir))
	var repo todo.Repository = storagemem.NewInMemoryStore()
	handler := todo.NewHandler(repo)
	readyHandler := ReadyHandler{repo}

	mux := &router.Router{}
	mux.Handle(http.MethodGet, "/ui", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "static/form.html")
	}))
	mux.Handle(http.MethodGet, "/static", middleware.Logging(http.StripPrefix("/static/", fs)))
	mux.Handle(http.MethodGet, "", middleware.Logging(http.HandlerFunc(todo.HelloMessage)))
	mux.Handle(http.MethodGet, "/healthz", http.HandlerFunc(healthz))
	mux.Handle(http.MethodGet, "/readyz", readyHandler)
	mux.Group("/api/v1", func(api *router.Router) {
		api.Use(middleware.Logging)
		api.Group("todos", func(todos *router.Router) {
			todos.Handle(http.MethodPost, "", http.HandlerFunc(handler.Create))
			todos.Handle(http.MethodGet, ":id", http.HandlerFunc(handler.GetByID))
			todos.Handle(http.MethodDelete, ":id", http.HandlerFunc(handler.RemoveById))
		})
	})

	addr := cfg.Port
	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}

	log.Println("Starting on port", addr)
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	log.Println("Server started on ", addr)

	// What is going on?
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Shutting down gracefully...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Forced shutdown: ", err)
	}
	log.Println("Server stopped")
	signal.Stop(stop)
}

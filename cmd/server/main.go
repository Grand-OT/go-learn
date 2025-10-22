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
	"todo-api/internal/middleware"
	"todo-api/internal/todo"
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

	_, _ = w.Write(buf.Bytes())
}

func main() {
	cfg := config.Load()

	fs := http.FileServer(http.Dir(cfg.StaticDir))
	handler := todo.NewHandler(todo.NewInMemoryStore())

	mux := http.NewServeMux()
	mux.Handle("/static/", middleware.Logging(http.StripPrefix("/static/", fs)))
	mux.Handle("/", middleware.Logging(http.HandlerFunc(todo.HelloMessage)))
	mux.HandleFunc("/healthz", healthz)
	mux.Handle("/api/v1/todos", middleware.Logging(handler))

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

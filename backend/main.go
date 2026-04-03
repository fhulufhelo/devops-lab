package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {
	// Structured JSON logging — critical for cloud observability
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Choose store: PostgreSQL if DATABASE_URL is set, otherwise in-memory
	var store Store
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		slog.Info("connecting to PostgreSQL...")
		pgStore, err := NewPostgresStore(dbURL)
		if err != nil {
			slog.Error("failed to connect to database", "error", err)
			os.Exit(1)
		}
		defer pgStore.Close()
		store = pgStore
		slog.Info("connected to PostgreSQL")
	} else {
		slog.Info("using in-memory store (set DATABASE_URL for PostgreSQL)")
		store = NewMemoryStore()
	}

	srv := NewServer(store)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", srv.handleHealthCheck)
	mux.HandleFunc("GET /api/ready", srv.handleReadyCheck)
	mux.HandleFunc("GET /api/tasks", srv.handleListTasks)
	mux.HandleFunc("POST /api/tasks", srv.handleCreateTask)
	mux.HandleFunc("GET /api/tasks/{id}", srv.handleGetTask)
	mux.HandleFunc("PUT /api/tasks/{id}", srv.handleUpdateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", srv.handleDeleteTask)

	handler := rateLimitMiddleware(requestIDMiddleware(loggingMiddleware(corsMiddleware(mux))))

	slog.Info("server starting", "port", 8080)
	if err := http.ListenAndServe(":8080", handler); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

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

	store := NewTaskStore()
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

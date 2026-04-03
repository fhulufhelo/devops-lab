package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewTaskStore()
	srv := NewServer(store)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", srv.handleHealthCheck)
	mux.HandleFunc("GET /api/tasks", srv.handleListTasks)
	mux.HandleFunc("POST /api/tasks", srv.handleCreateTask)
	mux.HandleFunc("GET /api/tasks/{id}", srv.handleGetTask)
	mux.HandleFunc("PUT /api/tasks/{id}", srv.handleUpdateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", srv.handleDeleteTask)

	handler := loggingMiddleware(corsMiddleware(mux))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

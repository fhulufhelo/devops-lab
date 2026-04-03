package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestServer() (*Server, *http.ServeMux) {
	store := NewMemoryStore()
	srv := NewServer(store)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", srv.handleHealthCheck)
	mux.HandleFunc("GET /api/tasks", srv.handleListTasks)
	mux.HandleFunc("POST /api/tasks", srv.handleCreateTask)
	mux.HandleFunc("GET /api/tasks/{id}", srv.handleGetTask)
	mux.HandleFunc("PUT /api/tasks/{id}", srv.handleUpdateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", srv.handleDeleteTask)
	return srv, mux
}

func TestHealthCheck(t *testing.T) {
	_, mux := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %s", resp["status"])
	}
}

func TestCreateTask(t *testing.T) {
	_, mux := setupTestServer()
	body := `{"title":"Test Task","description":"A test","status":"todo"}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var task Task
	json.NewDecoder(w.Body).Decode(&task)
	if task.Title != "Test Task" {
		t.Fatalf("expected title 'Test Task', got '%s'", task.Title)
	}
	if task.ID == "" {
		t.Fatal("expected non-empty ID")
	}
}

func TestListTasks(t *testing.T) {
	_, mux := setupTestServer()

	// Create two tasks
	for _, title := range []string{"Task 1", "Task 2"} {
		body := `{"title":"` + title + `","status":"todo"}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var tasks []Task
	json.NewDecoder(w.Body).Decode(&tasks)
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestGetTask(t *testing.T) {
	_, mux := setupTestServer()

	// Create a task
	body := `{"title":"Get Me","status":"todo"}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var created Task
	json.NewDecoder(w.Body).Decode(&created)

	// Get it by ID
	req = httptest.NewRequest(http.MethodGet, "/api/tasks/"+created.ID, nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var task Task
	json.NewDecoder(w.Body).Decode(&task)
	if task.ID != created.ID {
		t.Fatalf("expected ID %s, got %s", created.ID, task.ID)
	}
}

func TestGetTaskNotFound(t *testing.T) {
	_, mux := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUpdateTask(t *testing.T) {
	_, mux := setupTestServer()

	// Create a task
	body := `{"title":"Original","status":"todo"}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var created Task
	json.NewDecoder(w.Body).Decode(&created)

	// Update it
	updateBody := `{"title":"Updated","description":"new desc","status":"done"}`
	req = httptest.NewRequest(http.MethodPut, "/api/tasks/"+created.ID, bytes.NewBufferString(updateBody))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var updated Task
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.Title != "Updated" {
		t.Fatalf("expected title 'Updated', got '%s'", updated.Title)
	}
	if updated.Status != "done" {
		t.Fatalf("expected status 'done', got '%s'", updated.Status)
	}
}

func TestDeleteTask(t *testing.T) {
	_, mux := setupTestServer()

	// Create a task
	body := `{"title":"Delete Me","status":"todo"}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var created Task
	json.NewDecoder(w.Body).Decode(&created)

	// Delete it
	req = httptest.NewRequest(http.MethodDelete, "/api/tasks/"+created.ID, nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// Verify it's gone
	req = httptest.NewRequest(http.MethodGet, "/api/tasks/"+created.ID, nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w.Code)
	}
}

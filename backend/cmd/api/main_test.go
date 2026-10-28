package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	healthHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatal("expected a JSON response")
	}
	if strings.TrimSpace(response.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("unexpected health response: %s", response.Body.String())
	}
}

func TestInvalidTaskRequests(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "null character in new title",
			method: http.MethodPost,
			path:   "/api/tasks",
			body:   `{"title":"bad\u0000title"}`,
		},
		{
			name:   "null character in updated title",
			method: http.MethodPatch,
			path:   "/api/tasks/1",
			body:   `{"title":"bad\u0000title"}`,
		},
		{
			name:   "blank title",
			method: http.MethodPost,
			path:   "/api/tasks",
			body:   `{"title":"   "}`,
		},
		{
			name:   "title exceeds limit",
			method: http.MethodPost,
			path:   "/api/tasks",
			body:   `{"title":"` + strings.Repeat("a", 201) + `"}`,
		},
		{
			name:   "unknown field",
			method: http.MethodPost,
			path:   "/api/tasks",
			body:   `{"title":"Learn Go","unexpected":true}`,
		},
		{
			name:   "multiple JSON values",
			method: http.MethodPost,
			path:   "/api/tasks",
			body:   `{"title":"Learn Go"} {}`,
		},
		{
			name:   "empty update",
			method: http.MethodPatch,
			path:   "/api/tasks/1",
			body:   `{}`,
		},
		{
			name:   "invalid completion type",
			method: http.MethodPatch,
			path:   "/api/tasks/1",
			body:   `{"completed":"yes"}`,
		},
		{
			name:   "invalid task ID",
			method: http.MethodDelete,
			path:   "/api/tasks/invalid",
		},
	}

	// Invalid requests must be rejected before accessing the database.
	app := &application{}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/tasks", app.createTaskHandler)
	mux.HandleFunc("PATCH /api/tasks/{id}", app.updateTaskHandler)
	mux.HandleFunc("DELETE /api/tasks/{id}", app.deleteTaskHandler)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			response := httptest.NewRecorder()

			mux.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d: %s", response.Code, response.Body.String())
			}
		})
	}
}

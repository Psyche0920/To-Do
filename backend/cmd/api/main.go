package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	// The blank import registers pgx; DSN values configure a connection, not the driver.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Task is a saved task returned by the API.
type Task struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type application struct {
	db *sql.DB
}

func main() {
	// DSN supplies host, port, user, password and database name; Go does not load .env itself.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	db, err := openDatabase(dsn)
	if err != nil {
		log.Fatal(err)
	}
	// Deferred cleanup runs on normal return; log.Fatal exits immediately.
	defer db.Close()
	log.Println("Database connection established")

	app := &application{db: db}

	// One mux can map many HTTP method/path pairs to handler functions.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /wehan", wehanHandler)
	mux.HandleFunc("GET /api/tasks", app.listTasksHandler)
	mux.HandleFunc("POST /api/tasks", app.createTaskHandler)
	mux.HandleFunc("PATCH /api/tasks/{id}", app.updateTaskHandler)
	mux.HandleFunc("DELETE /api/tasks/{id}", app.deleteTaskHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
		// This limits reading HTTP request headers, not database work or the entire request.
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API server listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	}); err != nil {
		log.Printf("write health response: %v", err)
	}
}

func wehanHandler(w http.ResponseWriter, r *http.Request) {
	// A custom header is separate from the Content-Type header.
	w.Header().Set("wehan", "to do")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	}); err != nil {
		log.Printf("write wehan response: %v", err)
	}
}

func (app *application) listTasksHandler(w http.ResponseWriter, r *http.Request) {
	// Limit the query to five seconds; inherit request cancellation.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := app.db.QueryContext(
		ctx,
		"SELECT id, title, completed FROM tasks ORDER BY id",
	)

	if err != nil {
		log.Printf("query tasks: %v", err)
		http.Error(w, "Unable to retrieve tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Return [] instead of null when the table is empty.
	tasks := make([]Task, 0)

	for rows.Next() {
		var task Task
		// Read columns into fields in the same order as SELECT.
		if err := rows.Scan(&task.ID, &task.Title, &task.Completed); err != nil {
			log.Printf("scan task: %v", err)
			http.Error(w, "Unable to retrieve tasks", http.StatusInternalServerError)
			return
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		log.Printf("read tasks: %v", err)
		http.Error(w, "Unable to retrieve tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Printf("write tasks response: %v", err)
	}
}

func (app *application) createTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Only the title is accepted when creating a task.
	var input struct {
		Title string `json:"title"`
	}

	// Limit the request body to 1 MiB.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	decodeErr := decoder.Decode(&input)
	if decodeErr != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Accept exactly one JSON value.
	var extra any
	extraErr := decoder.Decode(&extra)
	if extraErr != io.EOF {
		http.Error(w, "Request body must contain one JSON object", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(input.Title)
	if strings.ContainsRune(title, '\x00') {
		http.Error(w, "Title cannot contain a null character", http.StatusBadRequest)
		return
	}
	titleLength := len([]rune(title))

	if titleLength == 0 || titleLength > 200 {
		http.Error(w, "Title must contain 1 to 200 characters", http.StatusBadRequest)
		return
	}

	// Database work follows the request cancellation and a five-second timeout.
	requestContext := r.Context()
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(requestContext, timeout)
	defer cancel()

	var task Task

	// Insert and return the new task.
	row := app.db.QueryRowContext(
		ctx,
		"INSERT INTO tasks (title) VALUES ($1) RETURNING id, title, completed",
		title,
	)

	scanErr := row.Scan(&task.ID, &task.Title, &task.Completed)
	if scanErr != nil {
		log.Printf("create task: %v", scanErr)
		http.Error(w, "Unable to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encodeErr := encoder.Encode(task)
	if encodeErr != nil {
		log.Printf("write created task response: %v", encodeErr)
	}
}

func (app *application) updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	idText := r.PathValue("id")
	id, idErr := strconv.ParseInt(idText, 10, 64)
	if idErr != nil || id <= 0 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Pointers distinguish omitted fields from explicit values.
	var input struct {
		Title     *string `json:"title"`
		Completed *bool   `json:"completed"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	decodeErr := decoder.Decode(&input)
	if decodeErr != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var extra any
	extraErr := decoder.Decode(&extra)
	if extraErr != io.EOF {
		http.Error(w, "Request body must contain one JSON object", http.StatusBadRequest)
		return
	}

	if input.Completed == nil && input.Title == nil {
		http.Error(w, "Provide title or completed", http.StatusBadRequest)
		return
	}

	if input.Title != nil {
		cleanTitle := strings.TrimSpace(*input.Title)
		if strings.ContainsRune(cleanTitle, '\x00') {
			http.Error(w, "Title cannot contain a null character", http.StatusBadRequest)
			return
		}
		titleLength := len([]rune(cleanTitle))
		if titleLength == 0 || titleLength > 200 {
			http.Error(w, "Title must contain 1 to 200 characters", http.StatusBadRequest)
			return
		}
		input.Title = &cleanTitle
	}

	// Limit the database update and returned row to five seconds.
	requestContext := r.Context()
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(requestContext, timeout)
	defer cancel()

	var task Task

	// Update and return the task.
	row := app.db.QueryRowContext(
		ctx,
		`UPDATE tasks
		 SET title = COALESCE($1, title), completed = COALESCE($2, completed)
		 WHERE id = $3
		 RETURNING id, title, completed`,
		input.Title,
		input.Completed,
		id,
	)

	scanErr := row.Scan(&task.ID, &task.Title, &task.Completed)
	taskNotFound := errors.Is(scanErr, sql.ErrNoRows)
	if taskNotFound {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if scanErr != nil {
		log.Printf("update task: %v", scanErr)
		http.Error(w, "Unable to update task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)
	encodeErr := encoder.Encode(task)
	if encodeErr != nil {
		log.Printf("write updated task response: %v", encodeErr)
	}
}

func (app *application) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	idText := r.PathValue("id")
	id, idErr := strconv.ParseInt(idText, 10, 64)
	if idErr != nil || id <= 0 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Limit the database deletion to five seconds.
	requestContext := r.Context()
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(requestContext, timeout)
	defer cancel()

	// Delete the task without returning a result row.
	result, deleteErr := app.db.ExecContext(
		ctx,
		"DELETE FROM tasks WHERE id = $1",
		id,
	)
	if deleteErr != nil {
		log.Printf("delete task: %v", deleteErr)
		http.Error(w, "Unable to delete task", http.StatusInternalServerError)
		return
	}

	// Read the deleted-row count reported by the database: 0 means not found.
	affected, affectedErr := result.RowsAffected()
	if affectedErr != nil {
		log.Printf("read deleted task count: %v", affectedErr)
		http.Error(w, "Unable to delete task", http.StatusInternalServerError)
		return
	}

	if affected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Successful deletion has no response body.
	w.WriteHeader(http.StatusNoContent)
}

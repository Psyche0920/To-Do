// Files in this directory share one package; no file-to-file import is needed.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	// The blank import registers pgx; DSN values configure a connection, not the driver.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Task holds task data; JSON tags choose output keys, not values.
type Task struct {
	Id        int64  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// application groups shared dependencies; it is not the PostgreSQL database.
type application struct {
	// sql.DB is the connection-pool type; *sql.DB is a pointer to that pool.
	db *sql.DB
}

func main() {
	// DSN supplies host, port, user, password and database name; Go does not load .env itself.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	// Receive two values: a pool pointer and an error, not the database server itself.
	db, err := openDatabase(dsn)
	if err != nil {
		log.Fatal(err)
	}
	// Schedule pool cleanup on normal return; log.Fatal exits without running defers.
	defer db.Close()
	log.Println("Database connection established")

	// & takes the struct address; the left db is its field, the right db is the local pool.
	app := &application{db: db}

	// One mux can map many HTTP method/path pairs to handler functions.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /wehan", wehanHandler)
	// Binding the method carries app (and app.db) while keeping the HTTP handler signature.
	mux.HandleFunc("GET /api/tasks", app.listTasksHandler)
	mux.HandleFunc("POST /api/tasks", app.createTaskHandler)

	// & creates a pointer to this server struct; configuring it does not start listening.
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
		// This limits reading HTTP request headers, not database work or the entire request.
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("API server listening on :8080")
	// ListenAndServe starts listening; the preceding log line only announces the attempt.
	log.Fatal(server.ListenAndServe())
}

// w writes the response to the client; r is a pointer to the incoming request.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Header sets metadata; declaring JSON does not itself encode the body.
	w.Header().Set("Content-Type", "application/json")
	// WriteHeader sends the status code (200); set header fields before this call.
	w.WriteHeader(http.StatusOK)

	// map[string]string is a string-key/string-value map; Encode writes JSON plus a newline.
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	}); err != nil {
		// %v formats the error value; logging writes to the server terminal, not the client.
		log.Printf("write health response: %v", err)
	}
}

func wehanHandler(w http.ResponseWriter, r *http.Request) {
	// A custom header is separate from the Content-Type header.
	w.Header().Set("wehan", "to do")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	}); err != nil {
		log.Printf("write health response: %v", err)
	}
}

// The receiver is like Python self: app points to the struct; app.db points to the pool.
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
		if err := rows.Scan(&task.Id, &task.Title, &task.Completed); err != nil {
			log.Printf("scan task : %v", err)
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

	scanErr := row.Scan(&task.Id, &task.Title, &task.Completed)
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

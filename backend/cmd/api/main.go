package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Task struct {
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	db, err := openDatabase(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Database connection established")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /wehan", wehanHandler)
	mux.HandleFunc("GET /api/tasks", listTasksHandler)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("API server listening on :8080")
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
	w.Header().Set("wehan", "to do")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	}); err != nil {
		log.Printf("write health response: %v", err)
	}
}

func listTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	tasks := []Task{
		{Title: "hackthon", Completed: true},
		{Title: "sleep", Completed: false},
	}

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Printf("write tasks response: %v", err)
	}
}

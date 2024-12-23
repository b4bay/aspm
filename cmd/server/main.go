package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
)

var db *sql.DB

type RequestBody struct {
	Data string `json:"data"`
}

func init() {
	var err error
	// Get datasource name from environment variable or use in-memory DB by default
	datasourceName := os.Getenv("DATASOURCE_NAME")
	if datasourceName == "" {
		datasourceName = ":memory:"
	}

	// Initialize SQLite database
	db, err = sql.Open("sqlite3", datasourceName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create tables if not exist
	createTableQuery := `CREATE TABLE IF NOT EXISTS requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		method TEXT NOT NULL,
		data TEXT NOT NULL
	)`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

func collectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body RequestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	saveToDatabase("collect", body.Data)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data collected successfully"))
}

func originHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body RequestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	saveToDatabase("origin", body.Data)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Origin data stored successfully"))
}

func gwHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("GW endpoint is functional"))
}

func saveToDatabase(method, data string) {
	insertQuery := `INSERT INTO requests (method, data) VALUES (?, ?)`
	_, err := db.Exec(insertQuery, method, data)
	if err != nil {
		log.Printf("Failed to save data to database: %v", err)
	}
}

func main() {
	http.HandleFunc("/api/v1/collect", collectHandler)
	http.HandleFunc("/api/v1/origin", originHandler)
	http.HandleFunc("/api/v1/gw", gwHandler)

	fmt.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

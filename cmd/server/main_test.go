package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	createTableQuery := `CREATE TABLE IF NOT EXISTS requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		method TEXT NOT NULL,
		data TEXT NOT NULL
	)`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		panic(err)
	}

	return db
}

func TestCollectHandler(t *testing.T) {
	db = setupTestDB()
	reqBody := RequestBody{Data: "test collect data"}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/collect", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	collectHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.StatusCode)
	}

	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM requests WHERE method = 'collect' AND data = ?`, reqBody.Data).Scan(&count)
	if err != nil || count != 1 {
		t.Errorf("expected one record in database, got %d", count)
	}
}

func TestOriginHandler(t *testing.T) {
	db = setupTestDB()
	reqBody := RequestBody{Data: "test origin data"}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/origin", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	originHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.StatusCode)
	}

	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM requests WHERE method = 'origin' AND data = ?`, reqBody.Data).Scan(&count)
	if err != nil || count != 1 {
		t.Errorf("expected one record in database, got %d", count)
	}
}

func TestGwHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/gw", nil)
	w := httptest.NewRecorder()

	gwHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.StatusCode)
	}

	body := w.Body.String()
	if body != "GW endpoint is functional" {
		t.Errorf("unexpected response body: %s", body)
	}
}

func TestInvalidMethods(t *testing.T) {
	tests := []struct {
		method string
		url    string
	}{
		{method: http.MethodGet, url: "/api/v1/collect"},
		{method: http.MethodGet, url: "/api/v1/origin"},
		{method: http.MethodPost, url: "/api/v1/gw"},
	}

	for _, test := range tests {
		req := httptest.NewRequest(test.method, test.url, nil)
		w := httptest.NewRecorder()

		switch test.url {
		case "/api/v1/collect":
			collectHandler(w, req)
		case "/api/v1/origin":
			originHandler(w, req)
		case "/api/v1/gw":
			gwHandler(w, req)
		}

		resp := w.Result()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected status Method Not Allowed for %s %s, got %v", test.method, test.url, resp.StatusCode)
		}
	}
}

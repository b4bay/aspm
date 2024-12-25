package main

import (
	"bytes"
	"encoding/json"
	"github.com/b4bay/aspm/internal/server"
	"github.com/b4bay/aspm/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to in-memory database")
	}
	db.AutoMigrate(&shared.Product{}, &shared.Link{})
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
}

func TestOriginHandler(t *testing.T) {
	db = setupTestDB()

	tests := []struct {
		name           string
		requestBody    shared.OriginMessageBody
		expectedStatus int
	}{
		{
			name: "Valid request",
			requestBody: shared.OriginMessageBody{
				Environment: map[string]string{
					"CI_PROJECT_PATH":       "b4bay/read-it-later-be",
					"GITLAB_USER_NAME":      "Alex Goncharov",
					"CI_RUNNER_DESCRIPTION": "4-blue.saas-linux-small-amd64.runners-manager.gitlab.com/default",
				},
				ProductName: "product.name",
				ProductType: shared.ArtefactTypeGit,
				ProductId:   "product-123",
				OriginIds:   []string{"origin-1", "origin-2"},
				ProdMethod:  shared.ProductionMethodCompile,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request body
			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/origin", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			respRecorder := httptest.NewRecorder()

			// Call handler
			originHandler(respRecorder, req)

			// Validate response
			if respRecorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, respRecorder.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				// Validate database entries
				var product shared.Product
				if err := db.First(&product, "id = ?", tt.requestBody.ProductId).Error; err != nil {
					t.Errorf("Product not found in database: %v", err)
				}

				if product.Project != server.GetProjectFromEnvironment(tt.requestBody.Environment) ||
					product.Author != server.GetAuthorFromEnvironment(tt.requestBody.Environment) ||
					product.Worker != server.GetWorkerFromEnvironment(tt.requestBody.Environment) {
					t.Errorf("Product fields not properly populated: %+v", product)
				}

				for _, originID := range tt.requestBody.OriginIds {
					var origin shared.Product
					if err := db.First(&origin, "id = ?", originID).Error; err != nil {
						t.Errorf("Origin not found in database: %v", err)
					}
				}

				var links []shared.Link
				if err := db.Where("product_id = ?", tt.requestBody.ProductId).Find(&links).Error; err != nil {
					t.Errorf("Failed to find links in database: %v", err)
				}

				if len(links) != len(tt.requestBody.OriginIds) {
					t.Errorf("Expected %d links, got %d", len(tt.requestBody.OriginIds), len(links))
				}
			}
		})
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

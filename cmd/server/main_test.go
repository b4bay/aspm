package main

import (
	"bytes"
	"encoding/json"
	"github.com/b4bay/aspm/internal/server"
	"github.com/b4bay/aspm/internal/server/sarif"
	"github.com/b4bay/aspm/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	db.AutoMigrate(&server.Product{}, &server.Link{}, &server.Engagement{}, &server.Vulnerability{}, &server.Status{})
	return db
}

func TestCollectHandler(t *testing.T) {
	db = setupTestDB()

	t.Run("valid request", func(t *testing.T) {
		// Create a valid CollectMessageBody
		body := shared.CollectMessageBody{
			ArtefactId: "test-artifact",
			Reports:    map[string]string{"gosec.sarif": sarif.MockGosecReport, "govuncheck.sarif": sarif.MockGovulncheckReport},
		}

		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		// Create a new HTTP request
		req := httptest.NewRequest(http.MethodPost, "/api/v1/collect", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Create a ResponseRecorder to capture the response
		rec := httptest.NewRecorder()

		// Call the handler
		server.CollectHandler(rec, req)

		// Validate the response
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if rec.Body.String() != "Data collected successfully" {
			t.Errorf("Expected response body \"Data collected successfully\", got %s", rec.Body.String())
		}

		// Validate the database entries
		var product server.Product
		if err := db.First(&product, "product_id = ?", body.ArtefactId).Error; err != nil {
			t.Errorf("Failed to find product: %v", err)
		}

		var engagements []server.Engagement
		if err := db.Where("product_id = ?", product.ProductID).Preload(clause.Associations).Find(&engagements).Error; err != nil {
			t.Errorf("Failed to find engagements: %v", err)
		}
		if len(engagements) != len(body.Reports) {
			t.Errorf("Expected %d engagements, got %d", len(body.Reports), len(engagements))
		}

		for _, e := range engagements {
			var vulnerabilities []server.Vulnerability
			if err := db.Where("engagement_id = ?", e.ID).Preload(clause.Associations).Find(&vulnerabilities).Error; err != nil {
				t.Errorf("Failed to find vulberabilities for %s: %v", e.Tool, err)
			}
			if len(vulnerabilities) != len(e.Report().Runs[0].Results) {
				t.Errorf("Expected %d vulnerabilities for %s, got %d", len(e.Report().Runs[0].Results), e.Tool, len(vulnerabilities))
			}
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		// Create an invalid JSON body
		invalidJSON := "{invalid}"

		// Create a new HTTP request
		req := httptest.NewRequest(http.MethodPost, "/api/v1/collect", bytes.NewReader([]byte(invalidJSON)))
		req.Header.Set("Content-Type", "application/json")

		// Create a ResponseRecorder to capture the response
		rec := httptest.NewRecorder()

		// Call the handler
		server.CollectHandler(rec, req)

		// Validate the response
		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
		if rec.Body.String() != "Invalid JSON\n" {
			t.Errorf("Expected response body \"Invalid JSON\\n\", got %s", rec.Body.String())
		}
	})
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
				Product: shared.ProductMessage{
					Name: "product.name",
					Type: shared.ArtefactTypeBin,
					Id:   "product-123",
				},
				Origins: []shared.ProductMessage{
					{Id: "origin-1", Name: "main", Type: shared.ArtefactTypeGit},
					{Id: "origin-2", Name: "library.so", Type: shared.ArtefactTypeBin},
				},
				ProductionMethod: shared.ProductionMethodCompile,
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
			server.OriginHandler(respRecorder, req)

			// Validate response
			if respRecorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, respRecorder.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				// Validate database entries
				var product server.Product
				if err := db.First(&product, "product_id = ?", tt.requestBody.Product.Id).Error; err != nil {
					t.Errorf("Product not found in database: %v", err)
				}

				if product.Project != server.GetProjectFromEnvironment(tt.requestBody.Environment) ||
					product.Author != server.GetAuthorFromEnvironment(tt.requestBody.Environment) ||
					product.Worker != server.GetWorkerFromEnvironment(tt.requestBody.Environment) {
					t.Errorf("Product fields not properly populated: %+v", product)
				}

				for _, o := range tt.requestBody.Origins {
					var origin server.Product
					if err := db.First(&origin, "product_id = ?", o.Id).Error; err != nil {
						t.Errorf("Origin not found in database: %v", err)
					}
				}

				var links []server.Link
				if err := db.Preload(clause.Associations).Where("product_id = ?", product.ProductID).Find(&links).Error; err != nil {
					t.Errorf("Failed to find links in database: %v", err)
				}

				if len(links) != len(tt.requestBody.Origins) {
					t.Errorf("Expected %d links, got %d", len(tt.requestBody.Origins), len(links))
				}
			}
		})
	}
}

func TestGwHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/gw", nil)
	w := httptest.NewRecorder()

	server.GWHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.StatusCode)
	}

	body := w.Body.String()
	if body != "GW endpoint is functional" {
		t.Errorf("unexpected response body: %s", body)
	}
}

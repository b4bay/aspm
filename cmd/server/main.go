package main

import (
	"encoding/json"
	"fmt"
	"github.com/b4bay/aspm/internal/server"
	"github.com/b4bay/aspm/internal/shared"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"time"
)

var db *gorm.DB

type RequestBody struct {
	Data string `json:"data"`
}

func init() {
	var err error

	// Get datasource name from environment variable or use in-memory DB by default
	datasourceName := os.Getenv("DATASOURCE_NAME")
	if datasourceName == "" {
		datasourceName = "file::memory:?cache=shared"
	}

	db, err = gorm.Open(sqlite.Open(datasourceName), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&shared.Product{}, &shared.Link{})

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

	// TODO: Save data

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data collected successfully"))
}

func originHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body shared.OriginMessageBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if body.ProdMethod == "" {
		body.ProdMethod = shared.ProductionMethodDefault
	}

	var project = server.GetProjectFromEnvironment(body.Environment)
	var author = server.GetAuthorFromEnvironment(body.Environment)
	var worker = server.GetWorkerFromEnvironment(body.Environment)

	db.Transaction(func(tx *gorm.DB) error {
		// Ensure Product exists
		var product shared.Product
		if err := tx.FirstOrCreate(&product, shared.Product{
			ID:        body.ProductId,
			CreatedAt: time.Now(),
			Name:      body.ProductName,
			Type:      body.ProductType,
			Project:   project,
			Author:    author,
			Worker:    worker,
		}).Error; err != nil {
			http.Error(w, "Failed to create or find product", http.StatusInternalServerError)
			return err
		}

		var needToUpdate = false
		// Check and update empty fields in Product
		if product.Name == "" && body.ProductName != "" {
			product.Name = body.ProductName
			needToUpdate = true
		}
		if product.Type == "" && body.ProductType != "" {
			product.Type = body.ProductType
			needToUpdate = true
		}
		if product.Project == "" && project != "" {
			product.Project = project
			needToUpdate = true
		}
		if product.Author == "" && author != "" {
			product.Author = author
			needToUpdate = true
		}
		if product.Worker == "" && worker != "" {
			product.Worker = worker
			needToUpdate = true
		}

		// Save updated product if necessary
		if needToUpdate {
			if err := tx.Save(&product).Error; err != nil {
				http.Error(w, "Failed to update product", http.StatusInternalServerError)
				return err
			}
		}

		// Process OriginIds and create Links
		for _, originID := range body.OriginIds {
			var origin shared.Product
			if err := tx.FirstOrCreate(&origin, shared.Product{
				ID:        originID,
				CreatedAt: time.Now(),
			}).Error; err != nil {
				http.Error(w, "Failed to create or find origin", http.StatusInternalServerError)
				return err
			}

			link := shared.Link{
				ProductID: product.ID,
				OriginID:  origin.ID,
				Type:      body.ProdMethod,
				CreatedAt: time.Now(),
			}
			if err := tx.Create(&link).Error; err != nil {
				http.Error(w, "Failed to create link", http.StatusInternalServerError)
				return err
			}
		}

		return nil
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Links created successfully"))
}

func gwHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("GW endpoint is functional"))
}

func main() {
	http.HandleFunc("/api/v1/collect", collectHandler)
	http.HandleFunc("/api/v1/origin", originHandler)
	http.HandleFunc("/api/v1/gw", gwHandler)

	fmt.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

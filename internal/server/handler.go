package server

import (
	"encoding/json"
	"github.com/b4bay/aspm/internal/shared"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"time"
)

type RequestBody struct {
	Data string `json:"data"`
}

func CollectHandler(w http.ResponseWriter, r *http.Request) {
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

func OriginHandler(w http.ResponseWriter, r *http.Request) {
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

	var project = GetProjectFromEnvironment(body.Environment)
	var author = GetAuthorFromEnvironment(body.Environment)
	var worker = GetWorkerFromEnvironment(body.Environment)

	DB.Transaction(func(tx *gorm.DB) error {
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

func GWHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("GW endpoint is functional"))
}

func UIProductHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch all products from the database
	var products []shared.Product
	if err := DB.Find(&products).Error; err != nil {
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}

	// Set the response header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Marshal the products into JSON and write to the response
	if err := json.NewEncoder(w).Encode(products); err != nil {
		http.Error(w, "Failed to encode products to JSON", http.StatusInternalServerError)
		return
	}
}

func UIVersionHandler(w http.ResponseWriter, r *http.Request) {
	// Read the VERSION file in the current directory
	versionFile := "VERSION"
	data, err := os.ReadFile(versionFile)
	if err != nil {
		// Log the error and send a 500 Internal Server Error response
		log.Printf("Error reading VERSION file: %v", err)
		http.Error(w, "Failed to read version file", http.StatusInternalServerError)
		return
	}

	// Create a response with the version content
	version := string(data)

	// Prepare the response struct
	response := shared.VersionResponse{
		Version: version,
	}

	// Set the response header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Marshal the response to JSON and write to the HTTP response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding version response to JSON: %v", err)
		http.Error(w, "Failed to encode version to JSON", http.StatusInternalServerError)
	}
}

package server

import (
	"encoding/json"
	"github.com/b4bay/aspm/internal/shared"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type RequestBody struct {
	Data string `json:"data"`
}

func CollectHandler(w http.ResponseWriter, r *http.Request) {
	var body shared.CollectMessageBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	for _, report := range body.Reports {
		DB.Transaction(func(tx *gorm.DB) error {
			// Ensure Product exists
			var product shared.Product
			if err := tx.FirstOrCreate(&product, shared.Product{
				ID: body.ArtefactId,
			}).Error; err != nil {
				http.Error(w, "Failed to create or find product", http.StatusInternalServerError)
				return err
			}

			// Check and update empty fields in Product
			if product.CreatedAt.IsZero() {
				product.CreatedAt = time.Now()
				if err := tx.Save(&product).Error; err != nil {
					http.Error(w, "Failed to update product", http.StatusInternalServerError)
					return err
				}
			}

			// Save Engagement
			engagement := shared.Engagement{
				CreatedAt: time.Now(),
				ProductID: product.ID,
				Tool:      "unknown", // TODO: Get tool name from report
				RawReport: report,
			}
			if result := tx.Create(&engagement); result.Error != nil {
				http.Error(w, "Failed to create engagement", http.StatusInternalServerError)
				return err
			}

			return nil
		})

	}

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

	if body.ProductionMethod == "" {
		body.ProductionMethod = shared.ProductionMethodDefault
	}

	var project = GetProjectFromEnvironment(body.Environment)
	var worker = GetWorkerFromEnvironment(body.Environment)
	var author string
	if body.Product.Author != "" {
		author = body.Product.Author
	} else {
		author = GetAuthorFromEnvironment(body.Environment)
	}

	DB.Transaction(func(tx *gorm.DB) error {
		// Ensure Product exists
		var product shared.Product
		if err := tx.FirstOrCreate(&product, shared.Product{
			ID: body.Product.Id,
		}).Error; err != nil {
			http.Error(w, "Failed to create or find product", http.StatusInternalServerError)
			return err
		}

		var needToUpdate = false
		// Check and update empty fields in Product
		if product.CreatedAt.IsZero() {
			product.CreatedAt = time.Now()
			needToUpdate = true
		}
		if product.Name == "" && body.Product.Name != "" {
			product.Name = body.Product.Name
			needToUpdate = true
		}
		if product.Type == "" && body.Product.Type != "" {
			product.Type = body.Product.Type
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
		for _, o := range body.Origins {
			var origin shared.Product
			if err := tx.FirstOrCreate(&origin, shared.Product{
				ID: o.Id,
			}).Error; err != nil {
				http.Error(w, "Failed to create or find origin", http.StatusInternalServerError)
				return err
			}

			needToUpdate = false
			// Check and update empty fields in Product
			if origin.CreatedAt.IsZero() {
				origin.CreatedAt = time.Now()
				needToUpdate = true
			}
			if origin.Name == "" && o.Name != "" {
				origin.Name = o.Name
				needToUpdate = true
			}
			if origin.Type == "" && o.Type != "" {
				origin.Type = o.Type
				needToUpdate = true
			}
			if origin.Author == "" && o.Author != "" {
				origin.Author = o.Author
				needToUpdate = true
			}

			// Save updated product if necessary
			if needToUpdate {
				if err := tx.Save(&origin).Error; err != nil {
					http.Error(w, "Failed to update origin", http.StatusInternalServerError)
					return err
				}
			}

			var link = shared.Link{}
			if err := tx.FirstOrCreate(&link, shared.Link{
				ProductID: product.ID,
				OriginID:  origin.ID,
				Type:      body.ProductionMethod,
			}).Error; err != nil {
				http.Error(w, "Failed to create or find link", http.StatusInternalServerError)
				return err
			}

			needToUpdate = false
			// Check and update empty fields in Product
			if link.CreatedAt.IsZero() {
				link.CreatedAt = time.Now()
				needToUpdate = true
			}

			// Save updated link if necessary
			if needToUpdate {
				if err := tx.Save(&link).Error; err != nil {
					http.Error(w, "Failed to update link", http.StatusInternalServerError)
					return err
				}
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

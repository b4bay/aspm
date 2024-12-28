package server

import (
	"encoding/json"
	"gorm.io/gorm/clause"
	"log"
	"net/http"
	"os"
	"strings"
)

func UIProductHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch all products from the database
	var products []Product
	if err := DB.Preload(clause.Associations).Find(&products).Error; err != nil {
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}

	// Map products to ProductResponse
	productResponses := []ProductResponse{}
	for _, product := range products {
		productResponses = append(productResponses, ProductResponse{
			ID:        product.ID,
			ProductID: product.ProductID,
			Name:      product.Name,
			Type:      product.Type,
			Project:   product.Project,
			Author:    product.Author,
			Worker:    product.Worker,
			CreatedAt: product.CreatedAt,
		})
	}

	// Set the response header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Marshal the products into JSON and write to the response
	if err := json.NewEncoder(w).Encode(productResponses); err != nil {
		http.Error(w, "Failed to encode products to JSON", http.StatusInternalServerError)
		return
	}
}

func UILinkHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch all links from the database
	var links []Link
	if err := DB.Preload(clause.Associations).Find(&links).Error; err != nil {
		http.Error(w, "Failed to fetch links", http.StatusInternalServerError)
		return
	}

	// Map links to LinkResponse
	linkResponses := []LinkResponse{}
	for _, link := range links {
		linkResponses = append(linkResponses, LinkResponse{
			ID:        link.ID,
			ProductID: link.Product.ProductID,
			OriginID:  link.Origin.ProductID,
			Type:      link.Type,
			CreatedAt: link.CreatedAt,
		})
	}

	// Set the response header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Marshal the links into JSON and write to the response
	if err := json.NewEncoder(w).Encode(linkResponses); err != nil {
		http.Error(w, "Failed to encode links to JSON", http.StatusInternalServerError)
		return
	}
}

func UIEngagementHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch all engagements from the database
	var engagements []Engagement
	if err := DB.Preload(clause.Associations).Find(&engagements).Error; err != nil {
		http.Error(w, "Failed to fetch engagements", http.StatusInternalServerError)
		return
	}

	// Map Engagement to EngagementResponse
	engagementResponses := []EngagementResponse{}
	for _, engagement := range engagements {
		engagementResponses = append(engagementResponses, EngagementResponse{
			ID:           engagement.ID,
			ProductID:    engagement.Product.ProductID,
			Tool:         engagement.Tool,
			ReportLength: len(engagement.RawReport),
			CreatedAt:    engagement.CreatedAt,
		})
	}

	// Set the response header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Marshal the engagements into JSON and write to the response
	if err := json.NewEncoder(w).Encode(engagementResponses); err != nil {
		http.Error(w, "Failed to encode engagements to JSON", http.StatusInternalServerError)
		return
	}
}

func UIVulnerabilityHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch all links from the database
	var vulnerabilities []Vulnerability
	if err := DB.Preload(clause.Associations).Find(&vulnerabilities).Error; err != nil {
		http.Error(w, "Failed to fetch vulnerabilities", http.StatusInternalServerError)
		return
	}

	// Map Engagement to EngagementResponse
	vulnerabilityResponses := []VulnerabilityResponse{}
	for _, vulnerability := range vulnerabilities {
		vulnerabilityResponses = append(vulnerabilityResponses, VulnerabilityResponse{
			ID:              vulnerability.ID,
			VulnerabilityID: vulnerability.VulnerabilityID,
			ProductID:       vulnerability.Product.ProductID,
			LocationHash:    vulnerability.LocationHash,
			Level:           vulnerability.Level,
			Text:            vulnerability.Text,
			CWE:             vulnerability.CWE,
			CVE:             vulnerability.CVE,
			EngagementID:    vulnerability.Engagement.ID,
			CreatedAt:       vulnerability.CreatedAt,
		})
	}

	// Set the response header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Marshal the vulnerabilities into JSON and write to the response
	if err := json.NewEncoder(w).Encode(vulnerabilityResponses); err != nil {
		http.Error(w, "Failed to encode engagements to JSON", http.StatusInternalServerError)
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
	version := strings.TrimSuffix(string(data), "\n")

	// Prepare the response struct
	response := VersionResponse{
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

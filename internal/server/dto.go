package server

import (
	"github.com/b4bay/aspm/internal/server/sarif"
	"github.com/b4bay/aspm/internal/shared"
	"time"
)

type EngagementResponse struct {
	ID           uint      `json:"id"`
	ProductID    string    `json:"product_id"`
	Tool         string    `json:"tool"`
	ReportLength int       `json:"report_length"`
	CreatedAt    time.Time `json:"created_at"`
}

type ProductResponse struct {
	ID        uint                `json:"id"`
	ProductID string              `json:"product_id"`
	Name      string              `json:"name"`
	Type      shared.ArtefactType `json:"type"`
	Project   string              `json:"project"`
	Author    string              `json:"author"`
	Worker    string              `json:"worker"`
	CreatedAt time.Time           `json:"created_at"`
}

type LinkResponse struct {
	ID        uint                    `json:"id"`
	ProductID string                  `json:"product_id"`
	OriginID  string                  `json:"origin_id"`
	Type      shared.ProductionMethod `json:"type"`
	CreatedAt time.Time               `json:"created_at"`
}

type VersionResponse struct {
	Version string `json:"version"`
}

type VulnerabilityResponse struct {
	ID              uint   `json:"id"`
	VulnerabilityID string `json:"vuln_id"`
	LocationHash    string `json:"location_hash"`
	ProductID       string `json:"product_id"`
	Level           sarif.Level
	Text            string
	CWE             string
	CVE             string
	EngagementID    uint      `gorm:"index;not null"`
	CreatedAt       time.Time `json:"created_at"`
}

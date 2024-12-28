package server

import (
	"github.com/b4bay/aspm/internal/shared"
	"gorm.io/gorm"
)

type Link struct {
	gorm.Model
	ProductID string `gorm:"index;not null"`
	OriginID  string `gorm:"index;not null"`
	Type      shared.ProductionMethod
	// Associations
	Product Product `gorm:"constraint:OnDelete:CASCADE;foreignKey:ProductID;references:ProductID"`
	Origin  Product `gorm:"constraint:OnDelete:CASCADE;foreignKey:OriginID;references:ProductID"`
}

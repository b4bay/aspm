package server

import (
	"github.com/b4bay/aspm/internal/shared"
	"gorm.io/gorm"
)

type Link struct {
	gorm.Model
	ProductID uint `gorm:"index;not null"`
	OriginID  uint `gorm:"index;not null"`
	Type      shared.ProductionMethod
	// Associations
	Product Product `gorm:"constraint:OnDelete:CASCADE;foreignKey:ProductID;references:ID"`
	Origin  Product `gorm:"constraint:OnDelete:CASCADE;foreignKey:OriginID;references:ID"`
}

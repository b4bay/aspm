package server

import (
	"github.com/b4bay/aspm/internal/shared"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	ProductID string `gorm:"index,unique,not null"`
	Name      string
	Type      shared.ArtefactType
	Project   string
	Author    string
	Worker    string
}

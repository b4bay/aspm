package server

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

var DB *gorm.DB

func InitDB() error {
	var err error

	// Get datasource name from environment variable or use in-memory DB by default
	datasourceName := os.Getenv("DATASOURCE_NAME")
	if datasourceName == "" {
		datasourceName = "file::memory:?cache=shared"
	}

	DB, err = gorm.Open(sqlite.Open(datasourceName), &gorm.Config{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(&Product{}, &Link{}, &Engagement{}, &Vulnerability{}, &Status{})

	return err
}

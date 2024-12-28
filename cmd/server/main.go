package main

import (
	"fmt"
	"github.com/b4bay/aspm/internal/server"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
	"log"
	"net/http"
)

var db *gorm.DB

func init() {
	if err := server.InitDB(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("POST /api/v1/collect", server.CollectHandler)
	http.HandleFunc("POST /api/v1/origin", server.OriginHandler)
	http.HandleFunc("GET /api/v1/gw", server.GWHandler)
	http.HandleFunc("GET /api/v1/ui/product", server.UIProductHandler)
	http.HandleFunc("GET /api/v1/ui/link", server.UILinkHandler)
	http.HandleFunc("GET /api/v1/ui/engagement", server.UIEngagementHandler)
	http.HandleFunc("GET /api/v1/ui/vulnerability", server.UIVulnerabilityHandler)
	http.HandleFunc("GET /api/v1/ui/version", server.UIVersionHandler)

	fmt.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

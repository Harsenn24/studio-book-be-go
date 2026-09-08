package app

import (
	"log"
	"os"
	"studio-book-be-go/internal/config"
	"studio-book-be-go/internal/route"
	"studio-book-be-go/internal/service/auth"
	"github.com/gin-gonic/gin"
)

func StartApp() {
	db := config.ConnectDB()

	authModule := auth.InitAuthModule(db)
	

	router := gin.New()

	configRouter := route.RouterConfig{
		AuthHandler: authModule.Handler, 
	}

	route.SetupRoutes(router, configRouter)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server berhasil berjalan di port :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server Gin: %v", err)
	}
}
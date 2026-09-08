package route

import (
	"studio-book-be-go/internal/service/auth"
	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	AuthHandler *auth.AuthHandler
}

func SetupRoutes(router *gin.Engine, config RouterConfig) {
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	MapUserRoutes(router, config)
}
package route

import "github.com/gin-gonic/gin"

func MapUserRoutes(router *gin.Engine, config RouterConfig) {
	userRoutes := router.Group("/:role")
	{
		userRoutes.POST("/auth/login", config.AuthHandler.LoginHandler)
		userRoutes.POST("/auth/register", config.AuthHandler.RegisterHandler)
		userRoutes.POST("/auth/check-user", config.AuthHandler.CheckUserHandler)
	}
}
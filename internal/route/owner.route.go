package route

import "github.com/gin-gonic/gin"

func MapOwnerRoutes(router *gin.Engine, config RouterConfig) {
	ownerRoutes := router.Group("/:role")
	{
		ownerRoutes.POST("/address/province", config.AddressHandler.ProvincePagination)
	}
}
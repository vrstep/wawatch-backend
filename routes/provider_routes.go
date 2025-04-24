package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/vrstep/wawatch-backend/controller"
	"github.com/vrstep/wawatch-backend/middleware"
)

func ProviderRoute(router *gin.Engine) {
	// Assuming provider management requires authentication (and maybe admin role check in controller)
	providers := router.Group("/providers")
	providers.Use(middleware.RequireAuth) // Or a specific Admin middleware
	{
		providers.PUT("/:provider_id", controller.UpdateWatchProvider)    // New Endpoint 8
		providers.DELETE("/:provider_id", controller.DeleteWatchProvider) // New Endpoint 7
	}
}

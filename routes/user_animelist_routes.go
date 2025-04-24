package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/vrstep/wawatch-backend/controller"
	"github.com/vrstep/wawatch-backend/middleware"
)

func UserAnimeListRoute(router *gin.Engine) {
	list := router.Group("/animelist")
	list.Use(middleware.RequireAuth) // All routes require authentication
	{
		// Get user's anime list (optionally filtered by status)
		list.GET("/", controller.GetUserAnimeList)

		// Add anime to list or update if already exists
		list.POST("/", controller.AddToAnimeList)

		// Update a specific list entry
		list.PATCH("/:id", controller.UpdateListEntry)

		// Delete a list entry
		list.DELETE("/:id", controller.DeleteListEntry)

		list.GET("/stats", controller.GetUserAnimeListStats) // New Endpoint 3

	}
}

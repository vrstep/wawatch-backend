package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/vrstep/wawatch-backend/controller"
	"github.com/vrstep/wawatch-backend/middleware"
)

func AnimeRoute(router *gin.Engine) {
	anime := router.Group("/anime")
	{
		anime.GET("/search", middleware.RequireAuth, controller.SearchAnime)
		anime.GET("/:id", controller.GetAnimeDetails)

		// Public discovery endpoints
		anime.GET("/popular", controller.GetPopularAnime)               // New Endpoint 4
		anime.GET("/trending", controller.GetTrendingAnime)             // New Endpoint 5
		anime.GET("/season/:year/:season", controller.GetAnimeBySeason) // New Endpoint 6

		// Protected routes
		anime.POST("/provider", middleware.RequireAuth, controller.AddWatchProvider)
		anime.GET("/recommendations", middleware.RequireAuth, controller.GetAnimeRecommendations) // New Endpoint 10

		anime.GET("/:id/list-status", middleware.RequireAuth, controller.GetAnimeInUserList)
	}
}

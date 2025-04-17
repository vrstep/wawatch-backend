package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/vrstep/wawatch-backend/controller"
	"github.com/vrstep/wawatch-backend/middleware"
)

func AnimeRoute(router *gin.Engine) {
	anime := router.Group("/anime")
	{
		anime.GET("/search", controller.SearchAnime)
		anime.GET("/:id", controller.GetAnimeDetails)

		// Protected routes
		anime.POST("/provider", middleware.RequireAuth, controller.AddWatchProvider)

		anime.GET("/:id/list-status", middleware.RequireAuth, controller.GetAnimeInUserList)
	}
}

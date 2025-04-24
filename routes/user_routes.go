package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/vrstep/wawatch-backend/controller"
	"github.com/vrstep/wawatch-backend/middleware"
)

func UserRoute(router *gin.Engine) {
	// ...existing code...
	router.POST("/signup", controller.Signup)
	router.POST("/login", controller.Login)
	router.GET("/validate", middleware.RequireAuth, controller.Validate)

	// Profile routes (require authentication)
	profile := router.Group("/profile")
	profile.Use(middleware.RequireAuth)
	{
		profile.GET("/", controller.GetMyProfile)    // New Endpoint 1
		profile.PUT("/", controller.UpdateMyProfile) // New Endpoint 2
	}

	// Public user list view
	router.GET("/users/:username/animelist", controller.GetUserPublicAnimeList) // New Endpoint 9
}

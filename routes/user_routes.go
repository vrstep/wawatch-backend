package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/vrstep/wawatch-backend/controller"
	"github.com/vrstep/wawatch-backend/middleware"
)

func UserRoute(router *gin.Engine) {
	router.GET("/", controller.GetUsers)
	router.POST("/", controller.CreateUser)
	router.GET("/:id", controller.GetUser)
	router.PUT("/:id", middleware.RequireAuth, controller.UpdateUser)
	router.DELETE("/:id", middleware.RequireAuth, controller.DeleteUser)

	router.POST("/signup", controller.Signup)
	router.POST("/login", controller.Login)
	router.GET("/validate", middleware.RequireAuth, controller.Validate)
}

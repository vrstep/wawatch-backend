package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/vrstep/wawatch-backend/controller"
)

func UserRoute(router *gin.Engine) {
	router.GET("/", controller.GetUsers)
	router.POST("/", controller.CreateUser)
	router.GET("/:id", controller.GetUser)
	router.PUT("/:id", controller.UpdateUser)
	router.DELETE("/:id", controller.DeleteUser)
}

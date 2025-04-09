package routes

import "github.com/gin-gonic/gin"

func UserRoute(router *gin.Engine) {
	router.GET("/", func(c *gin.Context) {
		c.String(200, "User route")
	})
}

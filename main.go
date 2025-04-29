package main

import (
	"github.com/gin-gonic/gin"
	"github.com/vrstep/wawatch-backend/config"
	"github.com/vrstep/wawatch-backend/middleware"
	"github.com/vrstep/wawatch-backend/routes"
)

func main() {
	router := gin.New()
	router.Use(middleware.Logging())

	config.ConnectDB()

	// Apply CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup routes
	routes.UserRoute(router)
	routes.AnimeRoute(router)
	routes.UserAnimeListRoute(router)
	routes.ProviderRoute(router)

	router.Run(":8080")
	router.Run(":8081")
}

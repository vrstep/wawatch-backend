package main

import (
	"github.com/gin-gonic/gin"
	"github.com/vrstep/wawatch-backend/config"
	"github.com/vrstep/wawatch-backend/routes"
)

func main() {

	router := gin.New()

	config.ConnectDB()

	routes.UserRoute(router)

	router.Run(":8080")

}

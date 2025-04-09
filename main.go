package main

import (
	"github.com/gin-gonic/gin"
	"github.com/vrstep/wawatch-backend/routes"
)

func main() {

	router := gin.New()

	routes.UserRoute(router)

	router.Run(":8080")

}

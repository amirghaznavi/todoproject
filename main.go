package main

import (
	"log"
	"todoproject/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Add your logging middleware here if needed

	routes.AuthRoutes(r)

	log.Println("Server started at http://localhost:8080")
	r.Run(":8080")
}

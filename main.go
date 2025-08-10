package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create a new Gin router with default middleware (logger & recovery)
	router := gin.Default()

	// Serve static files (CSS, JS, images) from "static" folder
	router.Static("/static", "./static")

	// Load HTML templates from "templates" folder
	router.LoadHTMLGlob("templates/*")

	// Home page route
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "Todo Project",
		})
	})

	// Start the server
	log.Println("🚀 Starting Todo Project server on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}

package main

import (
	"backend/db"
	"backend/router"

	"github.com/gin-gonic/gin"
)

func main() {
	db.StartDbEngine()
	engine := gin.New()
	router.MapUrls(engine)

	// Endpoint de salud
	engine.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})

	engine.Run(":8080")
}

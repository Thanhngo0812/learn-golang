package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()

	r.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Hello World từ Gin và Docker!",
		})
	})

	// Server chạy ở cổng 8080 bên trong container
	r.Run(":8080")
}
package middleware

import (
	"log"
	"time"
	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now() // Bắt đầu đếm giờ

		// Xử lý request
		c.Next()

		// Sau khi xử lý xong
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		// Log định dạng: [STATUS] | LATENCY | METHOD | PATH
		log.Printf("[GIN] %3d | %13v | %-7s | %s\n",
			status,
			latency,
			method,
			path,
		)
	}
}
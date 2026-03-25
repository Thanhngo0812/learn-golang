package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// NoCache is a middleware that prevents the browser from caching the response.
// This is useful for static files during development to ensure the latest frontend code is always served.
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", time.Unix(0, 0).Format(time.RFC1123))
		c.Next()
	}
}

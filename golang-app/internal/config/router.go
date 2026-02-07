package config // Vì bạn đang để chung package config

import (
	"github.com/gin-gonic/gin"
)

// Sửa tham số đầu vào từ *controller.UserController thành *Server
func NewRouter(server *Server) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		// Lấy UserController từ trong cái túi Server ra dùng
		v1.POST("/users", server.UserController.CreateUser)
	}

	return r
}
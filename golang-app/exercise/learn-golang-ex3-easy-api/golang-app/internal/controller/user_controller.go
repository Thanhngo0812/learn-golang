package controller

import (
	"net/http"
	"golang-app/internal/model"
	"golang-app/internal/service"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	service service.UserService
}

func NewUserController(service service.UserService) *UserController {
	return &UserController{service: service}
}

// CreateUser godoc
func (uc *UserController) CreateUser(c *gin.Context) {
	var newUser model.User

	// 1. Bind JSON từ request body vào struct (có kiểm tra validate required)
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Gọi Service để xử lý
	if err := uc.service.CreateUser(&newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 3. Trả về kết quả thành công
	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
		"data":    newUser,
	})
}
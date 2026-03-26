package model

// User đại diện cho bảng users trong database
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name" binding:"required"` // binding:"required" để Gin tự validate
	Phone string `json:"phone" binding:"required"`
}
package dto

// DTO cho Request đăng ký
type RegisterRequest struct {
	FullName    string `json:"full_name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	PhoneNumber string `json:"phone_number" binding:"required,len=10"`
	Password    string `json:"password" binding:"required,min=6"`
	Role        string `json:"role" binding:"omitempty,oneof=admin seller bidder"`
}
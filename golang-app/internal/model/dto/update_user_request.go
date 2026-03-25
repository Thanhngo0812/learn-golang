package dto

type UpdateUserRequest struct {
	FullName    *string `json:"full_name,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Password    *string `json:"password,omitempty" binding:"omitempty,min=6"`
	IsActive    *bool   `json:"is_active,omitempty"` // admin dùng
	Role        *string `json:"role,omitempty"`      // admin dùng
}

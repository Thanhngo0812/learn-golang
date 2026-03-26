package service

import (
	"golang-app/internal/model"
	"golang-app/internal/repository"
)

// UserService định nghĩa các nghiệp vụ liên quan đến User
type UserService interface {
	CreateUser(user *model.User) error
}

type userService struct {
	repo repository.UserRepository
}

// NewUserService khởi tạo Service và tiêm (Inject) Repository vào
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(user *model.User) error {
	// Tại đây có thể thêm logic nghiệp vụ, ví dụ: Validate số điện thoại
	// Hiện tại gọi trực tiếp xuống Repo
	return s.repo.CreateUser(user)
}
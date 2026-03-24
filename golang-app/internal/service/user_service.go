package service

import (
	"golang-app/internal/config"
	"golang-app/internal/model/dto"
	"golang-app/internal/model/entity"
	"golang-app/internal/repository"
	"golang-app/pkg/apperror"
	"strings"
)

type UserService interface {
	Register(req *dto.RegisterRequest) (*entity.User, error)
	PublicRegister(req *dto.RegisterRequest) (*entity.User, error)
	Login(req *dto.LoginRequest, cfg *config.Config) (*dto.LoginResponse, error)
	GetListUsers() ([]entity.User, error)
	GetUserDetail(id int) (*entity.User, error)
	GetMyProfile(userID uint) (*entity.User, error)
	UpdateMyProfile(userID uint, req *dto.UpdateUserRequest) (*entity.User, error)
	UpdateUser(id int, req *dto.UpdateUserRequest) (*entity.User, error)
	DeleteUser(id int) error
	LockUser(id int) (*entity.User, error)
	UnlockUser(id int) (*entity.User, error)
}
type userService struct {
	repo    repository.UserRepository
	bidRepo repository.BidRepository
}

func NewUserService(repo repository.UserRepository, bidRepo repository.BidRepository) UserService {
	return &userService{repo: repo, bidRepo: bidRepo}
}

func (s *userService) Register(req *dto.RegisterRequest) (*entity.User, error) {
	// 1. Chuẩn bị dữ liệu (Mapping DTO -> Entity)
	userRole := "bidder"
	if req.Role != "" {
		userRole = req.Role
	}

	newUser := &entity.User{
		FullName:     req.FullName,
		Email:        req.Email,
		PhoneNumber:  req.PhoneNumber,
		PasswordHash: req.Password, // Hook Entity sẽ lo việc hash
		Role:         userRole,
		IsActive:     true,
	}
	if req.Role == "admin" {
		if err := s.repo.Create(newUser); err != nil {
			if err.Error() == "conflict" {
				return nil, apperror.NewConflict(nil, "email hoặc số điện thoại đã tồn tại")
			}
			return nil, apperror.NewInternal(err)
		}
	} else {
		newWallet := &entity.Wallet{} // Tạo struct rỗng, Repo sẽ điền UserID và Balance
		// 2. Gọi Repository để thực thi Transaction
		// Service không cần biết bên dưới là SQL Transaction hay gì cả
		err := s.repo.RegisterUserTx(newUser, newWallet)

		if err != nil {
			// Mapping lỗi từ Repo sang lỗi App (nếu cần)
			if strings.HasPrefix(err.Error(), "conflict") {
				return nil, apperror.NewConflict(nil, "email hoặc số điện thoại đã tồn tại")
			}
			return nil, apperror.NewInternal(err)
		}
		// 3. Gán ví vào user để trả về client luôn
		newUser.Wallet = newWallet
	}

	return newUser, nil
}

func (s *userService) PublicRegister(req *dto.RegisterRequest) (*entity.User, error) {
	// 1. Force the role to specific values (ignore admin attempts here)
	if req.Role == "admin" || req.Role == "" {
		return nil, apperror.NewBadRequest(nil, "Chỉ được phép đăng ký vai trò 'bidder' hoặc 'seller'")
	}

	newUser := &entity.User{
		FullName:     req.FullName,
		Email:        req.Email,
		PhoneNumber:  req.PhoneNumber,
		PasswordHash: req.Password, // Hook Entity sẽ lo việc hash
		Role:         req.Role,     // Either bidder or seller
		IsActive:     true,
	}

	newWallet := &entity.Wallet{}
	err := s.repo.RegisterUserTx(newUser, newWallet)

	if err != nil {
		if strings.HasPrefix(err.Error(), "conflict") {
			return nil, apperror.NewConflict(nil, "email hoặc số điện thoại đã tồn tại")
		}
		return nil, apperror.NewInternal(err)
	}

	newUser.Wallet = newWallet
	return newUser, nil
}

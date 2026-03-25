package service

import (
	"errors"
	"golang-app/internal/config"
	"golang-app/internal/model/dto"
	"golang-app/internal/model/entity"
	"golang-app/internal/repository"
	"golang-app/pkg/apperror"
	"golang-app/pkg/utils"
	"strings"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"
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
	}else{
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

func (s *userService) Login(req *dto.LoginRequest, cfg *config.Config) (*dto.LoginResponse, error) {
	// 1. Tìm user theo email
	user, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound(err, "Email không tồn tại")
		}
		return nil, apperror.NewInternal(err)
	}

	// 2. Kiểm tra tài khoản có bị khóa không
	if !user.IsActive {
		return nil, apperror.NewBadRequest(nil, "Tài khoản của bạn đã bị khóa. Vui lòng liên hệ Admin")
	}

	// 3. So sánh mật khẩu
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, apperror.NewBadRequest(err, "Mật khẩu không chính xác")
	}

	// 4. Tạo JWT Token
	token, err := utils.GenerateToken(user.ID, user.Role, cfg.App.JWTSecret, cfg.App.JWTExpiration)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}

	// 5. Trả về response
	return &dto.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// 1. Get List
func (s *userService) GetListUsers() ([]entity.User, error) {
	users, err := s.repo.GetAll()
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return users, nil
}

// 2. Get Detail
func (s *userService) GetUserDetail(id int) (*entity.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound(err, "Người dùng không tồn tại")
		}
		return nil, apperror.NewInternal(err)
	}
	return user, nil
}

// 3. Update User
func (s *userService) UpdateUser(id int, req *dto.UpdateUserRequest) (*entity.User, error) {
	// Bước 1: Tìm user cũ trước
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound(err, "Không tìm thấy người dùng để cập nhật")
		}
		return nil, apperror.NewInternal(err)
	}

	// Bước 2: Cập nhật thông tin mới (nếu có gửi lên)
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// Bước 3: Lưu xuống DB
	if err := s.repo.Update(user); err != nil {
		return nil, apperror.NewInternal(err)
	}

	return user, nil
}

// 4. Delete User
func (s *userService) DeleteUser(id int) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NewNotFound(err, "Người dùng không tồn tại")
		}
		return apperror.NewInternal(err)
	}
	if user.Role == "admin" {
		return apperror.NewBadRequest(nil, "Không thể xóa người dùng có quyền Quản trị viên")
	}

	// Kiểm tra user đã tham gia bid chưa
	hasBids, err := s.bidRepo.HasUserBids(user.ID)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if hasBids {
		return apperror.NewBadRequest(nil, "Không thể xóa người dùng đã tham gia đấu giá. Hãy khóa tài khoản thay vì xóa")
	}

	// Kiểm tra user có sản phẩm không
	hasProducts, err := s.repo.HasProducts(user.ID)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if hasProducts {
		return apperror.NewBadRequest(nil, "Không thể xóa người dùng đã có sản phẩm. Hãy khóa tài khoản thay vì xóa")
	}

	if err := s.repo.Delete(id); err != nil {
		return apperror.NewInternal(err)
	}
	return nil
}

// GetMyProfile lấy thông tin cá nhân kèm ví
func (s *userService) GetMyProfile(userID uint) (*entity.User, error) {
	user, err := s.repo.GetByIDWithWallet(int(userID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound(err, "Người dùng không tồn tại")
		}
		return nil, apperror.NewInternal(err)
	}
	return user, nil
}

// UpdateMyProfile cập nhật thông tin cá nhân (chỉ full_name, phone_number)
func (s *userService) UpdateMyProfile(userID uint, req *dto.UpdateUserRequest) (*entity.User, error) {
	user, err := s.repo.GetByID(int(userID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound(err, "Người dùng không tồn tại")
		}
		return nil, apperror.NewInternal(err)
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}
	// Không cho sửa IsActive từ đây

	if err := s.repo.Update(user); err != nil {
		return nil, apperror.NewInternal(err)
	}
	return user, nil
}

// LockUser khóa tài khoản (Admin only)
func (s *userService) LockUser(id int) (*entity.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound(err, "Người dùng không tồn tại")
		}
		return nil, apperror.NewInternal(err)
	}
	if user.Role == "admin" {
		return nil, apperror.NewBadRequest(nil, "Không thể khóa tài khoản Admin")
	}
	user.IsActive = false
	if err := s.repo.Update(user); err != nil {
		return nil, apperror.NewInternal(err)
	}
	return user, nil
}

// UnlockUser mở khóa tài khoản (Admin only)
func (s *userService) UnlockUser(id int) (*entity.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound(err, "Người dùng không tồn tại")
		}
		return nil, apperror.NewInternal(err)
	}
	user.IsActive = true
	if err := s.repo.Update(user); err != nil {
		return nil, apperror.NewInternal(err)
	}
	return user, nil
}
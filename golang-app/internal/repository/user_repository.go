package repository

import (
	"errors"
	"golang-app/internal/model/entity"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *entity.User) error
	RegisterUserTx(user *entity.User, wallet *entity.Wallet) error
	GetByEmail(email string) (*entity.User, error)
	GetByID(id int) (*entity.User, error)
	GetByIDWithWallet(id int) (*entity.User, error)
	GetAll() ([]entity.User, error)
	Update(user *entity.User) error
	Delete(id int) error
	HasProducts(userID uint) (bool, error)
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

// Create: Thêm user mới
func (r *userRepo) Create(user *entity.User) error {
	// GORM tự động sinh SQL INSERT và trả về ID vào struct user
	// 1. Kiểm tra trùng lặp (Dùng Scope ngay trong Repo)
	var existingUser entity.User
	// Lưu ý: Gọi Scope từ tầng Entity
	err := r.db.Scopes(entity.UserByEmailOrPhone(user.Email, user.PhoneNumber)).First(&existingUser).Error
	if err == nil {
		return errors.New("conflict")
	}
	if err != gorm.ErrRecordNotFound {
		return err // Lỗi DB khác
	}

	return r.db.Create(user).Error
}

func (r *userRepo) RegisterUserTx(user *entity.User, wallet *entity.Wallet) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Kiểm tra trùng lặp (Dùng Scope ngay trong Repo)
		var existingUser entity.User
		// Lưu ý: Gọi Scope từ tầng Entity
		err := tx.Scopes(entity.UserByEmailOrPhone(user.Email, user.PhoneNumber)).First(&existingUser).Error

		if err == nil {
			return errors.New("conflict: email hoặc số điện thoại đã tồn tại")
		}
		if err != gorm.ErrRecordNotFound {
			return err // Lỗi DB khác
		}

		// 2. Tạo User (Hook BeforeSave trong Entity sẽ tự Hash pass)
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// 3. Gán ID user vừa tạo cho Wallet
		wallet.UserID = user.ID
		wallet.Balance = 0
		wallet.FrozenBalance = 0

		// 4. Tạo Wallet
		if err := tx.Create(wallet).Error; err != nil {
			return err // Sẽ tự động Rollback bước 2
		}

		// 5. Nếu return nil -> Auto Commit
		return nil
	})
}

// GetByEmail: Tìm user theo email
func (r *userRepo) GetByEmail(email string) (*entity.User, error) {
	var user entity.User
	// SELECT * FROM users WHERE email = ? LIMIT 1
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 1. Read Detail
func (r *userRepo) GetByID(id int) (*entity.User, error) {
	var user entity.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 2. Read List (Có thể thêm phân trang sau này)
func (r *userRepo) GetAll() ([]entity.User, error) {
	var users []entity.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// 3. Update
func (r *userRepo) Update(user *entity.User) error {
	// Save sẽ update tất cả các trường, bao gồm cả UpdatedAt
	return r.db.Save(user).Error
}

// 4. Delete
func (r *userRepo) Delete(id int) error {
	return r.db.Delete(&entity.User{}, id).Error
}

// GetByIDWithWallet lấy user kèm thông tin ví
func (r *userRepo) GetByIDWithWallet(id int) (*entity.User, error) {
	var user entity.User
	if err := r.db.Scopes(entity.WithWallet).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// HasProducts kiểm tra user có sản phẩm không
func (r *userRepo) HasProducts(userID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.Product{}).Where("seller_id = ?", userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

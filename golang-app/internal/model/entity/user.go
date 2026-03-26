package entity

import (
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User Entity
type User struct {
	ID           uint      `gorm:"primaryKey;column:id" json:"id"`
	FullName     string    `gorm:"column:full_name;not null" json:"full_name"`
	Email        string    `gorm:"column:email;unique;not null" json:"email"`
	PhoneNumber  string    `gorm:"column:phone_number;not null" json:"phone_number"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	Role         string    `gorm:"column:role;default:'bidder'" json:"role"`
	IsActive     bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`

	// Relationships
	Wallet *Wallet `gorm:"foreignKey:UserID" json:"wallet,omitempty"`
}

func (User) TableName() string {
	return "users"
}

// ==========================================
// 1. HOOKS (Logic tự động)
// ==========================================

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	// Chuẩn hóa Email
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))

	// Hash Password
	if u.PasswordHash != "" && !strings.HasPrefix(u.PasswordHash, "$2a$") {
		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.PasswordHash = string(hashedPwd)
	}
	return nil
}

// ==========================================
// 2. SCOPES (Tái sử dụng câu truy vấn)
// ==========================================

// Scope: Tìm User theo Email hoặc Số điện thoại
// Dùng để kiểm tra trùng lặp khi đăng ký
func UserByEmailOrPhone(email, phone string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("email = ? OR phone_number =?", email, phone)
		return db.Where("email = ? OR phone_number = ?", email, phone)
	}
}

// Scope: Eager Load (Nạp sẵn) Ví tiền
// Dùng khi xem chi tiết profile
func WithWallet(db *gorm.DB) *gorm.DB {
	return db.Preload("Wallet")
}

// Scope: Chỉ lấy User đang hoạt động
func ActiveUsers(db *gorm.DB) *gorm.DB {
	return db.Where("is_active = ?", true)
}

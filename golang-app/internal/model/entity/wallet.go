package entity

import "time"

// Wallet Entity khớp với bảng wallets trong DB
type Wallet struct {
	// UserID vừa là Khóa chính (PK), vừa là Khóa ngoại (FK) trỏ về Users
	// Tag 'primaryKey' để GORM biết đây là ID
	UserID uint `gorm:"primaryKey;column:user_id" json:"user_id"`

	// Số dư khả dụng (Tiền có thể dùng để bid)
	// Tag 'check:balance >= 0' để tạo ràng buộc SQL, không cho phép âm tiền
	Balance float64 `gorm:"column:balance;default:0;check:balance >= 0" json:"balance"`

	// Số dư đóng băng (Tiền đang bị giữ khi đang bid)
	FrozenBalance float64 `gorm:"column:frozen_balance;default:0" json:"frozen_balance"`

	// Thời gian cập nhật số dư cuối cùng (Rất quan trọng để truy vết)
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	// --- RELATIONSHIP (Optional) ---
	// Belongs To: Ví này thuộc về User nào?
	// (Thường ít dùng chiều này, chủ yếu dùng User.Wallet, nhưng khai báo cho đủ bộ)
	User *User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
}

// TableName: Đảm bảo map đúng vào bảng 'wallets'
func (Wallet) TableName() string {
	return "wallets"
}
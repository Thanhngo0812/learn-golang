package entity

import (
	"time"
)

type Wallet struct {
	ID            uint      `gorm:"primaryKey;column:id" json:"id"`
	UserID        uint      `gorm:"column:user_id;not null;uniqueIndex" json:"user_id"`
	Balance       int64     `gorm:"column:balance;not null;default:0" json:"balance"`
	FrozenBalance int64     `gorm:"column:frozen_balance;not null;default:0" json:"frozen_balance"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`

	// Quan hệ ngược (optional)
	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (Wallet) TableName() string {
	return "wallets"
}

package entity

import (
	"time"
)

type Notification struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	UserID    uint      `gorm:"column:user_id;index;not null" json:"user_id"`
	Title     string    `gorm:"column:title;not null" json:"title"`
	Content   string    `gorm:"column:content;not null" json:"content"`
	Type      string    `gorm:"column:type;not null" json:"type"` // e.g., "outbid", "ended", "status"
	IsRead    bool      `gorm:"column:is_read;default:false" json:"is_read"`
	Link      string    `gorm:"column:link" json:"link"` // Optional link to page (e.g., auction room)
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`

	// Relationship
	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Notification) TableName() string {
	return "notifications"
}

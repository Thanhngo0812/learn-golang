package repository

import (
	"golang-app/internal/model/entity"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(notif *entity.Notification) error
	GetByUserID(userID uint, limit int, offset int) ([]entity.Notification, int64, error)
	MarkAsRead(id uint, userID uint) error
	MarkAllAsRead(userID uint) error
}

type notificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) Create(notif *entity.Notification) error {
	return r.db.Create(notif).Error
}

func (r *notificationRepo) GetByUserID(userID uint, limit int, offset int) ([]entity.Notification, int64, error) {
	var notifs []entity.Notification
	var total int64

	query := r.db.Model(&entity.Notification{}).Where("user_id = ?", userID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&notifs).Error
	return notifs, total, err
}

func (r *notificationRepo) MarkAsRead(id uint, userID uint) error {
	return r.db.Model(&entity.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

func (r *notificationRepo) MarkAllAsRead(userID uint) error {
	return r.db.Model(&entity.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

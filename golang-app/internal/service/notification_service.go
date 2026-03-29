package service

import (
	"golang-app/internal/model/entity"
	"golang-app/internal/repository"
	"golang-app/pkg/websocket"
)

type NotificationService interface {
	NotifyUser(userID uint, title string, content string, nType string, link string) error
	GetNotifications(userID uint, page int, limit int) ([]entity.Notification, int64, error)
	MarkAsRead(id uint, userID uint) error
	MarkAllAsRead(userID uint) error
}

type notificationService struct {
	repo repository.NotificationRepository
	hub  *websocket.Hub
}

func NewNotificationService(repo repository.NotificationRepository, hub *websocket.Hub) NotificationService {
	return &notificationService{repo: repo, hub: hub}
}

func (s *notificationService) NotifyUser(userID uint, title string, content string, nType string, link string) error {
	notif := &entity.Notification{
		UserID:  userID,
		Title:   title,
		Content: content,
		Type:    nType,
		Link:    link,
		IsRead:  false,
	}

	// 1. Save to DB
	if err := s.repo.Create(notif); err != nil {
		return err
	}

	// 2. Push to WebSocket if user is online
	s.hub.PrivateBroadcast <- websocket.UserMessage{
		UserID:  userID,
		Payload: notif,
	}

	return nil
}

func (s *notificationService) GetNotifications(userID uint, page int, limit int) ([]entity.Notification, int64, error) {
	if page <= 0 { page = 1 }
	if limit <= 0 { limit = 20 }
	offset := (page - 1) * limit
	return s.repo.GetByUserID(userID, limit, offset)
}

func (s *notificationService) MarkAsRead(id uint, userID uint) error {
	return s.repo.MarkAsRead(id, userID)
}

func (s *notificationService) MarkAllAsRead(userID uint) error {
	return s.repo.MarkAllAsRead(userID)
}

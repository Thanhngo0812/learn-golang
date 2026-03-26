package dto

import "golang-app/internal/model/entity"

// AuctionListResponse đại diện cho phản hồi danh sách phiên đấu giá
type AuctionListResponse struct {
	Auctions   []entity.Auction `json:"auctions"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
}

// NotificationListResponse đại diện cho phản hồi danh sách thông báo
type NotificationListResponse struct {
	Notifications []entity.Notification `json:"notifications"`
	TotalCount    int64                 `json:"total_count"`
	Page          int                   `json:"page"`
	Limit         int                   `json:"limit"`
}

// ProductListResponse đại diện cho phản hồi danh sách sản phẩm
type ProductListResponse struct {
	Products   []entity.Product `json:"products"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
}

// SimpleMessageResponse đại diện cho phản hồi chỉ có tin nhắn
type SimpleMessageResponse struct {
	Message string `json:"message" example:"Thao tác thành công"`
}

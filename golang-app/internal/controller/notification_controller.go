package controller

import (
	"golang-app/internal/service"
	"golang-app/pkg/response"
	"golang-app/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	service service.NotificationService
}

func NewNotificationController(s service.NotificationService) *NotificationController {
	return &NotificationController{service: s}
}

// GetMyNotifications godoc
// @Summary      Danh sách thông báo của tôi
// @Description  Lấy danh sách các thông báo của người dùng đang đăng nhập
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        page   query     int  false  "Trang"
// @Param        limit  query     int  false  "Số lượng"
// @Success      200    {object}  response.SuccessResponse
// @Failure      401    {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /notifications [get]
func (c *NotificationController) GetMyNotifications(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(uint)
	page, _ := utils.GetIntQuery(ctx, "page", 1)
	limit, _ := utils.GetIntQuery(ctx, "limit", 20)

	notifs, total, err := c.service.GetNotifications(userID, page, limit)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, gin.H{
		"notifications": notifs,
		"total_count":   total,
		"page":          page,
		"limit":         limit,
	})
}

// MarkAsRead godoc
// @Summary      Đánh dấu đã đọc
// @Description  Đánh dấu một thông báo cụ thể là đã đọc
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Notification ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /notifications/{id}/read [patch]
func (c *NotificationController) MarkAsRead(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(uint)
	id, err := utils.GetIDFromParam(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	if err := c.service.MarkAsRead(uint(id), userID); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, gin.H{"message": "Đã đánh dấu là đã đọc"})
}

// MarkAllAsRead godoc
// @Summary      Đánh dấu tất cả đã đọc
// @Description  Đánh dấu toàn bộ thông báo của người dùng là đã đọc
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse
// @Failure      401  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /notifications/read-all [patch]
func (c *NotificationController) MarkAllAsRead(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(uint)

	if err := c.service.MarkAllAsRead(userID); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, gin.H{"message": "Đã đánh dấu tất cả là đã đọc"})
}

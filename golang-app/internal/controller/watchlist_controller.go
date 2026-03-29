package controller

import (
	"golang-app/internal/service"
	"golang-app/pkg/response"
	"golang-app/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WatchlistController struct {
	service service.WatchlistService
}

func NewWatchlistController(s service.WatchlistService) *WatchlistController {
	return &WatchlistController{service: s}
}

// ToggleWatch godoc
// @Summary      Quan tâm/Bỏ quan tâm
// @Description  Thêm hoặc xóa một phiên đấu giá khỏi danh sách quan tâm
// @Tags         watchlist
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Auction ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /watchlist/{id} [post]
func (c *WatchlistController) ToggleWatch(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(uint)
	auctionID, err := utils.GetIDFromParam(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	isWatching, err := c.service.ToggleWatch(userID, uint(auctionID))
	if err != nil {
		response.Error(ctx, err)
		return
	}

	msg := "Đã thêm vào danh sách quan tâm"
	if !isWatching {
		msg = "Đã xóa khỏi danh sách quan tâm"
	}

	response.Success(ctx, http.StatusOK, gin.H{
		"is_watching": isWatching,
		"message":     msg,
	})
}

// GetMyWatchlist godoc
// @Summary      Danh sách quan tâm của tôi
// @Description  Lấy danh sách các phiên đấu giá mà người dùng đang quan tâm
// @Tags         watchlist
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse
// @Failure      401  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /watchlist [get]
func (c *WatchlistController) GetMyWatchlist(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(uint)

	watchlist, err := c.service.GetWatchlist(userID)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, watchlist)
}

// CheckStatus godoc
// @Summary      Kiểm tra trạng thái quan tâm
// @Description  Kiểm tra xem người dùng có đang quan tâm phiên này không
// @Tags         watchlist
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Auction ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /watchlist/{id}/status [get]
func (c *WatchlistController) CheckStatus(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(uint)
	auctionID, err := utils.GetIDFromParam(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	isWatching, err := c.service.IsWatching(userID, uint(auctionID))
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, gin.H{"is_watching": isWatching})
}

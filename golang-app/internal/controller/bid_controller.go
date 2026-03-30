package controller

import (
	"golang-app/internal/service"
	"golang-app/pkg/apperror"
	"golang-app/pkg/response"
	"golang-app/pkg/utils"
	customWebSocket "golang-app/pkg/websocket"
	"net/http"
	"fmt"

	"github.com/gin-gonic/gin"
)

type BidController struct {
	bidService          service.BidService
	auctionService      service.AuctionService
	notificationService service.NotificationService
	hub                 *customWebSocket.Hub
}

func NewBidController(b service.BidService, a service.AuctionService, n service.NotificationService, h *customWebSocket.Hub) *BidController {
	return &BidController{
		bidService:          b,
		auctionService:      a,
		notificationService: n,
		hub:                 h,
	}
}

// Struct nhận JSON body khi User đặt giá
type PlaceBidRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// PlaceBid godoc
// @Summary      Đặt giá
// @Description  Người dùng tham gia đấu giá bằng cách đặt mức giá mới cao hơn
// @Tags         auctions
// @Accept       json
// @Produce      json
// @Param        id       path      int              true  "Auction ID"
// @Param        request  body      PlaceBidRequest  true  "Số tiền đặt giá"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      403      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /auctions/{id}/bids [post]
func (c *BidController) PlaceBid(ctx *gin.Context) {
	// 1. Lấy AuctionID từ URL params `/auctions/:id/bids`
	auctionID, err := utils.GetIDFromParam(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 2. Lấy UserID người đang request từ token (JWT middleware)
	userIDVal, exists := ctx.Get("userID")
	if !exists {
		response.Error(ctx, apperror.NewUnauthorized(nil, "Không tìm thấy user_id trong token"))
		return
	}
	userID := userIDVal.(uint)

	// Admin cannot bid
	userRole, _ := ctx.Get("userRole")
	if userRole == "admin" {
		response.Error(ctx, apperror.NewForbidden(nil, "Admin không được phép tham gia đặt giá"))
		return
	}

	// 3. Parse JSON Body
	var req PlaceBidRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, apperror.NewBadRequest(err, "Dữ liệu đầu vào không hợp lệ"))
		return
	}

	// 4. Lấy thông tin auction TRƯỚC khi bid để biết ai là người đang thắng
	oldAuction, _ := c.auctionService.GetAuctionDetail(int(auctionID))

	// 5. Gọi Service thực thi Logic DB
	if err := c.bidService.PlaceBid(uint(auctionID), uint(userID), req.Amount); err != nil {
		response.Error(ctx, err)
		return
	}

	// 6. Nếu thành công -> Invalidate cache & Broadcast data mới tới Websocket Hub
	_ = c.auctionService.InvalidateCache(int(auctionID))
	latestAuction, err := c.auctionService.GetAuctionDetail(int(auctionID))
	if err == nil {
		bidderName := ""
		if len(latestAuction.Bids) > 0 {
			bidderName = latestAuction.Bids[0].User.FullName
		}

		// Gửi thông báo Outbid cho người cũ
		if oldAuction != nil && oldAuction.WinnerID != nil && *oldAuction.WinnerID != userID {
			title := "Bạn đã bị vượt mặt!"
			content := "Ai đó đã đặt giá cao hơn bạn trong phiên đấu giá " + oldAuction.Product.Name
			link := "/frontend/auction-detail.html?id=" + fmt.Sprintf("%d", uint(auctionID))
			_ = c.notificationService.NotifyUser(*oldAuction.WinnerID, title, content, "outbid", link)
		}

		msg := customWebSocket.Message{
			AuctionID: uint(auctionID),
			Payload: gin.H{
				"type":          "bid_update",
				"current_price": latestAuction.CurrentPrice,
				"bid_count":     latestAuction.BidCount,
				"bidder_name":   bidderName, // Người vừa bid xong (nếu có dữ liệu bids)
				"auction":       latestAuction,
			},
		}
		c.hub.Broadcast <- msg
	}

	// Trả về response Http thông thường trước báo hiệu đã Bid thành công (201 Created)
	response.Success(ctx, http.StatusCreated, gin.H{"message": "Đặt giá thành công"})
}

// GetMyBids godoc
// @Summary      Lịch sử đặt giá của tôi
// @Description  Lấy danh sách các lượt đặt giá mà người dùng hiện tại đã thực hiện
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse
// @Failure      401  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /users/me/bids [get]
func (c *BidController) GetMyBids(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		response.Error(ctx, apperror.NewUnauthorized(nil, "Không tìm thấy user_id trong token"))
		return
	}
	bids, err := c.bidService.GetMyBids(userID.(uint))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, bids)
}

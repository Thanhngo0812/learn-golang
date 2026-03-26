package controller

import (
	"golang-app/internal/model/entity"
	"golang-app/internal/service"
	"golang-app/pkg/apperror"
	"golang-app/pkg/response"
	"golang-app/pkg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuctionController struct{
	service service.AuctionService
}

func NewAuctionController(s service.AuctionService) *AuctionController {
	return &AuctionController{service: s}
}

// GetAuctionByID godoc
// @Summary      Chi tiết phiên đấu giá
// @Description  Lấy thông tin chi tiết một phiên đấu giá theo ID
// @Tags         auctions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Auction ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /auctions/{id} [get]
func (c *AuctionController) GetAuctionByID(ctx *gin.Context) {
	id, err := utils.GetIDFromParam(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	auction, err := c.service.GetAuctionDetail(id)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, auction)
}


// GetHotAuctions xử lý API Request để lấy danh sách Sản phẩm hot
// GetHotAuctions godoc
// @Summary      Danh sách phiên đấu giá HOT
// @Description  Lấy danh sách các phiên đấu giá đang hot (nhiều người quan tâm)
// @Tags         auctions
// @Accept       json
// @Produce      json
// @Param        limit  query     int  false  "Số lượng kết quả (mặc định 20)"
// @Success      200    {object}  response.SuccessResponse
// @Failure      400    {object}  response.ErrorResponse
// @Router       /auctions/hot [get]
func (c *AuctionController) GetHotAuctions(ctx *gin.Context) {
	limit := 20 // Cứng 20 sản phẩm theo yêu cầu đề tài hoặc lấy từ query params

	auctions, err := c.service.GetHotAuctions(limit)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, auctions)
}

// GetMyWonAuctions lấy danh sách phiên đấu giá đã thắng
// GetMyWonAuctions godoc
// @Summary      Danh sách phiên đấu giá đã thắng
// @Description  Lấy danh sách các phiên đấu giá mà người dùng hiện tại đã thắng
// @Tags         auctions
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse
// @Failure      401  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /users/me/auctions/won [get]
func (c *AuctionController) GetMyWonAuctions(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		response.Error(ctx, apperror.NewUnauthorized(nil, "Không tìm thấy user_id trong token"))
		return
	}
	auctions, err := c.service.GetWonAuctions(userID.(uint))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, auctions)
}

// GetAuctions lấy tất cả danh sách phiên đấu giá có lọc và phân trang
// GetAuctions godoc
// @Summary      Danh sách phiên đấu giá (có lọc và phân trang)
// @Description  Lấy danh sách phiên đấu giá dựa trên các tiêu chí lọc
// @Tags         auctions
// @Accept       json
// @Produce      json
// @Param        status      query     string  false  "Trạng thái (active, ended...)"
// @Param        product     query     string  false  "Tên sản phẩm"
// @Param        seller      query     string  false  "Tên người bán"
// @Param        category_id query     int     false  "ID danh mục"
// @Param        page        query     int     false  "Số trang (mặc định 1)"
// @Param        limit       query     int     false  "Số bản ghi mỗi trang (mặc định 12)"
// @Success      200         {object}  response.SuccessResponse
// @Failure      400         {object}  response.ErrorResponse
// @Router       /auctions [get]
func (c *AuctionController) GetAuctions(ctx *gin.Context) {
	status := ctx.Query("status")
	product := ctx.Query("product")
	seller := ctx.Query("seller")
	sellerID, _ := utils.GetIntQuery(ctx, "seller_id", 0)
	categoriesStr := ctx.Query("categories")
	categoryIDs := utils.ParseIntSlice(categoriesStr)
	page, _ := utils.GetIntQuery(ctx, "page", 1)
	limit, _ := utils.GetIntQuery(ctx, "limit", 12)

	auctions, total, err := c.service.GetAuctions(status, product, seller, uint(sellerID), categoryIDs, page, limit)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, gin.H{
		"auctions":    auctions,
		"total_count": total,
		"page":        page,
		"limit":       limit,
	})
}

// CreateAuction godoc
// @Summary      Tạo phiên đấu giá mới
// @Description  Seller tạo một phiên đấu giá cho sản phẩm của mình
// @Tags         auctions
// @Accept       json
// @Produce      json
// @Param        request  body      entity.Auction  true  "Thông tin phiên đấu giá"
// @Success      201      {object}  response.SuccessResponse{data=entity.Auction}
// @Failure      400      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /auctions [post]
func (c *AuctionController) CreateAuction(ctx *gin.Context) {
	var auction entity.Auction
	if err := ctx.ShouldBindJSON(&auction); err != nil {
		response.Error(ctx, err)
		return
	}

	if err := c.service.CreateAuction(&auction); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusCreated, auction)
}

// ExtendAuction godoc
// @Summary      Gia hạn phiên đấu giá
// @Description  Seller gia hạn thời gian kết thúc của phiên đấu giá
// @Tags         auctions
// @Accept       json
// @Produce      json
// @Param        id       path      int  true  "Auction ID"
// @Param        request  body      object  true  "Thời gian kết thúc mới"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /auctions/{id}/extend [patch]
func (c *AuctionController) ExtendAuction(ctx *gin.Context) {
	id, _ := utils.GetIDFromParam(ctx)
	var req struct {
		NewEndTime time.Time `json:"new_end_time" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, err)
		return
	}

	if err := c.service.ExtendAuction(uint(id), req.NewEndTime); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, gin.H{"message": "Đã gia hạn phiên đấu giá"})
}

// ConfirmDelivery godoc
// @Summary      Xác nhận đã gửi hàng
// @Description  Seller xác nhận đã gửi hàng cho người thắng
// @Tags         auctions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Auction ID"
// @Success      200  {object}  response.SuccessResponse{data=dto.SimpleMessageResponse}
// @Failure      400  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /auctions/{id}/confirm-delivery [patch]
func (c *AuctionController) ConfirmDelivery(ctx *gin.Context) {
	id, _ := utils.GetIDFromParam(ctx)
	userID := ctx.MustGet("userID").(uint)

	if err := c.service.ConfirmAction(uint(id), userID, true); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, gin.H{"message": "Đã xác nhận gửi hàng"})
}

// ConfirmReceipt godoc
// @Summary      Xác nhận đã nhận hàng
// @Description  Người thắng xác nhận đã nhận được hàng
// @Tags         auctions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Auction ID"
// @Success      200  {object}  response.SuccessResponse{data=dto.SimpleMessageResponse}
// @Failure      400  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /auctions/{id}/confirm-receipt [patch]
func (c *AuctionController) ConfirmReceipt(ctx *gin.Context) {
	id, _ := utils.GetIDFromParam(ctx)
	userID := ctx.MustGet("userID").(uint)

	if err := c.service.ConfirmAction(uint(id), userID, false); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, gin.H{"message": "Đã xác nhận nhận hàng"})
}

// RejectAuction godoc
// @Summary      Hủy hoặc từ chối phiên đấu giá
// @Description  Admin hoặc Seller hủy phiên đấu giá vì lý do cụ thể
// @Tags         auctions
// @Accept       json
// @Produce      json
// @Param        id       path      int     true  "Auction ID"
// @Param        request  body      object  true  "Lý do hủy"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /auctions/{id}/reject [patch]
func (ctrl *AuctionController) RejectAuction(c *gin.Context) {
	id, _ := utils.GetIDFromParam(c)
	userID := c.MustGet("userID").(uint)
	role := c.MustGet("userRole").(string)
	isAdmin := role == "admin"

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	if err := ctrl.service.RejectAuction(uint(id), userID, isAdmin, req.Reason); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Đã hủy phiên đấu giá"})
}

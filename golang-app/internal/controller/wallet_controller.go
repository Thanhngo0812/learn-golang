package controller

import (
	"golang-app/internal/model/dto"
	"golang-app/internal/service"
	"golang-app/pkg/apperror"
	"golang-app/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WalletController struct {
	service service.WalletService
}

func NewWalletController(s service.WalletService) *WalletController {
	return &WalletController{service: s}
}

// Deposit godoc
// @Summary      Nạp tiền vào ví
// @Description  Người dùng nạp tiền vào ví cá nhân
// @Tags         wallet
// @Accept       json
// @Produce      json
// @Param        request  body      dto.WalletRequest  true  "Số tiền nạp"
// @Success      200      {object}  response.SuccessResponse{data=entity.Wallet}
// @Failure      400      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /wallet/deposit [post]
func (c *WalletController) Deposit(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		response.Error(ctx, apperror.NewUnauthorized(nil, "Không tìm thấy user_id trong token"))
		return
	}

	var req dto.WalletRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, apperror.NewBadRequest(err, "Dữ liệu đầu vào không hợp lệ"))
		return
	}

	wallet, err := c.service.Deposit(userID.(uint), req.Amount)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, wallet)
}

// Withdraw godoc
// @Summary      Rút tiền từ ví
// @Description  Người dùng rút tiền từ ví cá nhân
// @Tags         wallet
// @Accept       json
// @Produce      json
// @Param        request  body      dto.WalletRequest  true  "Số tiền rút"
// @Success      200      {object}  response.SuccessResponse{data=entity.Wallet}
// @Failure      400      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /wallet/withdraw [post]
func (c *WalletController) Withdraw(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		response.Error(ctx, apperror.NewUnauthorized(nil, "Không tìm thấy user_id trong token"))
		return
	}

	var req dto.WalletRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, apperror.NewBadRequest(err, "Dữ liệu đầu vào không hợp lệ"))
		return
	}

	wallet, err := c.service.Withdraw(userID.(uint), req.Amount)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, wallet)
}

// GetTransactions godoc
// @Summary      Lịch sử giao dịch
// @Description  Lấy danh sách các giao dịch nạp/rút/tru tiền của người dùng
// @Tags         wallet
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse{data=[]entity.Transaction}
// @Failure      401  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /wallet/transactions [get]
func (c *WalletController) GetTransactions(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		response.Error(ctx, apperror.NewUnauthorized(nil, "Không tìm thấy user_id trong token"))
		return
	}

	transactions, err := c.service.GetTransactions(userID.(uint))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, transactions)
}

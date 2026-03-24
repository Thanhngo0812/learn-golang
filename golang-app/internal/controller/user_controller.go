package controller

import (
	"golang-app/internal/model/dto"
	"golang-app/internal/service"
	"golang-app/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service service.UserService
}

func NewUserController(s service.UserService) *UserController {
	return &UserController{service: s}
}

// Register godoc
// @Summary      Đăng ký người dùng (Admin)
// @Description  Admin tạo tài khoản người dùng mới
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RegisterRequest  true  "Thông tin đăng ký"
// @Success      201      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /users [post]
func (c *UserController) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, err)
		return
	}
	user, err := c.service.Register(&req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusCreated, user)
}

// PublicRegister godoc
// @Summary      Đăng ký tài khoản public
// @Description  Người dùng tự đăng ký tài khoản mới
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RegisterRequest  true  "Thông tin đăng ký"
// @Success      201      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Router       /auth/register [post]
func (c *UserController) PublicRegister(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, err)
		return
	}
	user, err := c.service.PublicRegister(&req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusCreated, user)
}

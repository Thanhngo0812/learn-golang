package controller

import (
	"golang-app/internal/config"
	"golang-app/internal/model/dto"
	"golang-app/internal/service"
	"golang-app/pkg/apperror"
	"golang-app/pkg/response"
	"golang-app/pkg/utils"
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

// GetMe xem thông tin cá nhân
// GetMe godoc
// @Summary      Xem thông tin cá nhân
// @Description  Lấy thông tin chi tiết của người dùng đang đăng nhập
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200      {object}  response.SuccessResponse
// @Failure      401      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /users/me [get]
func (c *UserController) GetMe(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		response.Error(ctx, apperror.NewUnauthorized(nil, "Không tìm thấy user_id trong token"))
		return
	}
	user, err := c.service.GetMyProfile(userID.(uint))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, user)
}

// UpdateMe cập nhật thông tin cá nhân
// UpdateMe godoc
// @Summary      Cập nhật thông tin cá nhân
// @Description  Người dùng tự cập nhật thông tin của mình
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      dto.UpdateUserRequest  true  "Thông tin cập nhật"
// @Success      200      {object}  response.SuccessResponse{data=entity.User}
// @Failure      400      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /users/me [put]
func (c *UserController) UpdateMe(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		response.Error(ctx, apperror.NewUnauthorized(nil, "Không tìm thấy user_id trong token"))
		return
	}
	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, err)
		return
	}
	user, err := c.service.UpdateMyProfile(userID.(uint), &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, user)
}

// Login godoc
// @Summary      Đăng nhập
// @Description  Xác thực người dùng và trả về token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.LoginRequest  true  "Thông tin đăng nhập"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      401      {object}  response.ErrorResponse
// @Router       /auth/login [post]
func (c *UserController) Login(ctx *gin.Context, cfg *config.Config) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, err)
		return
	}
	res, err := c.service.Login(&req, cfg)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, res)
}

// GetAllUsers godoc
// @Summary      Danh sách người dùng
// @Description  Lấy toàn bộ danh sách người dùng (Admin chỉ xem)
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200      {object}  response.SuccessResponse
// @Failure      401      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /users [get]
func (c *UserController) GetAllUsers(ctx *gin.Context) {
	users, err := c.service.GetListUsers()
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, users)
}

// GetUserByID godoc
// @Summary      Chi tiết người dùng
// @Description  Lấy thông tin chi tiết một người dùng theo ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      404  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /users/{id} [get]
func (c *UserController) GetUserByID(ctx *gin.Context) {
	id, err := utils.GetIDFromParam(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	user, err := c.service.GetUserDetail(id)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, user)
}

// UpdateUser godoc
// @Summary      Cập nhật người dùng (Admin)
// @Description  Admin cập nhật thông tin bất kỳ người dùng nào
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id       path      int                    true  "User ID"
// @Param        request  body      dto.UpdateUserRequest  true  "Thông tin cập nhật"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /users/{id} [put]
func (c *UserController) UpdateUser(ctx *gin.Context) {
	id, err := utils.GetIDFromParam(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, err)
		return
	}
	user, err := c.service.UpdateUser(id, &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, user)
}

// DeleteUser godoc
// @Summary      Xóa người dùng
// @Description  Xóa tài khoản người dùng khỏi hệ thống
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /users/{id} [delete]
func (c *UserController) DeleteUser(ctx *gin.Context) {
	id, err := utils.GetIDFromParam(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	err = c.service.DeleteUser(id)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, "Đã xóa người dùng thành công")
}

// LockUser khóa tài khoản (Admin only)
// LockUser godoc
// @Summary      Khóa tài khoản
// @Description  Admin khóa tài khoản người dùng
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /users/{id}/lock [patch]
func (c *UserController) LockUser(ctx *gin.Context) {
	id, err := utils.GetIDFromParam(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	user, err := c.service.LockUser(id)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, user)
}

// UnlockUser mở khóa tài khoản (Admin only)
// UnlockUser godoc
// @Summary      Mở khóa tài khoản
// @Description  Admin mở khóa tài khoản người dùng
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /users/{id}/unlock [patch]
func (c *UserController) UnlockUser(ctx *gin.Context) {
	id, err := utils.GetIDFromParam(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	user, err := c.service.UnlockUser(id)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, user)
}

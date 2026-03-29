package controller

import (
	"net/http"
	"strconv"
	"golang-app/internal/model/dto"
	"golang-app/internal/service"
	"golang-app/pkg/response"

	"github.com/gin-gonic/gin"
)

type CategoryController struct {
	service service.CategoryService
}

func NewCategoryController(service service.CategoryService) *CategoryController {
	return &CategoryController{service: service}
}

// CreateCategory godoc
// @Summary      Tạo danh mục mới
// @Description  Admin hoặc Seller tạo danh mục sản phẩm mới
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateCategoryRequest  true  "Thông tin danh mục"
// @Success      201      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /categories [post]
func (ctrl *CategoryController) CreateCategory(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	userID := c.MustGet("userID").(uint)
	role := c.MustGet("userRole").(string)

	category, err := ctrl.service.CreateCategory(req, userID, role)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusCreated, category)
}

// GetAllCategories godoc
// @Summary      Danh sách danh mục
// @Description  Lấy tất cả danh mục sản phẩm (có thể lọc theo trạng thái)
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        status  query     string  false  "Trạng thái (active, pending...)"
// @Success      200     {object}  response.SuccessResponse
// @Failure      400     {object}  response.ErrorResponse
// @Router       /categories [get]
func (ctrl *CategoryController) GetAllCategories(c *gin.Context) {
	status := c.Query("status")
	categories, err := ctrl.service.GetAllCategories(status)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, categories)
}

// ApproveCategory godoc
// @Summary      Duyệt danh mục
// @Description  Admin duyệt hoặc từ chối danh mục do Seller tạo
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id       path      int     true  "Category ID"
// @Param        request  body      object  true  "Thông tin duyệt"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /categories/{id}/status [patch]
func (ctrl *CategoryController) ApproveCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)

	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	err := ctrl.service.ApproveCategory(uint(id), req.Status, req.Reason)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Category status updated"})
}

// GetMyCategories godoc
// @Summary      Danh sách danh mục của tôi
// @Description  Seller lấy danh sách danh mục do mình tạo
// @Tags         categories
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse
// @Failure      401  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /categories/me [get]
func (ctrl *CategoryController) GetMyCategories(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	categories, err := ctrl.service.GetMyCategories(userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, categories)
}

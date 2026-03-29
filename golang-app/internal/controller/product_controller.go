package controller

import (
	"net/http"
	"strconv"
	"golang-app/internal/model/dto"
	"golang-app/internal/service"
	"golang-app/pkg/response"
	"golang-app/pkg/utils"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	service service.ProductService
}

func NewProductController(service service.ProductService) *ProductController {
	return &ProductController{service: service}
}

// CreateProduct godoc
// @Summary      Tạo sản phẩm mới
// @Description  Seller tạo sản phẩm mới để chuẩn bị đấu giá
// @Tags         products
// @Accept       multipart/form-data
// @Produce      json
// @Param        name         formData  string  true   "Tên sản phẩm"
// @Param        description  formData  string  true   "Mô tả sản phẩm"
// @Param        category_id  formData  int     true   "ID danh mục"
// @Param        images       formData  file    false  "Ảnh sản phẩm"
// @Success      201          {object}  response.SuccessResponse
// @Failure      400          {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /products [post]
func (ctrl *ProductController) CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err)
		return
	}

	// Handle file upload manually because gin binding for []*multipart.FileHeader in struct is tricky
	form, _ := c.MultipartForm()
	if form != nil {
		files := form.File["images"]
		req.Images = files
	}

	sellerID := c.MustGet("userID").(uint)

	product, err := ctrl.service.CreateProduct(req, sellerID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusCreated, product)
}

// GetProductByID godoc
// @Summary      Chi tiết sản phẩm
// @Description  Lấy thông tin chi tiết một sản phẩm theo ID
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /products/{id} [get]
func (ctrl *ProductController) GetProductByID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)

	product, err := ctrl.service.GetProductByID(uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, product)
}

// GetProducts godoc
// @Summary      Danh sách sản phẩm
// @Description  Lấy danh sách sản phẩm có lọc và phân trang
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        seller_id    query     int     false  "ID người bán"
// @Param        category_id  query     int     false  "ID danh mục"
// @Param        status       query     string  false  "Trạng thái"
// @Param        page         query     int     false  "Trang"
// @Param        limit        query     int     false  "Số lượng"
// @Success      200          {object}  response.SuccessResponse
// @Failure      400          {object}  response.ErrorResponse
// @Router       /products [get]
func (ctrl *ProductController) GetProducts(c *gin.Context) {
	sellerID, _ := utils.GetUintQuery(c, "seller_id", 0)
	categoryID, _ := utils.GetUintQuery(c, "category_id", 0)
	status := c.Query("status")
	page, _ := utils.GetIntQuery(c, "page", 1)
	limit, _ := utils.GetIntQuery(c, "limit", 12)

	products, total, err := ctrl.service.GetAllProducts(uint(sellerID), uint(categoryID), status, page, limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"products":    products,
		"total_count": total,
		"page":        page,
		"limit":       limit,
	})
}

// ApproveProduct godoc
// @Summary      Duyệt sản phẩm
// @Description  Admin duyệt hoặc từ chối sản phẩm
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      int     true  "Product ID"
// @Param        request  body      object  true  "Thông tin duyệt"
// @Success      200      {object}  response.SuccessResponse{data=dto.SimpleMessageResponse}
// @Failure      400      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /products/{id}/status [patch]
func (ctrl *ProductController) ApproveProduct(c *gin.Context) {
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

	if err := ctrl.service.ApproveProduct(uint(id), req.Status, req.Reason); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Đã cập nhật trạng thái sản phẩm"})
}

// UpdateProduct godoc
// @Summary      Cập nhật sản phẩm
// @Description  Seller cập nhật thông tin sản phẩm của mình
// @Tags         products
// @Accept       multipart/form-data
// @Produce      json
// @Param        id           path      int     true   "Product ID"
// @Param        name         formData  string  false  "Tên sản phẩm"
// @Param        description  formData  string  false  "Mô tả sản phẩm"
// @Param        category_id  formData  int     false  "ID danh mục"
// @Param        images       formData  file    false  "Ảnh sản phẩm thêm mới"
// @Success      200          {object}  response.SuccessResponse
// @Failure      400          {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /products/{id} [put]
func (ctrl *ProductController) UpdateProduct(c *gin.Context) {
	id, _ := utils.GetIDFromParam(c)
	var req dto.CreateProductRequest
	
	// Handle both JSON and Multipart
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err)
		return
	}

	// Handle file upload manually
	form, _ := c.MultipartForm()
	if form != nil {
		files := form.File["images"]
		req.Images = files
	}

	userID := c.MustGet("userID").(uint)
	role := c.MustGet("userRole").(string)
	isAdmin := role == "admin"

	if err := ctrl.service.UpdateProduct(uint(id), req, userID, isAdmin); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Cập nhật sản phẩm thành công"})
}

// DeleteProduct godoc
// @Summary      Xóa sản phẩm
// @Description  Seller hoặc Admin xóa sản phẩm
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /products/{id} [delete]
func (ctrl *ProductController) DeleteProduct(c *gin.Context) {
	id, _ := utils.GetIDFromParam(c)
	userID := c.MustGet("userID").(uint)
	role := c.MustGet("userRole").(string)
	isAdmin := role == "admin"

	if err := ctrl.service.DeleteProduct(uint(id), userID, isAdmin); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Xóa sản phẩm thành công"})
}

// ToggleBanned godoc
// @Summary      Khóa/Mở khóa sản phẩm
// @Description  Admin khóa hoặc mở khóa sản phẩm vì lý do vi phạm
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      int     true  "Product ID"
// @Param        request  body      object  true  "Lý do"
// @Success      200      {object}  response.SuccessResponse{data=dto.SimpleMessageResponse}
// @Failure      400      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /products/{id}/lock [patch]
func (ctrl *ProductController) ToggleBanned(c *gin.Context) {
	id, _ := utils.GetIDFromParam(c)
	role := c.MustGet("userRole").(string)
	isAdmin := role == "admin"

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	if err := ctrl.service.ToggleBanned(uint(id), isAdmin, req.Reason); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Đã cập nhật trạng thái khóa sản phẩm"})
}

// DeleteProductImage godoc
// @Summary      Xóa ảnh sản phẩm
// @Description  Xóa một ảnh cụ thể của sản phẩm
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      int  true  "Product ID"
// @Param        imageID  path      int  true  "Image ID"
// @Success      200      {object}  response.SuccessResponse{data=dto.SimpleMessageResponse}
// @Failure      400      {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /products/{id}/images/{imageID} [delete]
func (ctrl *ProductController) DeleteProductImage(c *gin.Context) {
	productID, _ := utils.GetIDFromParam(c)
	imageIDStr := c.Param("imageID")
	imageID, _ := strconv.ParseUint(imageIDStr, 10, 64)

	userID := c.MustGet("userID").(uint)
	role := c.MustGet("userRole").(string)
	isAdmin := role == "admin"

	if err := ctrl.service.DeleteProductImage(uint(productID), uint(imageID), userID, isAdmin); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Xóa ảnh thành công"})
}

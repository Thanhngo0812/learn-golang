package config

import (
	"database/sql"
	"golang-app/internal/controller"
	"golang-app/internal/repository"
	"golang-app/internal/service"
)

// Server chứa tất cả các dependencies của dự án
type Server struct {
	UserController *controller.UserController
	// Sau này thêm ProductController, OrderController vào đây
}

// NewServer chịu trách nhiệm khởi tạo toàn bộ dây chuyền
func NewServer(db *sql.DB) *Server {
	// 1. Init Repositories
	userRepo := repository.NewUserRepository(db)
	// productRepo := repository.NewProductRepository(db)

	// 2. Init Services
	userService := service.NewUserService(userRepo)
	// productService := service.NewProductService(productRepo)

	// 3. Init Controllers
	userController := controller.NewUserController(userService)
	// productController := controller.NewProductController(productService)

	return &Server{
		UserController: userController,
	}
}
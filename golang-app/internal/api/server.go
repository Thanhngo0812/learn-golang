package api

import (
	"log"
	"gorm.io/gorm"
	"golang-app/internal/config"
	"golang-app/internal/controller"
	"golang-app/internal/cron"
	"golang-app/internal/repository"
	"golang-app/internal/service"
	customWebSocket "golang-app/pkg/websocket"
	customRedis "golang-app/pkg/db"
	customCloudinary "golang-app/pkg/cloudinary"
	"golang-app/internal/worker"
	"time"
)

// Server chứa tất cả các dependencies của dự án
type Server struct {
	Config            *config.Config
	UserController    *controller.UserController
	AuctionController *controller.AuctionController
	BidController     *controller.BidController
	WalletController  *controller.WalletController
	WSController      *controller.WSController
	CategoryController *controller.CategoryController
	ProductController  *controller.ProductController
	NotificationController *controller.NotificationController
	WatchlistController    *controller.WatchlistController
	Hub               *customWebSocket.Hub
}

// NewServer chịu trách nhiệm khởi tạo toàn bộ dây chuyền
func NewServer(db *gorm.DB, cfg *config.Config) *Server {
	// 0.1 Khởi tạo Redis
	rdb, err := customRedis.InitRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// 0.2 Khởi tạo WebSocket Hub
	hub := customWebSocket.NewHub()
	go hub.Run()

	// 1. Init Repositories
	userRepo := repository.NewUserRepository(db)
	auctionRepo := repository.NewAuctionRepository(db)
	bidRepo := repository.NewBidRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	watchlistRepo := repository.NewWatchlistRepository(db)

	// 2. Init Services
	userService := service.NewUserService(userRepo, bidRepo)
	auctionService := service.NewAuctionService(auctionRepo, walletRepo, rdb)
	bidService := service.NewBidService(bidRepo)
	walletService := service.NewWalletService(walletRepo)

	// 2.1 Init Cloudinary & Worker
	cldClient, err := customCloudinary.NewCloudinaryClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}
	uploadWorker := worker.NewUploadWorker(cldClient, productRepo)
	uploadWorker.Start(5) // Start 5 workers

	categoryService := service.NewCategoryService(categoryRepo)
	productService := service.NewProductService(productRepo, auctionRepo, uploadWorker)
	notificationService := service.NewNotificationService(notificationRepo, hub)
	watchlistService := service.NewWatchlistService(watchlistRepo)

	// 3. Init Controllers
	userController := controller.NewUserController(userService)
	auctionController := controller.NewAuctionController(auctionService)
	bidController := controller.NewBidController(bidService, auctionService, notificationService, hub)
	walletController := controller.NewWalletController(walletService)
	wsController := controller.NewWSController(hub)
	categoryController := controller.NewCategoryController(categoryService)
	productController := controller.NewProductController(productService)
	notificationController := controller.NewNotificationController(notificationService)
	watchlistController := controller.NewWatchlistController(watchlistService)

	// 4. Init Workers
	auctionWorker := cron.NewAuctionWorker(auctionRepo, notificationService)
	auctionWorker.Start(1 * time.Minute)

	return &Server{
		Config:            cfg,
		Hub:               hub,
		UserController:    userController,
		AuctionController: auctionController,
		BidController:     bidController,
		WalletController:  walletController,
		WSController:      wsController,
		CategoryController: categoryController,
		ProductController:  productController,
		NotificationController: notificationController,
		WatchlistController:    watchlistController,
	}
}
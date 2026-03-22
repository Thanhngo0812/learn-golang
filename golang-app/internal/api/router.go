package api 

import (
	"github.com/gin-gonic/gin"
	"golang-app/internal/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "golang-app/docs"
)

func NewRouter(server *Server) *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())   // Cho phép Frontend gọi API
	r.Use(middleware.Logger()) // Middleware Log custom

	api := r.Group("/api/v1")
	{
		// 1. PUBLIC ROUTES (Không cần token)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", func(ctx *gin.Context) {
				server.UserController.Login(ctx, server.Config)
			})
			authGroup.POST("/register", server.UserController.PublicRegister) // Public route for bidders/sellers
		}

		// 2. PROTECTED ROUTES - USERS (Cần Token)
		usersProtected := api.Group("/users")
		usersProtected.Use(middleware.RequireAuth(server.Config))
		{
			// --- Thông tin cá nhân (Bidder/Seller/Admin) ---
			usersProtected.GET("/me", server.UserController.GetMe)       // Xem thông tin cá nhân + ví
			usersProtected.PUT("/me", server.UserController.UpdateMe)    // Sửa thông tin cá nhân
			usersProtected.GET("/me/bids", server.BidController.GetMyBids)            // Lịch sử bid
			usersProtected.GET("/me/auctions/won", server.AuctionController.GetMyWonAuctions) // Đấu giá đã thắng

			// --- Quản lý Users (Admin Only) ---
			usersProtected.POST("", middleware.RequireRole("admin"), server.UserController.Register)
			usersProtected.GET("", server.UserController.GetAllUsers)
			usersProtected.GET("/:id", server.UserController.GetUserByID)
			usersProtected.PUT("/:id", server.UserController.UpdateUser)
			usersProtected.DELETE("/:id", server.UserController.DeleteUser)
			usersProtected.PATCH("/:id/lock", middleware.RequireRole("admin"), server.UserController.LockUser)
			usersProtected.PATCH("/:id/unlock", middleware.RequireRole("admin"), server.UserController.UnlockUser)
		}

		// 3. WALLET ROUTES (Cần Token - Nạp/Rút tiền)
		walletGroup := api.Group("/wallet")
		walletGroup.Use(middleware.RequireAuth(server.Config))
		{
			walletGroup.POST("/deposit", server.WalletController.Deposit)               // Nạp tiền
			walletGroup.POST("/withdraw", server.WalletController.Withdraw)             // Rút tiền
			walletGroup.GET("/transactions", server.WalletController.GetTransactions)   // Lịch sử giao dịch
		}

		// 4. AUCTION ROUTES (Rate Limited)
		auctions := api.Group("/auctions")
		{
			auctions.GET("", server.AuctionController.GetAuctions)
			auctions.GET("/hot", server.AuctionController.GetHotAuctions)
			auctions.GET("/:id", middleware.ConcurrencyLimiter(200), server.AuctionController.GetAuctionByID)

			// Protected Auction Actions
			authAuctions := auctions.Group("")
			authAuctions.Use(middleware.RequireAuth(server.Config))
			{
				authAuctions.POST("/:id/bids", middleware.RequireRole("bidder"), server.BidController.PlaceBid)
				authAuctions.POST("", middleware.RequireRole("seller"), server.AuctionController.CreateAuction)
				authAuctions.PATCH("/:id/extend", middleware.RequireRole("seller"), server.AuctionController.ExtendAuction)
				authAuctions.PATCH("/:id/confirm-delivery", middleware.RequireRole("seller"), server.AuctionController.ConfirmDelivery)
				authAuctions.PATCH("/:id/confirm-receipt", middleware.RequireRole("bidder"), server.AuctionController.ConfirmReceipt)
				authAuctions.PATCH("/:id/reject", server.AuctionController.RejectAuction)
			}
		}

		// 6. CATEGORY ROUTES
		categories := api.Group("/categories")
		{
			categories.GET("", server.CategoryController.GetAllCategories)
			
			// Protected routes for creation
			categories.Use(middleware.RequireAuth(server.Config))
			categories.GET("/me", server.CategoryController.GetMyCategories)
			categories.POST("", middleware.RequireRole("admin", "seller"), server.CategoryController.CreateCategory)
			categories.PATCH("/:id/status", middleware.RequireRole("admin"), server.CategoryController.ApproveCategory)
		}

		// 7. PRODUCT ROUTES
		products := api.Group("/products")
		{
			products.GET("", middleware.RequireAuth(server.Config), server.ProductController.GetProducts)
			products.GET("/:id", server.ProductController.GetProductByID)
			products.PUT("/:id", middleware.RequireRole("seller", "admin"), server.ProductController.UpdateProduct)
			products.DELETE("/:id", middleware.RequireRole("seller", "admin"), server.ProductController.DeleteProduct)
			products.DELETE("/:id/images/:imageID", middleware.RequireRole("seller", "admin"), server.ProductController.DeleteProductImage)
			products.PATCH("/:id/lock", middleware.RequireRole("admin"), server.ProductController.ToggleBanned)

			products.Use(middleware.RequireAuth(server.Config))
			products.POST("", middleware.RequireRole("seller"), server.ProductController.CreateProduct)
			products.PATCH("/:id/status", middleware.RequireRole("admin"), server.ProductController.ApproveProduct)
		}

		// 9. NOTIFICATION ROUTES
		notifications := api.Group("/notifications")
		notifications.Use(middleware.RequireAuth(server.Config))
		{
			notifications.GET("", server.NotificationController.GetMyNotifications)
			notifications.PATCH("/:id/read", server.NotificationController.MarkAsRead)
			notifications.PATCH("/read-all", server.NotificationController.MarkAllAsRead)
		}

		// 10. WATCHLIST ROUTES
		watchlist := api.Group("/watchlist")
		watchlist.Use(middleware.RequireAuth(server.Config))
		{
			watchlist.GET("", server.WatchlistController.GetMyWatchlist)
			watchlist.POST("/:id", server.WatchlistController.ToggleWatch)
			watchlist.GET("/:id/status", server.WatchlistController.CheckStatus)
		}

		// 8. WEBSOCKET ROUTES 
		ws := api.Group("/ws")
		{
			ws.GET("/auctions/:id", server.WSController.ServeWS)
			ws.GET("/notifications", middleware.RequireAuth(server.Config), server.WSController.ServeNotifications)
		}
	}

	// 6. SERVE FRONTEND (Static Files)
	frontend := r.Group("/frontend")
	frontend.Use(middleware.NoCache())
	{
		frontend.Static("", "./frontend")
	}

	// 7. SWAGGER DOCUMENTATION
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/", func(c *gin.Context) {
		c.Redirect(301, "/frontend/index.html")
	})

	return r
}
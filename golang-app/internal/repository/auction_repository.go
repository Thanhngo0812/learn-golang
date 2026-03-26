package repository

import (
	"time"
	"golang-app/internal/model/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuctionRepository interface {
	GetAuctionByID(id int) (*entity.Auction, error)
	GetHotAuctions(limit int) ([]entity.Auction, error)
	GetWonAuctions(userID uint) ([]entity.Auction, error)
	GetExpiredAuctions(now time.Time) ([]entity.Auction, error) // Added
	GetAuctions(status string, productName string, sellerName string, sellerID uint, categoryIDs []int, page int, limit int) ([]entity.Auction, int64, error)
	Create(auction *entity.Auction) error
	Update(auction *entity.Auction) error
	ConfirmAction(id uint, isSeller bool) error
	SettleAuctionPayment(auctionID uint, winnerID uint, sellerID uint, amount float64) error
	CheckActiveAuction(productID uint) (bool, error)
}

type auctionRepo struct {
	db *gorm.DB
}

func NewAuctionRepository(db *gorm.DB) AuctionRepository {
	return &auctionRepo{db: db}
}

func (r *auctionRepo) GetAuctionByID(id int) (*entity.Auction, error) {
	var auction entity.Auction
	err := r.db.Table("auctions").
		Select("auctions.*, (SELECT COUNT(*) FROM bids WHERE bids.auction_id = auctions.id) as bid_count").
		Preload("Product").
		Preload("Product.Images").
		Preload("Product.Seller").
		Preload("Bids", func(db *gorm.DB) *gorm.DB {
			return db.Order("bid_time DESC").Limit(50) // Lấy tối đa 50 lượt gần nhất
		}).
		Preload("Bids.User"). // Nạp người bid để hiện tên
		Where("auctions.id = ?", id).
		First(&auction).Error

	if err != nil {
		return nil, err
	}
	return &auction, nil
}

// GetHotAuctions kết hợp Kỹ thuật 2 (Select Field) & Kỹ thuật 3 (Eager Loading Product)
func (r *auctionRepo) GetHotAuctions(limit int) ([]entity.Auction, error) {
	var auctions []entity.Auction

	err := r.db.Table("auctions").
		// 1. SELECT FIELD: Chỉ lấy các trường hiển thị, loại bỏ những dữ liệu không mong muốn và Đếm Bids
		Select("auctions.id, auctions.product_id, auctions.start_price, auctions.current_price, auctions.end_time, auctions.status, COUNT(bids.id) as bid_count").
		// Join với bảng Bids để tạo dữ liệu COUNT
		Joins("LEFT JOIN bids ON bids.auction_id = auctions.id").
		// Group By theo auction.id kết hợp với Index bên structs Bid để tính nhanh nhất
		Group("auctions.id").
		// Sắp xếp giảm dần theo lượt Bid
		Order("bid_count DESC").
		Limit(limit).
		// 2. PRELOAD EAGER LOADING: Nạp sẵn Product gắn với Auction để giải quyết lỗi N+1 Query
		Preload("Product").
		Preload("Product.Images").
		Preload("Product.Seller").
		Preload("Bids", func(db *gorm.DB) *gorm.DB {
			return db.Order("bid_time DESC")
		}).
		Preload("Bids.User").
		Find(&auctions).Error

	return auctions, err
}


func (r *auctionRepo) GetAuctions(status string, productName string, sellerName string, sellerID uint, categoryIDs []int, page int, limit int) ([]entity.Auction, int64, error) {
	var auctions []entity.Auction
	var total int64

	// Base query for counting
	countQuery := r.db.Model(&entity.Auction{}).
		Joins("JOIN products ON products.id = auctions.product_id").
		Joins("JOIN users ON users.id = products.seller_id")

	if status != "" {
		countQuery = countQuery.Where("auctions.status = ?", status)
	}
	if productName != "" {
		countQuery = countQuery.Where("products.name ILIKE ?", "%"+productName+"%")
	}
	if sellerName != "" {
		countQuery = countQuery.Where("users.full_name ILIKE ?", "%"+sellerName+"%")
	}
	if sellerID != 0 {
		countQuery = countQuery.Where("products.seller_id = ?", sellerID)
	}
	if len(categoryIDs) > 0 {
		countQuery = countQuery.Where("products.category_id IN ?", categoryIDs)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Data query with Select and Joins for bid_count
	offset := (page - 1) * limit
	err := countQuery.
		Select("auctions.*, COUNT(bids.id) as bid_count").
		Joins("LEFT JOIN bids ON bids.auction_id = auctions.id").
		Group("auctions.id, products.id, users.id").
		Preload("Product").
		Preload("Product.Images").
		Preload("Product.Seller").
		Preload("Bids", func(db *gorm.DB) *gorm.DB {
			return db.Order("bid_time DESC") // Lấy các lượt bid (sắp xếp mới nhất lên đầu)
		}).
		Preload("Bids.User").
		Order("auctions.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&auctions).Error

	return auctions, total, err
}

func (r *auctionRepo) Create(auction *entity.Auction) error {
	return r.db.Create(auction).Error
}

func (r *auctionRepo) Update(auction *entity.Auction) error {
	return r.db.Omit(clause.Associations).Save(auction).Error
}

func (r *auctionRepo) GetWonAuctions(userID uint) ([]entity.Auction, error) {
	var auctions []entity.Auction
	err := r.db.Preload("Product").Preload("Product.Images").
		Where("winner_id = ? AND (status = 'ended' OR status = 'sold')", userID).
		Order("end_time DESC").Find(&auctions).Error
	return auctions, err
}

func (r *auctionRepo) GetExpiredAuctions(now time.Time) ([]entity.Auction, error) {
	var auctions []entity.Auction
	err := r.db.Preload("Product").Preload("Product.Seller").
		Where("status = 'active' AND end_time <= ?", now).
		Find(&auctions).Error
	return auctions, err
}

func (r *auctionRepo) ConfirmAction(id uint, isSeller bool) error {
	column := "buyer_confirmed"
	if isSeller {
		column = "seller_confirmed"
	}
	return r.db.Model(&entity.Auction{}).Where("id = ?", id).Update(column, true).Error
}

func (r *auctionRepo) SettleAuctionPayment(auctionID uint, winnerID uint, sellerID uint, amount float64) error {
	if err := validateWalletAmount(amount); err != nil {
		return err
	}
	if winnerID == sellerID {
		return ErrWalletTransferSameAccount
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		var auction entity.Auction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&auction, auctionID).Error; err != nil {
			return err
		}

		if auction.Status == "sold" {
			return nil
		}

		firstID, secondID := winnerID, sellerID
		if firstID > secondID {
			firstID, secondID = secondID, firstID
		}

		firstWallet, err := lockWallet(tx, firstID)
		if err != nil {
			return err
		}
		secondWallet, err := lockWallet(tx, secondID)
		if err != nil {
			return err
		}

		wallets := map[uint]*entity.Wallet{
			firstID:  firstWallet,
			secondID: secondWallet,
		}
		fromWallet := wallets[winnerID]
		toWallet := wallets[sellerID]

		if err := ensureWalletState(fromWallet); err != nil {
			return err
		}
		if err := ensureWalletState(toWallet); err != nil {
			return err
		}
		if fromWallet.FrozenBalance < amount || fromWallet.Balance < amount {
			return ErrInsufficientFrozenFunds
		}

		fromWallet.Balance -= amount
		fromWallet.FrozenBalance -= amount
		if err := tx.Save(fromWallet).Error; err != nil {
			return err
		}

		toWallet.Balance += amount
		if err := tx.Save(toWallet).Error; err != nil {
			return err
		}

		if err := createWalletTransaction(tx, winnerID, -amount, "payment", "Thanh toán cho phiên đấu giá"); err != nil {
			return err
		}

		if err := createWalletTransaction(tx, sellerID, amount, "deposit", "Nhận tiền từ phiên đấu giá"); err != nil {
			return err
		}

		return tx.Model(&entity.Auction{}).
			Where("id = ? AND status <> ?", auctionID, "sold").
			Update("status", "sold").Error
	})
}

func (r *auctionRepo) CheckActiveAuction(productID uint) (bool, error) {
	var count int64
	err := r.db.Model(&entity.Auction{}).
		Where("product_id = ? AND (status IN ('pending', 'active', 'sold') OR (status = 'ended' AND winner_id IS NOT NULL))", productID).
		Count(&count).Error
	return count > 0, err
}

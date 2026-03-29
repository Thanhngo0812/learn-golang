package repository

import (
	"golang-app/internal/model/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BidRepository interface {
	PlaceBid(bid *entity.Bid) error
	GetBidsByUserID(userID uint) ([]entity.Bid, error)
	HasUserBids(userID uint) (bool, error)
}

type bidRepo struct {
	db *gorm.DB
}

func NewBidRepository(db *gorm.DB) BidRepository {
	return &bidRepo{db: db}
}

// PlaceBid attempts to insert a new bid into the database.
// All business logic (price validation, sniper protection, anti self-bidding)
// is handled automatically by the PostgreSQL BEFORE INSERT trigger set up in init.sql.
func (r *bidRepo) PlaceBid(bid *entity.Bid) error {
	// Bắt đầu một Transaction ở Go
	return r.db.Transaction(func(tx *gorm.DB) error {
		var auction entity.Auction
		
		// 1. KHÓA BI QUAN (PESSIMISTIC LOCK) Ở TẦNG GOLANG GORM
		// SELECT * FROM auctions WHERE id = ? FOR UPDATE
		// Lệnh này bắt buộc các session khác phải chờ trên Row của Auction này
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&auction, bid.AuctionID).Error; err != nil {
			return err
		}

		// 2. Tiếp tục thực thi Insert tạo Bid. SQL Trigger sẽ tiếp diễn sau bước này.
		return tx.Create(bid).Error
	})
}

// GetBidsByUserID lấy lịch sử bid của user, kèm thông tin Auction
func (r *bidRepo) GetBidsByUserID(userID uint) ([]entity.Bid, error) {
	var bids []entity.Bid
	if err := r.db.Where("user_id = ?", userID).Preload("Auction").Order("bid_time DESC").Find(&bids).Error; err != nil {
		return nil, err
	}
	return bids, nil
}

// HasUserBids kiểm tra user đã tham gia bid chưa
func (r *bidRepo) HasUserBids(userID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.Bid{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

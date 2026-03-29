package repository

import (
	"golang-app/internal/model/entity"
	"gorm.io/gorm"
)

type WatchlistRepository interface {
	Add(userID uint, auctionID uint) error
	Remove(userID uint, auctionID uint) error
	IsWatching(userID uint, auctionID uint) (bool, error)
	GetWatchlistByUserID(userID uint) ([]entity.Watchlist, error)
}

type watchlistRepo struct {
	db *gorm.DB
}

func NewWatchlistRepository(db *gorm.DB) WatchlistRepository {
	return &watchlistRepo{db: db}
}

func (r *watchlistRepo) Add(userID uint, auctionID uint) error {
	w := entity.Watchlist{UserID: userID, AuctionID: auctionID}
	return r.db.Create(&w).Error
}

func (r *watchlistRepo) Remove(userID uint, auctionID uint) error {
	return r.db.Where("user_id = ? AND auction_id = ?", userID, auctionID).Delete(&entity.Watchlist{}).Error
}

func (r *watchlistRepo) IsWatching(userID uint, auctionID uint) (bool, error) {
	var count int64
	err := r.db.Model(&entity.Watchlist{}).Where("user_id = ? AND auction_id = ?", userID, auctionID).Count(&count).Error
	return count > 0, err
}

func (r *watchlistRepo) GetWatchlistByUserID(userID uint) ([]entity.Watchlist, error) {
	var watchlist []entity.Watchlist
	err := r.db.Preload("Auction").Preload("Auction.Product").Preload("Auction.Product.Images").
		Where("user_id = ?", userID).Find(&watchlist).Error
	return watchlist, err
}

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang-app/internal/model/entity"
	"golang-app/internal/repository"
	"golang-app/pkg/apperror"
	"gorm.io/gorm"
)

func mapWalletRepoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.NewNotFound(err, "Ví người dùng không tồn tại")
	}
	if errors.Is(err, repository.ErrInvalidWalletAmount) || errors.Is(err, repository.ErrInsufficientAvailableFunds) || errors.Is(err, repository.ErrInsufficientFrozenFunds) {
		return apperror.NewBadRequest(err, "Trạng thái ví không hợp lệ để thực hiện thao tác này")
	}
	if errors.Is(err, repository.ErrInvalidWalletState) || errors.Is(err, repository.ErrWalletTransferSameAccount) {
		return apperror.NewInternal(err)
	}
	return apperror.NewInternal(err)
}
type AuctionService interface {
	GetAuctionDetail(id int) (*entity.Auction, error)
	GetHotAuctions(limit int) ([]entity.Auction, error)
	GetWonAuctions(userID uint) ([]entity.Auction, error)
	GetAuctions(status string, productName string, sellerName string, sellerID uint, categoryIDs []int, page int, limit int) ([]entity.Auction, int64, error)
	CreateAuction(auction *entity.Auction) error
	ExtendAuction(id uint, newEndTime time.Time) error
	ConfirmAction(id uint, userID uint, isSeller bool) error
	RejectAuction(id uint, userID uint, isAdmin bool, reason string) error
	InvalidateCache(id int) error
}

type auctionService struct {
	repo       repository.AuctionRepository
	walletRepo repository.WalletRepository
	redis      *redis.Client
}

func NewAuctionService(repo repository.AuctionRepository, walletRepo repository.WalletRepository, redisClient *redis.Client) AuctionService {
	return &auctionService{
		repo:       repo,
		walletRepo: walletRepo,
		redis:      redisClient,
	}
}

func (s *auctionService) GetAuctionDetail(id int) (*entity.Auction, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("auction:%d", id)

	// Step 1: Query Redis (Cache-Aside: Cache Hit)
	cachedData, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var auction entity.Auction
		if err = json.Unmarshal([]byte(cachedData), &auction); err == nil {
			return &auction, nil
		}
	}

	// Step 2: Cache Miss (Query DB)
	// Tạm thời sleep 1000ms để GIỮ kết nối lâu một chút báo cạn Rate Limiting

	auction, err := s.repo.GetAuctionByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound(err, "Phiên đấu giá không tồn tại")
		}
		return nil, apperror.NewInternal(err)
	}

	// Step 3: Lưu vào Redis (Set TTL = 60s)
	if bytes, err := json.Marshal(auction); err == nil {
		s.redis.Set(ctx, cacheKey, bytes, 60*time.Second)
	}

	return auction, nil
}

func (s *auctionService) InvalidateCache(id int) error {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("auction:%d", id)
	return s.redis.Del(ctx, cacheKey).Err()
}


func (s *auctionService) GetHotAuctions(limit int) ([]entity.Auction, error) {
	if limit <= 0 {
		limit = 20 // Default limit
	}
	auctions, err := s.repo.GetHotAuctions(limit)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return auctions, nil
}

// GetWonAuctions lấy danh sách phiên đấu giá mà user đã thắng
func (s *auctionService) GetWonAuctions(userID uint) ([]entity.Auction, error) {
	auctions, err := s.repo.GetWonAuctions(userID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return auctions, nil
}

func (s *auctionService) GetAuctions(status string, productName string, sellerName string, sellerID uint, categoryIDs []int, page int, limit int) ([]entity.Auction, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 12 // Default 12 auctions per page
	}

	auctions, total, err := s.repo.GetAuctions(status, productName, sellerName, sellerID, categoryIDs, page, limit)
	if err != nil {
		return nil, 0, apperror.NewInternal(err)
	}
	return auctions, total, nil
}

func (s *auctionService) CreateAuction(auction *entity.Auction) error {
	active, err := s.repo.CheckActiveAuction(auction.ProductID)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if active {
		return apperror.NewBadRequest(nil, "Sản phẩm này hiện đang có một phiên đấu giá chưa kết thúc")
	}
	return s.repo.Create(auction)
}

func (s *auctionService) ExtendAuction(id uint, newEndTime time.Time) error {
	auction, err := s.repo.GetAuctionByID(int(id))
	if err != nil {
		return err
	}
	auction.EndTime = newEndTime
	return s.repo.Update(auction)
}

func (s *auctionService) ConfirmAction(id uint, userID uint, isSeller bool) error {
	auction, err := s.repo.GetAuctionByID(int(id))
	if err != nil {
		return err
	}

	// Validate roles
	if isSeller {
		if auction.Product.SellerID != userID {
			return apperror.NewForbidden(nil, "Bạn không phải người bán của sản phẩm này")
		}
	} else {
		if auction.WinnerID == nil || *auction.WinnerID != userID {
			return apperror.NewForbidden(nil, "Bạn không phải người thắng cuộc của phiên đấu giá này")
		}
	}

	if err := s.repo.ConfirmAction(id, isSeller); err != nil {
		return err
	}

	// Re-fetch to check both
	updated, err := s.repo.GetAuctionByID(int(id))
	if err != nil {
		return apperror.NewInternal(err)
	}

	if updated.SellerConfirmed && updated.BuyerConfirmed && updated.Status != "sold" {
		// Proceed to payment
		if updated.WinnerID != nil && updated.Product != nil && updated.CurrentPrice > 0 {
			err := s.repo.SettleAuctionPayment(updated.ID, *updated.WinnerID, updated.Product.SellerID, updated.CurrentPrice)
			if err != nil {
				return mapWalletRepoError(err)
			}
		}
	}

	return nil
}

func (s *auctionService) RejectAuction(id uint, userID uint, isAdmin bool, reason string) error {
	auction, err := s.repo.GetAuctionByID(int(id))
	if err != nil {
		return err
	}

	// Admin can reject (ban) anything. Seller can reject their own active/pending auction. Winner can reject for limited reasons (handled by caller logic if needed).
	if !isAdmin {
		if auction.Product.SellerID != userID && (auction.WinnerID == nil || *auction.WinnerID != userID) {
			return apperror.NewForbidden(nil, "Bạn không có quyền từ chối phiên này")
		}
	}

	auction.Status = "cancelled"
	if isAdmin {
		auction.Status = "banned"
	}
	auction.RejectionReason = reason
	
	if err := s.repo.Update(auction); err != nil {
		return err
	}

	// Release frozen money for the winner if exists
	if auction.WinnerID != nil && auction.CurrentPrice > 0 {
		return mapWalletRepoError(s.walletRepo.UnfreezeMoney(*auction.WinnerID, auction.CurrentPrice))
	}

	return nil
}

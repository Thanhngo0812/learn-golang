package service

import (
	"golang-app/internal/model/entity"
	"golang-app/internal/repository"
	"golang-app/pkg/apperror"
	"strings"
	"sync"
)

type BidService interface {
	PlaceBid(auctionID uint, userID uint, amount float64) error
	GetMyBids(userID uint) ([]entity.Bid, error)
}

type bidService struct {
	repo       repository.BidRepository
	globalMu   sync.Mutex
	auctionMus map[uint]*sync.Mutex
}

func NewBidService(repo repository.BidRepository) BidService {
	return &bidService{
		repo:       repo,
		auctionMus: make(map[uint]*sync.Mutex),
	}
}

// getAuctionMutex trả về một Mutex riêng biệt cho từng phiên đấu giá (AuctionID)
// Điều này giúp chặn tranh chấp các luồng đấu giá cùng lúc vào CÙNG 1 sản phẩm.
// Các phiên đấu giá khác nhau vẫn có thể chạy song song.
func (s *bidService) getAuctionMutex(auctionID uint) *sync.Mutex {
	s.globalMu.Lock()
	defer s.globalMu.Unlock()
	
	if s.auctionMus == nil {
		s.auctionMus = make(map[uint]*sync.Mutex)
	}
	if _, exists := s.auctionMus[auctionID]; !exists {
		s.auctionMus[auctionID] = &sync.Mutex{}
	}
	return s.auctionMus[auctionID]
}

func (s *bidService) PlaceBid(auctionID uint, userID uint, amount float64) error {
	// 1. CƠ CHẾ ĐỒNG BỘ GO: Khóa Mutex cho phiên đấu giá này
	mu := s.getAuctionMutex(auctionID)
	mu.Lock()
	defer mu.Unlock()

	bid := &entity.Bid{
		AuctionID: auctionID,
		UserID:    userID,
		Amount:    amount,
	}

	err := s.repo.PlaceBid(bid)
	if err != nil {
		// Bắt lỗi từ PostgreSQL Trigger (RAISE EXCEPTION)
		errMsg := err.Error()
		
		// Phân tích thông báo lỗi rớt ra từ DB trigger
		if strings.Contains(errMsg, "Phiên đấu giá chưa bắt đầu") ||
			strings.Contains(errMsg, "Phiên đấu giá đã kết thúc") ||
			strings.Contains(errMsg, "Giá đặt lần đầu tiên") ||
			strings.Contains(errMsg, "Giá đặt tiếp theo") ||
			strings.Contains(errMsg, "Số dư khả dụng") ||
			strings.Contains(errMsg, "tự đấu giá") {
			
			// Làm sạch error message (Bỏ đi phần râu ria của pgsql driver nếu có)
			parts := strings.Split(errMsg, "ERROR: ")
			cleanMsg := errMsg
			if len(parts) > 1 {
				cleanParts := strings.Split(parts[1], " (SQLSTATE")
				if len(cleanParts) > 0 {
					cleanMsg = cleanParts[0]
				}
			}
			return apperror.NewBadRequest(nil, cleanMsg)
		}

		return apperror.NewInternal(err)
	}

	return nil
}

// GetMyBids lấy lịch sử bid của user
func (s *bidService) GetMyBids(userID uint) ([]entity.Bid, error) {
	bids, err := s.repo.GetBidsByUserID(userID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return bids, nil
}

package service

import (
	"golang-app/internal/model/entity"
	"golang-app/internal/repository"
)

type WatchlistService interface {
	ToggleWatch(userID uint, auctionID uint) (bool, error)
	GetWatchlist(userID uint) ([]entity.Watchlist, error)
	IsWatching(userID uint, auctionID uint) (bool, error)
}

type watchlistService struct {
	repo repository.WatchlistRepository
}

func NewWatchlistService(repo repository.WatchlistRepository) WatchlistService {
	return &watchlistService{repo: repo}
}

func (s *watchlistService) ToggleWatch(userID uint, auctionID uint) (bool, error) {
	watching, err := s.repo.IsWatching(userID, auctionID)
	if err != nil {
		return false, err
	}

	if watching {
		if err := s.repo.Remove(userID, auctionID); err != nil {
			return false, err
		}
		return false, nil
	} else {
		if err := s.repo.Add(userID, auctionID); err != nil {
			return false, err
		}
		return true, nil
	}
}

func (s *watchlistService) GetWatchlist(userID uint) ([]entity.Watchlist, error) {
	return s.repo.GetWatchlistByUserID(userID)
}

func (s *watchlistService) IsWatching(userID uint, auctionID uint) (bool, error) {
	return s.repo.IsWatching(userID, auctionID)
}

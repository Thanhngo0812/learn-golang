package service

import (
	"errors"
	"golang-app/internal/model/entity"
	"golang-app/internal/repository"
	"golang-app/pkg/apperror"

	"gorm.io/gorm"
)

type WalletService interface {
	Deposit(userID uint, amount float64) (*entity.Wallet, error)
	Withdraw(userID uint, amount float64) (*entity.Wallet, error)
	GetTransactions(userID uint) ([]entity.Transaction, error)
}

type walletService struct {
	repo repository.WalletRepository
}

func NewWalletService(repo repository.WalletRepository) WalletService {
	return &walletService{repo: repo}
}

func (s *walletService) Deposit(userID uint, amount float64) (*entity.Wallet, error) {
	wallet, err := s.repo.Deposit(userID, amount)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidWalletAmount) {
			return nil, apperror.NewBadRequest(err, "Số tiền phải lớn hơn 0")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound(err, "Ví người dùng không tồn tại")
		}
		if errors.Is(err, repository.ErrInvalidWalletState) {
			return nil, apperror.NewInternal(err)
		}
		return nil, apperror.NewInternal(err)
	}
	return wallet, nil
}

func (s *walletService) Withdraw(userID uint, amount float64) (*entity.Wallet, error) {
	wallet, err := s.repo.Withdraw(userID, amount)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidWalletAmount) {
			return nil, apperror.NewBadRequest(err, "Số tiền phải lớn hơn 0")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound(err, "Ví người dùng không tồn tại")
		}
		if errors.Is(err, repository.ErrInsufficientAvailableFunds) {
			return nil, apperror.NewBadRequest(err, "Số dư khả dụng không đủ. Bạn chỉ được rút tối đa = Số dư - Số tiền đang đóng băng")
		}
		if errors.Is(err, repository.ErrInvalidWalletState) {
			return nil, apperror.NewInternal(err)
		}
		return nil, apperror.NewInternal(err)
	}
	return wallet, nil
}

func (s *walletService) GetTransactions(userID uint) ([]entity.Transaction, error) {
	transactions, err := s.repo.GetTransactions(userID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return transactions, nil
}

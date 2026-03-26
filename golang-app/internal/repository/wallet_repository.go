package repository

import (
	"errors"
	"golang-app/internal/model/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInvalidWalletAmount        = errors.New("invalid wallet amount")
	ErrInsufficientAvailableFunds = errors.New("insufficient available funds")
	ErrInsufficientFrozenFunds    = errors.New("insufficient frozen funds")
	ErrInvalidWalletState         = errors.New("invalid wallet state")
	ErrWalletTransferSameAccount  = errors.New("wallet transfer requires different accounts")
)

type WalletRepository interface {
	GetByUserID(userID uint) (*entity.Wallet, error)
	Deposit(userID uint, amount float64) (*entity.Wallet, error)
	Withdraw(userID uint, amount float64) (*entity.Wallet, error)
	GetTransactions(userID uint) ([]entity.Transaction, error)
	TransferMoney(fromID, toID uint, amount float64) error
	UnfreezeMoney(userID uint, amount float64) error
}

type walletRepo struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepo{db: db}
}

func validateWalletAmount(amount float64) error {
	if amount <= 0 {
		return ErrInvalidWalletAmount
	}
	return nil
}

func lockWallet(tx *gorm.DB, userID uint) (*entity.Wallet, error) {
	var wallet entity.Wallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

func ensureWalletState(wallet *entity.Wallet) error {
	if wallet.FrozenBalance < 0 || wallet.Balance < 0 || wallet.FrozenBalance > wallet.Balance {
		return ErrInvalidWalletState
	}
	return nil
}

func createWalletTransaction(tx *gorm.DB, userID uint, amount float64, transactionType string, description string) error {
	return tx.Create(&entity.Transaction{
		UserID:      userID,
		Amount:      amount,
		Type:        transactionType,
		Description: description,
	}).Error
}

func (r *walletRepo) GetByUserID(userID uint) (*entity.Wallet, error) {
	var wallet entity.Wallet
	if err := r.db.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

// Deposit nạp tiền vào ví trong 1 Transaction DB
func (r *walletRepo) Deposit(userID uint, amount float64) (*entity.Wallet, error) {
	if err := validateWalletAmount(amount); err != nil {
		return nil, err
	}

	var wallet entity.Wallet
	err := r.db.Transaction(func(tx *gorm.DB) error {
		lockedWallet, err := lockWallet(tx, userID)
		if err != nil {
			return err
		}
		wallet = *lockedWallet
		if err := ensureWalletState(&wallet); err != nil {
			return err
		}

		wallet.Balance += amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		return createWalletTransaction(tx, userID, amount, "deposit", "Nạp tiền vào ví")
	})
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// Withdraw rút tiền từ ví, tối đa balance - frozen_balance
func (r *walletRepo) Withdraw(userID uint, amount float64) (*entity.Wallet, error) {
	if err := validateWalletAmount(amount); err != nil {
		return nil, err
	}

	var wallet entity.Wallet
	err := r.db.Transaction(func(tx *gorm.DB) error {
		lockedWallet, err := lockWallet(tx, userID)
		if err != nil {
			return err
		}
		wallet = *lockedWallet
		if err := ensureWalletState(&wallet); err != nil {
			return err
		}

		available := wallet.Balance - wallet.FrozenBalance
		if amount > available {
			return ErrInsufficientAvailableFunds
		}

		wallet.Balance -= amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		return createWalletTransaction(tx, userID, -amount, "withdraw", "Rút tiền từ ví")
	})
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// GetTransactions lấy lịch sử giao dịch
func (r *walletRepo) GetTransactions(userID uint) ([]entity.Transaction, error) {
	var transactions []entity.Transaction
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *walletRepo) TransferMoney(fromID, toID uint, amount float64) error {
	if err := validateWalletAmount(amount); err != nil {
		return err
	}
	if fromID == toID {
		return ErrWalletTransferSameAccount
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		firstID, secondID := fromID, toID
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
		fromWallet := wallets[fromID]
		toWallet := wallets[toID]

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
		if err := ensureWalletState(fromWallet); err != nil {
			return err
		}
		if err := tx.Save(fromWallet).Error; err != nil {
			return err
		}

		toWallet.Balance += amount
		if err := ensureWalletState(toWallet); err != nil {
			return err
		}
		if err := tx.Save(toWallet).Error; err != nil {
			return err
		}

		if err := createWalletTransaction(tx, fromID, -amount, "payment", "Thanh toán cho phiên đấu giá"); err != nil {
			return err
		}
		if err := createWalletTransaction(tx, toID, amount, "deposit", "Nhận tiền từ phiên đấu giá"); err != nil {
			return err
		}

		return nil
	})
}

func (r *walletRepo) UnfreezeMoney(userID uint, amount float64) error {
	if err := validateWalletAmount(amount); err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		wallet, err := lockWallet(tx, userID)
		if err != nil {
			return err
		}
		if err := ensureWalletState(wallet); err != nil {
			return err
		}
		if wallet.FrozenBalance < amount {
			return ErrInsufficientFrozenFunds
		}

		wallet.FrozenBalance -= amount
		if err := ensureWalletState(wallet); err != nil {
			return err
		}
		if err := tx.Save(wallet).Error; err != nil {
			return err
		}

		return nil
	})
}

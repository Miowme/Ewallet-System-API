package service

import (
	"errors"
	"log"
	"time"
	"e-wallet-system/internal/model"
	"e-wallet-system/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionService interface {
	TopUp(userID string, amount float64) (*model.Transaction, float64, error)
	Transfer(fromUserID, toUserID string, amount float64) (*model.Transaction, error)
	Pay(userID, merchantID string, amount float64, description string) (*model.Transaction, float64, error)
	GetHistory(userID string, page, limit int) ([]model.Transaction, map[string]interface{}, error)
}

type transactionService struct {
	walletRepo      repository.WalletRepository
	transactionRepo repository.TransactionRepository
	db              *gorm.DB
}

func NewTransactionService(walletRepo repository.WalletRepository, transactionRepo repository.TransactionRepository, db *gorm.DB) TransactionService {
	return &transactionService{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		db:              db,
	}
}

func StartExpiryWorker(db *gorm.DB) {
	ticker := time.NewTicker(1 * time.Minute) 
	go func() {
		for range ticker.C {
			expiryTime := time.Now().Add(-10 * time.Minute)
			
			result := db.Model(&model.Transaction{}).
				Where("status = ? AND created_at < ?", "pending", expiryTime).
				Update("status", "expired")

			if result.Error != nil {
				log.Printf("Gagal memperbarui transaksi kedaluwarsa: %v", result.Error)
			} else if result.RowsAffected > 0 {
				log.Printf("Berhasil menandai %d transaksi sebagai expired.", result.RowsAffected)
			}
		}
	}()
}

func (s *transactionService) TopUp(userID string, amount float64) (*model.Transaction, float64, error) {
	if amount <= 0 {
		return nil, 0, errors.New("amount must be greater than zero")
	}

	var transaction *model.Transaction
	var currentBalance float64

	wallet, err := s.walletRepo.FindByUserID(userID)
	if err != nil {
		return nil, 0, errors.New("wallet not found")
	}
	currentBalance = wallet.Balance

	err = s.db.Transaction(func(tx *gorm.DB) error {
		err := s.walletRepo.UpdateBalance(tx, userID, amount)
		if err != nil {
			return err
		}

		parsedUserID, _ := uuid.Parse(userID)
		transaction = &model.Transaction{
			Type:        "topup",
			Amount:      amount,
			ToUserID:    &parsedUserID,
			Status:      "completed",
			Description: stringPtr("Wallet top-up"),
		}

		return s.transactionRepo.Create(tx, transaction)
	})

	if err != nil {
		return nil, 0, err
	}

	newBalance := currentBalance + amount

	return transaction, newBalance, nil
}

func (s *transactionService) Transfer(fromUserID, toUserID string, amount float64) (*model.Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	if fromUserID == toUserID {
		return nil, errors.New("cannot transfer to the same account")
	}

	senderWallet, err := s.walletRepo.FindByUserID(fromUserID)
	if err != nil || senderWallet.Balance < amount {
		return nil, errors.New("insufficient funds")
	}

	_, err = s.walletRepo.FindByUserID(toUserID)
	if err != nil {
		return nil, errors.New("recipient wallet not found")
	}

	var transaction *model.Transaction

	err = s.db.Transaction(func(tx *gorm.DB) error {
		err = s.walletRepo.UpdateBalance(tx, fromUserID, -amount)
		if err != nil {
			return err
		}

		err = s.walletRepo.UpdateBalance(tx, toUserID, amount)
		if err != nil {
			return err
		}

		parsedFrom, _ := uuid.Parse(fromUserID)
		parsedTo, _ := uuid.Parse(toUserID)

		desc := "Fund transfer"
		transaction = &model.Transaction{
			Type:        "transfer_out",
			Amount:      amount,
			FromUserID:  &parsedFrom,
			ToUserID:    &parsedTo,
			Status:      "completed",
			Description: &desc,
		}

		return s.transactionRepo.Create(tx, transaction)
	})

	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (s *transactionService) Pay(userID, merchantID string, amount float64, description string) (*model.Transaction, float64, error) {
	if amount <= 0 {
		return nil, 0, errors.New("amount must be greater than zero")
	}

	var transaction *model.Transaction

	wallet, err := s.walletRepo.FindByUserID(userID)
	if err != nil || wallet.Balance < amount {
		return nil, 0, errors.New("insufficient funds")
	}
	currentBalance := wallet.Balance

	err = s.db.Transaction(func(tx *gorm.DB) error {
		err := s.walletRepo.UpdateBalance(tx, userID, -amount)
		if err != nil {
			return err
		}

		parsedUser, _ := uuid.Parse(userID)
		desc := description
		if desc == "" {
			desc = "Merchant payment"
		}

		transaction = &model.Transaction{
			Type:        "payment",
			Amount:      amount,
			FromUserID:  &parsedUser,
			MerchantID:  &merchantID,
			Status:      "completed",
			Description: &desc,
		}

		return s.transactionRepo.Create(tx, transaction)
	})

	if err != nil {
		return nil, 0, err
	}

	newBalance := currentBalance - amount

	return transaction, newBalance, nil
}

func (s *transactionService) GetHistory(userID string, page, limit int) ([]model.Transaction, map[string]interface{}, error) {
	offset := (page - 1) * limit

	transactions, totalItems, err := s.transactionRepo.GetHistoryByUserID(userID, limit, offset)
	if err != nil {
		return nil, nil, err
	}

	totalPages := int((totalItems + int64(limit) - 1) / int64(limit))
	pagination := map[string]interface{}{
		"currentPage": page,
		"totalPages":  totalPages,
		"totalItems":  totalItems,
	}

	return transactions, pagination, nil
}

func stringPtr(s string) *string {
	return &s
}
package repository

import (
	"e-wallet-system/internal/model"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(tx *gorm.DB, transaction *model.Transaction) error
	GetHistoryByUserID(userID string, limit, offset int) ([]model.Transaction, int64, error)
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(tx *gorm.DB, transaction *model.Transaction) error {
	dbConn := r.db
	if tx != nil {
		dbConn = tx
	}
	return dbConn.Create(transaction).Error
}

func (r *transactionRepository) GetHistoryByUserID(userID string, limit, offset int) ([]model.Transaction, int64, error) {
	var transactions []model.Transaction
	var totalItems int64

	query := r.db.Model(&model.Transaction{}).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID)

	err := query.Count(&totalItems).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	if err != nil {
		return nil, 0, err
	}

	return transactions, totalItems, nil
}
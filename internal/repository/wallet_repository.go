package repository

import (
	"e-wallet-system/internal/model"

	"gorm.io/gorm"
)

type WalletRepository interface {
	Create(wallet *model.Wallet) error
	FindByUserID(userID string) (*model.Wallet, error)
	UpdateBalance(tx *gorm.DB, userID string, amount float64) error
}

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) Create(wallet *model.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *walletRepository) FindByUserID(userID string) (*model.Wallet, error) {
	var wallet model.Wallet
	err := r.db.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) UpdateBalance(tx *gorm.DB, userID string, amount float64) error {
	dbConn := r.db
	if tx != nil {
		dbConn = tx
	}
	return dbConn.Model(&model.Wallet{}).
		Where("user_id = ?", userID).
		Update("balance", gorm.Expr("balance + ?", amount)).Error
}
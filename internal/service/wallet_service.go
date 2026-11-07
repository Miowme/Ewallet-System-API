package service

import (
	"errors"

	"e-wallet-system/internal/model"
	"e-wallet-system/internal/repository"
)

type WalletService interface {
	GetBalance(userID string) (*model.Wallet, error)
}

type walletService struct {
	walletRepo repository.WalletRepository
}

func NewWalletService(walletRepo repository.WalletRepository) WalletService {
	return &walletService{walletRepo: walletRepo}
}

func (s *walletService) GetBalance(userID string) (*model.Wallet, error) {
	wallet, err := s.walletRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("wallet not found")
	}
	return wallet, nil
}
package service

import (
	"errors"
	"os"
	"time"

	"e-wallet-system/internal/model"
	"e-wallet-system/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(name, email, password string) (*model.User, error)
	Login(email, password string) (string, error)
}

type authService struct {
	userRepo   repository.UserRepository
	walletRepo repository.WalletRepository
	db         *gorm.DB
}

func NewAuthService(userRepo repository.UserRepository, walletRepo repository.WalletRepository, db *gorm.DB) AuthService {
	return &authService{
		userRepo:   userRepo,
		walletRepo: walletRepo,
		db:         db,
	}
}

func (s *authService) Register(name, email, password string) (*model.User, error) {
	existingUser, _ := s.userRepo.FindByEmail(email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Create(user); err != nil {
			return err
		}

		wallet := &model.Wallet{
			UserID:  user.ID,
			Balance: 0.00,
		}
		if err := s.walletRepo.Create(wallet); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(email, password string) (string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	secretKey := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), 
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
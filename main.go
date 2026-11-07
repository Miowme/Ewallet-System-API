package main

import (
	"log"
	"os"

	"e-wallet-system/config"
	"e-wallet-system/internal/controller"
	"e-wallet-system/internal/middleware"
	"e-wallet-system/internal/repository"
	"e-wallet-system/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan environment sistem secara langsung")
	}

	config.ConnectDB()

	// Repository
	userRepo := repository.NewUserRepository(config.DB)
	walletRepo := repository.NewWalletRepository(config.DB)
	transactionRepo := repository.NewTransactionRepository(config.DB)

	// Service
	authService := service.NewAuthService(userRepo, walletRepo, config.DB)
	walletService := service.NewWalletService(walletRepo)
	transactionService := service.NewTransactionService(walletRepo, transactionRepo, config.DB)

	// Controller
	authController := controller.NewAuthController(authService)
	walletController := controller.NewWalletController(walletService)
	transactionController := controller.NewTransactionController(transactionService)

	// Setup Router Gin
	r := gin.Default()

	// Base API Group sesuai spesifikasi: /api/v1
	api := r.Group("/api/v1")
	{
		// Public Routes (User Auth)
		api.POST("/users/register", authController.Register)
		api.POST("/users/login", authController.Login)

		// Protected Routes (Memerlukan JWT Bearer Token)
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// Wallet
			protected.GET("/wallet/balance", walletController.GetBalance)

			// Transactions
			protected.POST("/transactions/topup", transactionController.TopUp)
			protected.POST("/transactions/transfer", transactionController.Transfer)
			protected.POST("/transactions/pay", transactionController.Pay)
			protected.GET("/transactions/history", transactionController.GetHistory)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server e-wallet berjalan pada port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
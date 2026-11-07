package controller

import (
	"net/http"

	"e-wallet-system/internal/service"

	"github.com/gin-gonic/gin"
)

type WalletController struct {
	walletService service.WalletService
}

func NewWalletController(walletService service.WalletService) *WalletController {
	return &WalletController{walletService: walletService}
}

func (c *WalletController) GetBalance(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	wallet, err := c.walletService.GetBalance(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"userID":  wallet.UserID.String(),
		"balance": wallet.Balance,
	})
}
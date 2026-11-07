package controller

import (
	"net/http"
	"strconv"

	"e-wallet-system/internal/service"

	"github.com/gin-gonic/gin"
)

type TransactionController struct {
	transactionService service.TransactionService
}

func NewTransactionController(transactionService service.TransactionService) *TransactionController {
	return &TransactionController{transactionService: transactionService}
}

func (c *TransactionController) TopUp(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trx, newBalance, err := c.transactionService.TopUp(userID, req.Amount)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"transactionID": trx.ID.String(),
		"message":       "Top-up successful",
		"newBalance":    newBalance,
	})
}

func (c *TransactionController) Transfer(ctx *gin.Context) {
	fromUserID := ctx.MustGet("userID").(string)

	var req struct {
		ToUserID string  `json:"toUserID" binding:"required,uuid"`
		Amount   float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trx, err := c.transactionService.Transfer(fromUserID, req.ToUserID, req.Amount)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"transactionID": trx.ID.String(),
		"status":        trx.Status,
		"message":       "Transfer initiated successfully.",
	})
}

func (c *TransactionController) Pay(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	var req struct {
		MerchantID  string  `json:"merchantID" binding:"required"`
		Amount      float64 `json:"amount" binding:"required,gt=0"`
		Description string  `json:"description"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trx, newBalance, err := c.transactionService.Pay(userID, req.MerchantID, req.Amount, req.Description)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"transactionID": trx.ID.String(),
		"message":       "Payment successful",
		"newBalance":    newBalance,
	})
}

func (c *TransactionController) GetHistory(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	transactions, pagination, err := c.transactionService.GetHistory(userID, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format output sesuai spesifikasi
	var formattedTrxs []map[string]interface{}
	for _, t := range transactions {
		fromStr := ""
		if t.FromUserID != nil {
			fromStr = t.FromUserID.String()
		}
		toStr := ""
		if t.ToUserID != nil {
			toStr = t.ToUserID.String()
		}
		if t.MerchantID != nil && *t.MerchantID != "" {
			toStr = *t.MerchantID
		}

		formattedTrxs = append(formattedTrxs, map[string]interface{}{
			"transactionID": t.ID.String(),
			"type":          t.Type,
			"amount":        t.Amount,
			"from":          fromStr,
			"to":            toStr,
			"timestamp":     t.CreatedAt,
			"status":        t.Status,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"transactions": formattedTrxs,
		"pagination":   pagination,
	})
}
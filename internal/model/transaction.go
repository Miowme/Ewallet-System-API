package model

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ReferenceID *string    `gorm:"type:varchar(100);unique"`
	Type        string     `gorm:"type:varchar(30);not null"` 
	Amount      float64    `gorm:"type:numeric(18,2);not null"`
	FromUserID  *uuid.UUID `gorm:"type:uuid"`
	ToUserID    *uuid.UUID `gorm:"type:uuid"`
	MerchantID  *string    `gorm:"type:varchar(100)"`
	Status      string     `gorm:"type:varchar(20);not null;default:'completed'"` 
	Description *string    `gorm:"type:text"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
}

func (Transaction) TableName() string {
	return "transactions"
}
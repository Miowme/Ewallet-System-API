package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"` 
	Name         string    `gorm:"type:varchar(100);not null"`
	Email        string    `gorm:"type:varchar(150);unique;not null"`
	PasswordHash string    `gorm:"type:varchar(255);not null;column:password_hash"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	Wallet       *Wallet   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

func (User) TableName() string {
	return "users"
}
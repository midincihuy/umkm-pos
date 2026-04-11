package model

import (
	"time"

	"github.com/google/uuid"
)

type AccountType string

const (
	AccountBank       AccountType = "bank"
	AccountCash       AccountType = "cash"
	AccountEwallet    AccountType = "ewallet"
	AccountInvestment AccountType = "investment"
)

type Account struct {
	ID             uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         uuid.UUID   `gorm:"type:uuid;not null;index"`
	Name           string      `gorm:"type:varchar(100);not null"`
	Type           AccountType `gorm:"type:account_type;not null;default:'bank'"`
	OpeningBalance float64     `gorm:"type:numeric(15,2);not null;default:0"`
	CurrentBalance float64     `gorm:"type:numeric(15,2);not null;default:0"`
	Icon           string      `gorm:"type:varchar(50)"`
	Color          string      `gorm:"type:varchar(10)"`
	IsActive       bool        `gorm:"not null;default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	User User `gorm:"foreignKey:UserID"`
}

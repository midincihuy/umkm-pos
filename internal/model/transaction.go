package model

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TxIncome         TransactionType = "income"
	TxExpense        TransactionType = "expense"
	TxTransfer       TransactionType = "transfer"
	TxDebtPayment    TransactionType = "debt_payment"
	TxOpeningBalance TransactionType = "opening_balance"
	TxAdjustment     TransactionType = "adjustment"
)

type Transaction struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         uuid.UUID       `gorm:"type:uuid;not null;index"`
	AccountID      uuid.UUID       `gorm:"type:uuid;not null;index"`
	CategoryID     *uuid.UUID      `gorm:"type:uuid;index"`
	DebtID         *uuid.UUID      `gorm:"type:uuid"`
	TransferPairID *uuid.UUID      `gorm:"type:uuid"`
	Type           TransactionType `gorm:"type:transaction_type;not null"`
	Amount         float64         `gorm:"type:numeric(15,2);not null"`
	BalanceImpact  float64         `gorm:"type:numeric(15,2);not null"`
	Date           time.Time       `gorm:"type:date;not null;index"`
	Description    string          `gorm:"type:varchar(255)"`
	Notes          string          `gorm:"type:text"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	User     User      `gorm:"foreignKey:UserID"`
	Account  Account   `gorm:"foreignKey:AccountID"`
	Category *Category `gorm:"foreignKey:CategoryID"`
}

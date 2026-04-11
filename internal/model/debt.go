package model

import (
	"time"

	"github.com/google/uuid"
)

type DebtType   string
type DebtStatus string

const (
	DebtTypeDebt        DebtType   = "debt"
	DebtTypeReceivable  DebtType   = "receivable"
	DebtStatusActive    DebtStatus = "active"
	DebtStatusPaid      DebtStatus = "paid"
	DebtStatusCancelled DebtStatus = "cancelled"
)

type Debt struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	Type        DebtType   `gorm:"type:debt_type;not null"`
	ContactName string     `gorm:"type:varchar(100);not null"`
	TotalAmount float64    `gorm:"type:numeric(15,2);not null"`
	PaidAmount  float64    `gorm:"type:numeric(15,2);not null;default:0"`
	StartDate   time.Time  `gorm:"type:date;not null"`
	DueDate     *time.Time `gorm:"type:date"`
	Status      DebtStatus `gorm:"type:debt_status;not null;default:'active'"`
	Notes       string     `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Payments []DebtPayment `gorm:"foreignKey:DebtID"`
}

func (d *Debt) RemainingAmount() float64 {
	return d.TotalAmount - d.PaidAmount
}

func (d *Debt) ProgressPct() float64 {
	if d.TotalAmount == 0 {
		return 0
	}
	return (d.PaidAmount / d.TotalAmount) * 100
}

type DebtPayment struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DebtID        uuid.UUID `gorm:"type:uuid;not null;index"`
	AccountID     uuid.UUID `gorm:"type:uuid;not null"`
	TransactionID uuid.UUID `gorm:"type:uuid"`
	Amount        float64   `gorm:"type:numeric(15,2);not null"`
	Date          time.Time `gorm:"type:date;not null"`
	Notes         string    `gorm:"type:text"`
	CreatedAt     time.Time

	Account Account `gorm:"foreignKey:AccountID"`
}

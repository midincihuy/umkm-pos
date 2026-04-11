package model

import (
	"time"

	"github.com/google/uuid"
)

type CashAdjustment struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index"`
	AccountID     uuid.UUID `gorm:"type:uuid;not null"`
	BalanceBefore float64   `gorm:"type:numeric(15,2);not null"`
	ActualBalance float64   `gorm:"type:numeric(15,2);not null"`
	Difference    float64   `gorm:"type:numeric(15,2);not null"`
	Date          time.Time `gorm:"type:date;not null"`
	Notes         string    `gorm:"type:text"`
	CreatedAt     time.Time

	Account Account `gorm:"foreignKey:AccountID"`
}
